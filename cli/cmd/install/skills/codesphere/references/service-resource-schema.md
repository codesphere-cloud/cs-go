# Managed Service Resource Schema Reference

> **Last updated:** 2026-07-28 · **Source:** Codesphere public API — `service-provider` resource schema (JSON Schema export)

> The documentation changes regularly. If you are facing issues, please refer to the official documentation:
> <https://docs.codesphere.com/managed-services/overview>

## Overview

The exact shape of a **Managed Service resource** as returned by `GET /managed-services` (array) and `GET /managed-services/{id}` (single object), derived directly from the platform's JSON Schema for this resource. Use this reference when parsing API responses — e.g. to check `status.state`, inspect `backups.entries`, or read `recentEvents` — as distinct from [README.md](./README.md) (architecture/lifecycle) and the per-provider files (provider-specific `config`/`secrets`/`details` schemas).

## Core Concepts

- **`provider` requires all three of `name`, `version`, AND `schemaVersion` together** on the returned resource — they are not alternates of each other on the API response, even though `ci.yml` input syntax only uses one of `provider.version` (schemaVersion v0.2/v0.3) or `provider.schemaVersion` (v0.4) at a time. See "Known Documentation Discrepancies" below.
- **`status` is a discriminated union keyed by `state`** — each state has its own required fields (e.g. `deleted`/`deleting` require `deletedAt`; `synchronized` requires `detailsRef`).
- **`backups` is a two-shape union keyed by `enabled`** — the `true` shape requires `deleteRetentionDays`, `intervalH`, `entries`, and `config`; the `false` shape only requires `entries`.
- **`recentEvents`** follow an OpenTelemetry-log-record-like shape (`eventName`, `timestamp`, `observedTimestamp`, `severityNumber`, `body`, `resource`, `attributes`).
- **Top-level `version`** on the resource is a semver string for the _managed service resource itself_ — distinct from `provider.version`/`provider.schemaVersion`, which version the _provider definition_.

## API / Syntax

### Managed Service Resource — Top-Level Fields

| Name           | Type               | Required | Description                                                                                                                                                            |
| -------------- | ------------------ | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `id`           | string (uuid)      | Yes      | Resource ID, e.g. `8316ee2f-87f8-424e-b925-7382dc50d662`.                                                                                                              |
| `provider`     | object             | Yes      | See "Provider Sub-Object" below.                                                                                                                                       |
| `name`         | string             | Yes      | Alphanumeric with spaces allowed in between, max length 127. Pattern: `^(?!\s)[a-zA-Z0-9\-_\s]{0,126}[a-zA-Z0-9\-_]$` (no leading whitespace, no trailing whitespace). |
| `backups`      | object             | Yes      | See "Backups Sub-Object" below.                                                                                                                                        |
| `pause`        | boolean            | Yes      | Current pause flag.                                                                                                                                                    |
| `plan`         | object             | Yes      | See "Plan Sub-Object" below.                                                                                                                                           |
| `createdAt`    | string (date-time) | Yes      | Creation timestamp.                                                                                                                                                    |
| `config`       | object             | Yes      | Provider-specific config; `additionalProperties` — any defined value, schema is provider-specific.                                                                     |
| `recentEvents` | array              | No       | See "Recent Events" below.                                                                                                                                             |
| `status`       | object             | Yes      | Discriminated union by `state` — see "Status Sub-Object" below.                                                                                                        |
| `version`      | string (semver)    | No       | Version of the managed service resource itself, e.g. `1.0.0`.                                                                                                          |

### Provider Sub-Object

| Name            | Type   | Required | Description                                                                                                                                                                              |
| --------------- | ------ | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `name`          | string | Yes      | Pattern `^[a-z0-9-_]+$` — matches the provider name used in `ci.yml`/API (e.g. `postgres`, `valkey`, `virtual-k8s`).                                                                     |
| `version`       | string | Yes      | Pattern `^v[0-9][0-9a-z]*$` (e.g. `v1`, `v0`).                                                                                                                                           |
| `schemaVersion` | string | Yes      | Same pattern as `version`. **Both fields are present and required together on the API resource**, regardless of which `ci.yml` schemaVersion (v0.2/v0.3 vs. v0.4) was used to create it. |

### Backups Sub-Object — `enabled: true` shape

| Name                  | Type    | Required | Description                                                                                                                                                       |
| --------------------- | ------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `enabled`             | boolean | Yes      | Must be `true` for this shape.                                                                                                                                    |
| `deleteRetentionDays` | integer | Yes      | Range: `1`–`365`.                                                                                                                                                 |
| `intervalH`           | integer | Yes      | Range: `1`–`730`.                                                                                                                                                 |
| `entries`             | array   | Yes      | See "Backup Entry" below.                                                                                                                                         |
| `config`              | object  | Yes      | Backup target config (`additionalProperties`, any defined value — e.g. `endpointUrl`, `destinationPath` for S3-compatible targets, per the provider backup docs). |

### Backups Sub-Object — `enabled: false` shape

| Name      | Type    | Required | Description                                                                                   |
| --------- | ------- | -------- | --------------------------------------------------------------------------------------------- |
| `enabled` | boolean | Yes      | Must be `false` for this shape.                                                               |
| `entries` | array   | Yes      | Historical backup entries persist even after backups are disabled — see "Backup Entry" below. |

### Backup Entry

| Name          | Type               | Required | Description                             |
| ------------- | ------------------ | -------- | --------------------------------------- |
| `id`          | string (uuid)      | Yes      | Backup entry ID.                        |
| `scheduledAt` | string (date-time) | Yes      | When the backup was scheduled.          |
| `initiatedAt` | string (date-time) | No       | When the backup actually started.       |
| `confirmedAt` | string (date-time) | No       | When the backup was confirmed complete. |

### Plan Sub-Object

| Name         | Type    | Required | Description                                                                                                                              |
| ------------ | ------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| `id`         | integer | Yes      | Minimum `0`. Plan tier id, e.g. `0` for `Small`.                                                                                         |
| `parameters` | object  | Yes      | `additionalProperties` of type `integer` — e.g. `storage`, `cpu`, `memory` (plan-specific, matches the provider's plan parameter table). |

### Status Sub-Object — State Machine

`status` is a discriminated union on `state`. Each state has its own required shape:

| `state`            | Additional required fields | Description                                                                                                                       |
| ------------------ | -------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `creating`         | —                          | Initial provisioning in progress.                                                                                                 |
| `updating`         | —                          | An update (config/plan/secrets/backups/pause) is being applied.                                                                   |
| `pausing`          | —                          | Transitioning to paused.                                                                                                          |
| `unpausing`        | —                          | Transitioning out of paused.                                                                                                      |
| `paused`           | —                          | Compute released; for Codesphere-managed services the persistent volume is retained.                                              |
| `synchronized`     | `detailsRef` (string)      | Healthy and ready; `detailsRef` points to the provider-specific `details` payload (e.g. `hostname`/`port`/`dsn`).                 |
| `deleting`         | `deletedAt` (date-time)    | Soft-deleted, in the retention/teardown window.                                                                                   |
| `deleted`          | `deletedAt` (date-time)    | Hard-deleted — permanent, compute and volume removed.                                                                             |
| `unknown`          | —                          | State could not be determined.                                                                                                    |
| `invalid provider` | —                          | The referenced provider name/version no longer resolves (e.g. a custom provider was deleted, or a preview provider was disabled). |

- **Example — polling for readiness:**

```bash
curl -H "Authorization: Bearer $CS_TOKEN" "https://<instance>/api/managed-services/<id>" \
  | jq '.status.state'
# Poll until this returns "synchronized"
```

### Recent Events — Event Object

| Name                | Type               | Required | Description                                                         |
| ------------------- | ------------------ | -------- | ------------------------------------------------------------------- |
| `eventName`         | string             | Yes      | Event type/name.                                                    |
| `timestamp`         | string (date-time) | Yes      | Event timestamp.                                                    |
| `observedTimestamp` | string (date-time) | Yes      | When the event was observed/ingested.                               |
| `severityNumber`    | integer            | Yes      | Minimum `0` — OpenTelemetry-style severity level.                   |
| `body`              | string             | Yes      | Human-readable event message.                                       |
| `resource`          | string             | Yes      | The resource the event pertains to.                                 |
| `attributes`        | object             | Yes      | `additionalProperties` of type `string` — key-value event metadata. |

- **Example response fragment:**

```json
{
  "id": "8316ee2f-87f8-424e-b925-7382dc50d662",
  "provider": { "name": "postgres", "version": "v1", "schemaVersion": "v1" },
  "name": "app-db",
  "backups": {
    "enabled": true,
    "deleteRetentionDays": 30,
    "intervalH": 12,
    "entries": [
      {
        "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
        "scheduledAt": "2026-07-28T02:00:00Z",
        "initiatedAt": "2026-07-28T02:00:05Z",
        "confirmedAt": "2026-07-28T02:04:12Z"
      }
    ],
    "config": {
      "endpointUrl": "https://s3.eu-central-1.amazonaws.com",
      "destinationPath": "s3://my-codesphere-backups/"
    }
  },
  "pause": false,
  "plan": { "id": 0, "parameters": { "storage": 2048, "cpu": 1, "memory": 128 } },
  "createdAt": "2026-06-01T09:15:00Z",
  "config": { "version": "17.9", "userName": "app", "databaseName": "app" },
  "recentEvents": [],
  "status": {
    "state": "synchronized",
    "detailsRef": "https://<instance>/api/managed-services/8316ee2f-87f8-424e-b925-7382dc50d662/details"
  },
  "version": "1.0.0"
}
```

## Common Pitfalls

- Assuming `provider.version` and `provider.schemaVersion` are mutually exclusive on the API resource the way `ci.yml` input treats `version`/`schemaVersion` per schema version — on the returned resource, **both are present and required**.
- Treating `status` as a flat object with an optional `deletedAt`/`detailsRef` — it's a discriminated union; only fetch `deletedAt` when `state` is `deleting`/`deleted`, only fetch `detailsRef` when `state` is `synchronized`.
- Assuming `backups.deleteRetentionDays`/`intervalH`/`config` are always present — they only exist when `backups.enabled === true`.
- Assuming `recentEvents` or top-level `version` are always present — both are optional per the schema's `required` list.
- Validating `name` without accounting for the pattern's edge cases — no leading whitespace, but internal spaces are allowed; max 126 characters before the final character, 127 total.
- Assuming `plan.parameters` values can be non-integer (e.g. decimals for `cpu`) — the top-level resource schema types every `plan.parameters` value as `integer`, even though some provider plan tables (e.g. PostgreSQL's `cpu`) show `number` with fractional defaults — reconcile against the specific provider's plan schema if a fractional value is rejected.

## Known Documentation Discrepancies

> This section may be outdated. Please verify against the official documentation:
> <https://docs.codesphere.com/managed-services/overview>

- **Resolved (previously flagged, now confirmed):** the Virtual Kubernetes Cluster provider name is `virtual-k8s` — confirmed by the user against <https://docs.codesphere.com/managed-services/providers/virtual-kubernetes-cluster>. `ci-pipeline.md` and `migration-guide.md` (in the `landscape/` reference set) have been corrected accordingly; they previously said `virtual-kubernetes-cluster`.
- **Open:** this schema shows `provider.version` and `provider.schemaVersion` as both required simultaneously on the resource, which is not obviously consistent with `ci-pipeline.md`'s description of `ci.yml` input using only one of the two fields depending on `schemaVersion` (v0.2/v0.3 vs. v0.4). The likely explanation is that the API normalizes/populates both on the stored resource regardless of which one was supplied at creation — but this reference set does not have direct confirmation of that normalization behavior. Treat `ci.yml` input rules (input side) and this resource schema (output side) as two different surfaces rather than assuming a strict 1:1 mapping.
- The `plan.parameters` values are typed as `integer` at this general resource-schema level, while some individual provider plan tables (e.g. PostgreSQL's `cpu`, priced as `cpu-tenths`) are typed as `number` with fractional defaults (e.g. `1`) in the provider-specific docs — if a decimal `cpu` value is rejected against this general schema, the provider-specific plan schema should be treated as authoritative for that field.

## Further Reading

- Official docs: <https://docs.codesphere.com/managed-services/overview>
- Managed services architecture & lifecycle: [README.md](./README.md)
- `ci.yml` Managed Service field shape (input side): `../landscape/ci-pipeline.md`
- Per-provider `config`/`secrets`/`details` schemas: sibling files in this folder (e.g. `postgresql.md`, `valkey.md`)
- Interactive API schema: `https://<instance>/api/scalar-ui`, tag `managed-services`
