# PostgreSQL (Codesphere Managed Service) Reference

> **Last updated:** 2026-07-28 · **Source:** https://docs.codesphere.com/managed-services/providers/postgresql

> The documentation changes regularly. If you are facing issues, please refer to the official documentation:
> <https://docs.codesphere.com/managed-services/providers/postgresql>

## Overview

General-purpose relational database for transactional workloads, application state, and structured analytics. **GA** (not preview). Each service is a dedicated PostgreSQL server on Codesphere's Kubernetes infrastructure, powered by the CloudNativePG operator, running Codesphere's own PostgreSQL images with a large extension set preinstalled (PostGIS, pgvector, ...). Use this reference when generating `ci.yml` Landscape definitions, API payloads, or connection code for the `postgres` provider.

| Property                 | Value                                                  |
| ------------------------ | ------------------------------------------------------ |
| Provider name            | `postgres`                                             |
| Category                 | `Database`                                             |
| Scope                    | `global` (team-scoped, not tied to one workspace)      |
| Team singleton           | `false` — multiple instances allowed                   |
| Pause support            | `true`                                                 |
| Backups / PITR           | `true` / `true`                                        |
| High availability        | Not available — single instance only                   |
| Automatic upgrades       | Never — manual minor version bump via `version` config |
| Automatic storage growth | Never — bump `storage` plan parameter manually         |

## Core Concepts

- **Dedicated instance**: own pod, own persistent volume (Ceph RBD, replicated at the storage layer). Nothing shared with other services/teams.
- **CloudNativePG**: the Kubernetes operator managing lifecycle, config changes, minor version updates, volume resizing, recovery.
- **Single instance / no replicas**: config updates and version upgrades can cause brief downtime.
- **Barman Cloud plugin**: implements backups — periodic physical base backups + continuous WAL archiving to S3-compatible storage, enabling point-in-time recovery (PITR).
- **Deterministic hostnames**: `ms-{providerName}-{providerVersion}-{teamId}-landscape-{workspaceId}-{serviceName}.ms-postgres`, lowercased, invalid characters replaced with hyphens.

## API / Syntax

### Config Schema

| Name           | Type   | Required | Description                                                                                                                                    |
| -------------- | ------ | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `version`      | string | No       | PostgreSQL engine version. Default `17.9`. Allowed: `17.9`, `17.6`, `16.13`, `16.10`, `15.17`, `15.14`, `14.22`, `14.19`. Minor upgrades only. |
| `userName`     | string | No       | Default `app`. Immutable after creation. Cannot be `postgres`.                                                                                 |
| `databaseName` | string | No       | Default `app`. Immutable after creation.                                                                                                       |

### Secrets Schema

| Name                | Type   | Required | Description                                     |
| ------------------- | ------ | -------- | ----------------------------------------------- |
| `userPassword`      | string | Yes      | Password for the application user (`userName`). |
| `superuserPassword` | string | Yes      | Password for the `postgres` superuser.          |

### Details / Output Schema

| Name       | Type    | Availability       | Description                               |
| ---------- | ------- | ------------------ | ----------------------------------------- |
| `hostname` | string  | after provisioning | Internal service hostname.                |
| `port`     | integer | after provisioning | PostgreSQL port (`5432`).                 |
| `dsn`      | string  | after provisioning | Full connection string.                   |
| `ready`    | boolean | after provisioning | Whether the instance accepts connections. |

### Plan: `Small` (`id: 0`)

| Name      | Type    | Default | Min   | Max | Static | Description              |
| --------- | ------- | ------- | ----- | --- | ------ | ------------------------ |
| `cpu`     | number  | `1`     | -     | -   | Yes    | Priced as `cpu-tenths`.  |
| `memory`  | integer | `128`   | -     | -   | Yes    | Priced as `ram-mib`.     |
| `storage` | integer | `1024`  | `512` | -   | No     | Priced as `storage-mib`. |

### Landscape Example

- **Description:** Deploy a PostgreSQL service inside a `ci.yml` Landscape and reference its credentials via the vault.
- **Example:**

```yaml
schemaVersion: v0.4
run:
  app-db:
    provider:
      name: postgres
      schemaVersion: v1
    plan:
      id: 0
      parameters:
        storage: 2048
    config:
      version: "17.9"
      userName: "${{ workspace.env.PGUSER }}"
      databaseName: "${{ workspace.env.PGDATABASE }}"
    secrets:
      userPassword: "${{ vault.pgUserPassword }}"
      superuserPassword: "${{ vault.pgSuperuserPassword }}"
```

### Connecting — `psql`

- **Example:**

```bash
# Install psql
nix-env -iA nixpkgs.postgresql
# General syntax
psql "postgres://<username>@<hostname>:5432/<database>" -W
# Example
psql "postgres://admin@10.0.0.5:5432/mydb" -W
```

### Connecting — Node.js (`pg`)

- **Example:**

```javascript
const { Client } = require("pg");
const client = new Client({
  connectionString: "postgres://admin:secure-password@10.0.0.5:5432/mydb",
});
await client.connect();
const res = await client.query("SELECT $1::text as message", ["Hello Codesphere!"]);
console.log(res.rows[0].message); // Hello Codesphere!
await client.end();
```

### `PATCH /managed-services/{id}` — Enable Backups

- **Parameters:**

| Name                                                     | Type    | Required | Description                                     |
| -------------------------------------------------------- | ------- | -------- | ----------------------------------------------- |
| `backups.enabled`                                        | boolean | Yes      | Turn backups on.                                |
| `backups.intervalH`                                      | integer | Yes      | Backup interval in hours.                       |
| `backups.deleteRetentionDays`                            | integer | Yes      | Retention window in days.                       |
| `backups.config.endpointUrl`                             | string  | Yes      | **Regional** S3 endpoint (not bucket-specific). |
| `backups.config.destinationPath`                         | string  | Yes      | `s3://bucket/` target path.                     |
| `backups.config.accessKey` / `backups.secrets.secretKey` | string  | Yes      | S3 credentials.                                 |

- **Example:**

```bash
curl -X PATCH "https://api.codesphere.com/managed-services/YOUR_SERVICE_ID" \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "backups": {
      "enabled": true,
      "intervalH": 12,
      "deleteRetentionDays": 30,
      "config": {
        "endpointUrl": "https://s3.eu-central-1.amazonaws.com",
        "destinationPath": "s3://my-codesphere-backups/",
        "accessKey": "YOUR_S3_ACCESS_KEY"
      },
      "secrets": { "secretKey": "YOUR_S3_SECRET_KEY" }
    }
  }'
```

### Required S3 IAM Permissions (for backup store)

- **Example:**

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "BucketLevelOperations",
      "Effect": "Allow",
      "Action": ["s3:GetBucketLocation", "s3:ListBucket", "s3:ListBucketMultipartUploads"],
      "Resource": "arn:aws:s3:::YOUR_BUCKET_NAME"
    },
    {
      "Sid": "ObjectLevelOperations",
      "Effect": "Allow",
      "Action": [
        "s3:AbortMultipartUpload",
        "s3:DeleteObject",
        "s3:GetObject",
        "s3:ListMultipartUploadParts",
        "s3:PutObject",
        "s3:PutObjectTagging"
      ],
      "Resource": "arn:aws:s3:::YOUR_BUCKET_NAME/*"
    }
  ]
}
```

### Recovery — Point-in-Time

- **Description:** Restoring always creates a **new** managed service; the source service is untouched.
- **Parameters:**

| Name                                         | Type               | Required | Description                                        |
| -------------------------------------------- | ------------------ | -------- | -------------------------------------------------- |
| `recoverFrom.msId`                           | string             | Yes      | Source managed service ID.                         |
| `recoverFrom.time`                           | string (date-time) | Yes      | Target recovery timestamp (ISO 8601).              |
| `recoverFrom.config` / `recoverFrom.secrets` | object             | Yes      | Same S3 endpoint/credentials as the backup config. |

- **Example:**

```bash
curl -X POST "https://api.codesphere.com/managed-services" \
  -H "Authorization: Bearer YOUR_API_TOKEN" -H "Content-Type: application/json" \
  -d '{
    "teamId": 123,
    "name": "my-postgres-recovered",
    "provider": { "name": "postgres", "version": "v1" },
    "plan": { "id": 0, "parameters": { "storage": 2048 } },
    "config": { "version": "17.9" },
    "secrets": { "userPassword": "secure-password", "superuserPassword": "secure-superuser-password" },
    "recoverFrom": {
      "msId": "OLD_MANAGED_SERVICE_ID",
      "time": "2026-04-10T12:00:00Z",
      "config": {
        "endpointUrl": "https://s3.eu-central-1.amazonaws.com",
        "destinationPath": "s3://my-codesphere-backups/",
        "accessKey": "YOUR_S3_ACCESS_KEY"
      },
      "secrets": { "secretKey": "YOUR_S3_SECRET_KEY" }
    }
  }'
```

### Recovery — Specific Backup

- **Parameters:** same as above except `recoverFrom.id` (backup UUID) replaces `msId` + `time`.
- **Example:** identical payload shape with `"recoverFrom": { "id": "BACKUP_UUID_HERE", "config": {...}, "secrets": {...} }`.

### Extensions

- **Description:** All listed extensions ship on every instance but are inactive until enabled per database.
- **Enabling:**

```sql
-- Enable an extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
-- Verify which extensions are installed
SELECT extname, extversion FROM pg_extension;
-- Disable an extension again
DROP EXTENSION "uuid-ossp";
-- List all available extensions on the instance
SELECT name, default_version, installed_version FROM pg_available_extensions ORDER BY name;
```

- **Notable extensions:** `postgis` / `postgis_raster` / `postgis_sfcgal` / `postgis_topology` (3.6.0, spatial), `vector` (0.8.1, pgvector similarity search), `pgcrypto`, `pg_stat_statements`, `pg_trgm`, `hstore`, `uuid-ossp`, `citext`, `postgres_fdw`, `dblink`, `pgaudit`. Full list (50+ extensions) is in the source docs — query `pg_available_extensions` on the live instance for the authoritative, versioned list.

## Common Pitfalls

- Expecting HA/replicas — the service is always a single instance; version/config changes can cause brief downtime.
- Expecting storage to grow automatically — must bump the `storage` plan parameter manually.
- Trying to rename `userName`/`databaseName` after creation — both are immutable.
- Setting `userName` to `postgres` — rejected, that's the reserved superuser name.
- Recovering into a backup without at least one prior transaction, or targeting a time before the earliest backup / after the latest archived WAL — recovery will fail.
- Trying to change users/database name/passwords during recovery — must match what was in the backup.
- Using a bucket-specific S3 hostname for `endpointUrl` instead of the **regional** endpoint.

## Known Documentation Discrepancies

> This section may be outdated. Please verify against the official documentation:
> <https://docs.codesphere.com/managed-services/providers/postgresql>

- The full extension table (50+ entries) is reproduced from a point-in-time scrape; exact versions (e.g. `postgis` 3.6.0, `vector` 0.8.1) will drift as Codesphere updates its base images — treat `pg_available_extensions` on the live instance as authoritative over this file.
- Allowed `version` values (`17.9` down to `14.19`) will change as new minors are released and old ones deprecated.

## Further Reading

- Official docs: <https://docs.codesphere.com/managed-services/providers/postgresql>
- Backups (general concepts): <https://docs.codesphere.com/managed-services/backups>
- CloudNativePG: <https://cloudnative-pg.io/>
- CloudNativePG Barman Cloud plugin: <https://cloudnative-pg.io/plugin-barman-cloud/>
- `pg` Node.js driver: <https://node-postgres.com/>
