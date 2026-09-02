# Object Storage (Codesphere Managed Service) Reference

> **Last updated:** 2026-07-28 · **Source:** https://docs.codesphere.com/managed-services/providers/object-storage

> The documentation changes regularly. If you are facing issues, please refer to the official documentation:
> <https://docs.codesphere.com/managed-services/providers/object-storage>

## Overview

An S3-compatible API (Ceph/RGW-backed) for files, backups, assets, and other unstructured data — bucket-based object storage instead of a filesystem or relational database. **Preview** — not enabled by default, must be turned on by the operator; schema/plans/behavior may still change.

| Property          | Value                                                               |
| ----------------- | ------------------------------------------------------------------- |
| Provider name     | `s3`                                                                |
| Category          | `Storage`                                                           |
| Scope             | `global`                                                            |
| Team singleton    | `false`                                                             |
| Pause support     | `false`                                                             |
| Backups           | **None yet** — you are responsible for backing up your own data     |
| High availability | Yes — Ceph storage is redundant and replicated at the storage layer |

## Core Concepts

- **No backups (yet)**: unlike PostgreSQL/Babelfish/DocumentDB, this provider has no automated backup capability. Treat stored data as your own responsibility to back up.
- **Fixed managed endpoint**: for Codesphere-managed S3, the endpoint URL is always `http://rgw-load-balancer.rook-ceph.svc.cluster.local` — it does not vary per bucket or user.
- **Quota-based plan**: the single `Generic` plan is entirely about quota limits (buckets, objects, size, throughput, ops/s), not compute sizing.

## API / Syntax

### Config Schema

| Name                | Type   | Required | Description                                                                                                     |
| ------------------- | ------ | -------- | --------------------------------------------------------------------------------------------------------------- |
| `accessKey`         | string | Yes      | Cluster-unique access key. Must be exactly 20 uppercase letters or digits.                                      |
| `userDisplayName`   | string | No       | Default `My S3 User`. Friendly label for the generated user.                                                    |
| `initialBucketName` | string | Yes      | Cluster-unique initial bucket name. If already taken, the bucket is **not** created (service still provisions). |

### Secrets Schema

| Name        | Type   | Required | Description                                                    |
| ----------- | ------ | -------- | -------------------------------------------------------------- |
| `secretKey` | string | Yes      | Secret access key. Must be exactly 40 alphanumeric characters. |

### Details / Output Schema

| Name     | Type   | Availability       | Description                                                                                                          |
| -------- | ------ | ------------------ | -------------------------------------------------------------------------------------------------------------------- |
| `url`    | string | after provisioning | S3-compatible endpoint URL. Always `http://rgw-load-balancer.rook-ceph.svc.cluster.local` for Codesphere-managed S3. |
| `userId` | string | after provisioning | Internal identifier of the generated object-storage user.                                                            |

### Plan: `Generic` (`id: 0`)

- **Description:** All quota parameters are adjustable.

| Name                | Type    | Default     | Min | Max           | Static | Description                       |
| ------------------- | ------- | ----------- | --- | ------------- | ------ | --------------------------------- |
| `maxBuckets`        | integer | `50`        | `1` | `1000`        | No     | Maximum number of buckets.        |
| `maxObjects`        | integer | `100000`    | `1` | `10000000`    | No     | Maximum number of objects.        |
| `maxSizeKb`         | integer | `10000000`  | `1` | `10000000000` | No     | Total size limit (KB).            |
| `maxReadOpsPerS`    | integer | `1000`      | `1` | `10000`       | No     | Max read ops/sec.                 |
| `maxWriteOpsPerS`   | integer | `1000`      | `1` | `10000`       | No     | Max write ops/sec.                |
| `maxReadBytesPerS`  | integer | `100000000` | `1` | `10000000000` | No     | Max read throughput (bytes/sec).  |
| `maxWriteBytesPerS` | integer | `100000000` | `1` | `10000000000` | No     | Max write throughput (bytes/sec). |

### Landscape Example

- **Example:**

```yaml
schemaVersion: v0.4
run:
  uploads:
    provider:
      name: s3
      schemaVersion: v1
    plan:
      id: 0
      parameters:
        maxBuckets: 50
        maxObjects: 100000
        maxSizeKb: 10000000
        maxReadOpsPerS: 1000
        maxWriteOpsPerS: 1000
        maxReadBytesPerS: 100000000
        maxWriteBytesPerS: 100000000
    config:
      accessKey: "${{ workspace.env.S3_ACCESS_KEY }}"
      userDisplayName: "Landscape Upload User"
      initialBucketName: "${{ workspace.env.S3_BUCKET }}"
    secrets:
      secretKey: "${{ vault.s3SecretKey }}"
```

### Connecting — Terminal (`mc`, MinIO Client)

- **Example:**

```bash
# Install mc
nix-env -iA nixpkgs.minio-client
# Configure alias
mc alias set my-storage http://rgw-load-balancer.rook-ceph.svc.cluster.local "$ACCESS_KEY" "$SECRET_KEY"
# List buckets
mc ls my-storage
# Copy file
mc cp myfile.txt my-storage/my-bucket/
```

### Connecting — Node.js (AWS SDK v3)

- **Example:**

```javascript
const { S3 } = require("@aws-sdk/client-s3");
const s3 = new S3({
  endpoint: "http://rgw-load-balancer.rook-ceph.svc.cluster.local",
  region: "us-east-1",
  credentials: {
    accessKeyId: "YOUR_ACCESS_KEY",
    secretAccessKey: "YOUR_SECRET_KEY",
  },
  forcePathStyle: true,
  tls: false,
});
const { Buckets } = await s3.listBuckets({});
console.log(Buckets);
```

## Common Pitfalls

- Assuming automated backups exist for stored objects — they don't (yet); back up important data yourself.
- Using an `accessKey` that isn't exactly 20 uppercase letters/digits, or a `secretKey` that isn't exactly 40 alphanumeric characters — both are validated and will reject the request.
- Picking an `initialBucketName` that's already taken cluster-wide and assuming the service creation fails — it doesn't; the _bucket_ silently isn't created while the service still provisions.
- Forgetting `forcePathStyle: true` / omitting `tls: false` when pointing an S3 SDK at the internal Codesphere endpoint.
- Trying to pause an Object Storage service — not supported for this provider.

## Known Documentation Discrepancies

> This section may be outdated. Please verify against the official documentation:
> <https://docs.codesphere.com/managed-services/providers/object-storage>

- Backups are explicitly listed as "not yet available" at time of writing — check `capabilities.backups` via `GET /api/managed-services/providers` in case this has since shipped.
- As a preview feature, quota plan defaults/ranges (`maxBuckets`, `maxObjects`, etc.) may change.

## Further Reading

- Official docs: <https://docs.codesphere.com/managed-services/providers/object-storage>
- MinIO Client (`mc`): <https://min.io/docs/minio/linux/reference/minio-mc.html>
- AWS SDK for JavaScript v3 (`@aws-sdk/client-s3`): <https://www.npmjs.com/package/@aws-sdk/client-s3>
