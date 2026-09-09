# Codesphere Managed Services Reference

> **Last updated:** 2026-07-28 · **Source:** https://docs.codesphere.com/managed-services/overview

> The documentation changes regularly. If you are facing issues, please refer to the official documentation:
> <https://docs.codesphere.com/managed-services/overview>

## Overview

Managed Services let a Codesphere team provision backing infrastructure (databases, message queues, object storage, Kubernetes clusters, ...) declaratively, either through the UI, the public API, or as part of a Landscape (`ci.yml`). Three elements work together in a continuous reconciliation loop: desired state is described via UI/API, Codesphere stores it, and a reconciler polls backends to match reality to that state.

Read this reference before generating any `ci.yml` snippet or API call that touches managed services, and before pointing a user at a specific provider (`postgres`, `babelfish`, `ferretdb`, `virtual-k8s`, `s3`, `opensearch`, `rabbitmq`, `valkey`). Provider-specific details live in the sibling files in this folder.

## Core Concepts

- **Service Provider**: defines _what_ a service is and its configuration schema (config/secrets/details). A reproducible blueprint; one provider can back many independent deployments.
- **Provider Backend**: connects a provider to actual compute; executes create/update/delete and reports status back to the reconciler.
- **Service Infrastructure**: where the service actually runs — inside the Codesphere cluster for core (Codesphere-managed) services, or anywhere chosen (hyperscaler, on-prem) for custom services.
- **Codesphere-managed vs. Ecosystem/self-managed**: Codesphere-managed providers (PostgreSQL, Babelfish, Object Storage, ...) are fully operated by Codesphere across all layers. Self-managed providers are either **Landscape-based** (deployed as a Landscape workload, orchestrated via the Landscape API, no backend code needed) or **REST-based** (a custom backend implementing the Provider REST API spec against any infrastructure).
- **`provider.yml`**: the manifest that turns a Landscape repo into a one-click catalog entry (see below).
- **`x-update-constraint`**: JSON Schema extension restricting how a `configSchema` property may change after creation (`increase-only`, `immutable`).
- **`x-endpoint`**: JSON Schema extension on a `detailsSchema` property that fetches its value live (GET) from the running service.

## API / Syntax

### `GET /api/managed-services/providers`

- **Description:** Returns the authoritative current provider catalog for the target installation — names, versions, plans, capabilities (`pause`, `backups`). The catalog differs between cloud and self-hosted installations and changes over time; call this before assuming a provider name/version exists.
- **Parameters:**

| Name            | Type   | Required | Description        |
| --------------- | ------ | -------- | ------------------ |
| `Authorization` | header | Yes      | `Bearer $CS_TOKEN` |

- **Returns:** JSON array of provider definitions (name, version, category, scope, capabilities, plans).
- **Example:**

```bash
curl -H "Authorization: Bearer $CS_TOKEN" \
  "https://<instance>/api/managed-services/providers"
```

### `POST /api/managed-services`

- **Description:** Deploys a new standalone managed service instance.
- **Parameters:**

| Name                                 | Type    | Required                 | Description                                                                                                                                                                                                                                                                                                        |
| ------------------------------------ | ------- | ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `teamId`                             | integer | Yes                      | Owning team.                                                                                                                                                                                                                                                                                                       |
| `name`                               | string  | Yes                      | Unique service name.                                                                                                                                                                                                                                                                                               |
| `provider.name` / `provider.version` | string  | Yes                      | Target provider and schema version. **The confirmed resource schema also requires `provider.schemaVersion` on the stored object** (see [service-resource-schema.md](./service-resource-schema.md)) — send both `version` and `schemaVersion` with the same value if the API rejects a request containing only one. |
| `plan.id` / `plan.parameters`        | object  | Yes                      | Plan selection and plan-specific parameters (e.g. storage in GB).                                                                                                                                                                                                                                                  |
| `config`                             | object  | No                       | Provider-specific config fields.                                                                                                                                                                                                                                                                                   |
| `secrets`                            | object  | Yes (provider-dependent) | Provider-specific credentials.                                                                                                                                                                                                                                                                                     |
| `backups`                            | object  | No                       | Backup configuration; provider must support it.                                                                                                                                                                                                                                                                    |
| `recoverFrom`                        | object  | No                       | Bootstrap from a backup / point-in-time; requires `backups` config+secrets.                                                                                                                                                                                                                                        |

- **Returns:** The created managed service resource (status starts `Creating` → `Synchronized`).
- **Example:**

```bash
curl -X POST "https://<instance>/api/managed-services" \
  -H "Authorization: Bearer $CS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "teamId": 123,
    "name": "my-postgres-db",
    "provider": { "name": "postgresql", "version": "15" },
    "plan": { "id": 1, "parameters": { "storage": 10 } },
    "config": { "max_connections": "100" },
    "secrets": { "username": "admin", "password": "secure-password" },
    "backups": {
      "enabled": true,
      "intervalH": 24,
      "deleteRetentionDays": 7,
      "config": {
        "endpointUrl": "https://s3.eu-central-1.amazonaws.com",
        "destinationPath": "s3://my-codesphere-backups/"
      },
      "secrets": { "accessKey": "...", "secretKey": "..." }
    }
  }'
```

### `PATCH /api/managed-services/{id}`

- **Description:** Partial update of `config`, `plan`, `secrets`, `backups`, or `name`. Not every field stays editable after creation — check the provider's capabilities. Some updates restart the service (brief downtime). Also used for pause/resume: `{"pause": true}` / `{"pause": false}`.
- **Parameters:**

| Name | Type   | Required | Description                                                            |
| ---- | ------ | -------- | ---------------------------------------------------------------------- |
| `id` | path   | Yes      | Managed service ID.                                                    |
| body | object | Yes      | Any subset of `config`, `plan`, `secrets`, `backups`, `name`, `pause`. |

- **Returns:** Updated managed service resource.
- **Example:**

```bash
curl -X PATCH "https://<instance>/api/managed-services/<id>" \
  -H "Authorization: Bearer $CS_TOKEN" -H "Content-Type: application/json" \
  -d '{"pause": true}'
```

### `DELETE /api/managed-services/{id}`

- **Description:** Always a soft delete first (`deletedAt` timestamp, status `Deleting`). If backups are enabled, a final backup is taken before deletion. Record stays visible under "Show recently Deleted" during a retention window, then is hard-deleted — **permanent, compute and volume cannot be recovered** once hard delete happens.
- **Parameters:**

| Name | Type | Required | Description         |
| ---- | ---- | -------- | ------------------- |
| `id` | path | Yes      | Managed service ID. |

- **Returns:** 202/204 on accepted soft delete.

### `POST /api/managed-services/{id}/backups`

- **Description:** Triggers a manual backup for a service that supports backups.
- **Parameters:**

| Name | Type | Required | Description         |
| ---- | ---- | -------- | ------------------- |
| `id` | path | Yes      | Managed service ID. |

- **Returns:** Backup job reference.

### `POST /api/managed-services/providers` (publish a custom provider)

- **Description:** Publishes a Landscape-based provider, either from a Git URL (Codesphere fetches/validates `provider.yml`) or a full inline spec. **Requires cluster admin permissions** for `scope.type: global`; team admins can request a team-scoped provider.
- **Parameters:**

| Name                                                                  | Type             | Required    | Description                                                          |
| --------------------------------------------------------------------- | ---------------- | ----------- | -------------------------------------------------------------------- |
| `gitUrl`                                                              | string           | Conditional | Repo containing `provider.yml` (Git-URL method).                     |
| `gitRef`                                                              | string           | No          | Branch/ref; defaults to the repo's default branch.                   |
| `name`, `version`, `author`, `displayName`, `category`, `description` | string           | Conditional | Required for the inline-spec method.                                 |
| `backend.landscape.gitUrl` / `ciProfile`                              | string           | Conditional | Landscape repo + which `ci.<profile>.yml` to deploy (inline method). |
| `configSchema` / `secretsSchema` / `detailsSchema`                    | JSON Schema      | No          | User config, vault secrets, and runtime details schemas.             |
| `scope.type`                                                          | `global`\|`team` | Yes         | `global` needs cluster admin; `team` needs `teamIds`.                |

- **Returns:** The created provider definition.
- **Example (Git URL method):**

```bash
curl -X POST "https://<instance>/api/managed-services/providers" \
  -H "Authorization: Bearer $CS_TOKEN" -H "Content-Type: application/json" \
  -d '{
    "gitUrl": "https://github.com/your-org/mattermost-landscape",
    "gitRef": "my-branch",
    "scope": { "type": "global" }
  }'
```

## Common Pitfalls

- Assuming MySQL or MongoDB (native wire protocol) exist as first-party providers because they were common in older docs — the current equivalents are Babelfish (SQL Server/TDS) and DocumentDB (MongoDB-compatible via FerretDB); there is no dedicated `mysql` entry in the core catalog.
- Hardcoding provider names/versions/plans instead of calling `GET /api/managed-services/providers` first — the catalog differs between cloud and self-hosted installations and changes over time.
- Assuming a provider supports `pause` or `backups` without checking `capabilities` — not all providers do.
- Forgetting that `DELETE` is a soft delete with a retention window, then being surprised the record still shows up — but also forgetting the _hard_ delete afterwards is irreversible.
- Publishing a `global`-scope custom provider without cluster admin permissions (use `team` scope + `teamIds` instead).
- Assuming all `configSchema` fields stay editable post-creation — check `x-update-constraint` (`immutable` / `increase-only`).

## Known Documentation Discrepancies

> This section may be outdated. Please verify against the official documentation:
> <https://docs.codesphere.com/managed-services/overview>

- The provider catalog listed in this reference set (PostgreSQL, Babelfish, DocumentDB, Object Storage, Virtual Kubernetes Cluster, OpenSearch, RabbitMQ, Valkey) is a snapshot; several of these providers are explicitly marked preview/closed-testing in the source docs and may be renamed, removed, or change schema without notice. Always confirm via `GET /api/managed-services/providers` against the target instance.
- The custom REST backend contract (`create-custom-rest-backend`) is referenced but not reproduced in this reference set — fetch it live if a user needs to implement a REST-based provider backend.
- The point-in-time recovery workflow is only reproduced in provider-specific depth for PostgreSQL/Babelfish/DocumentDB; the general `managed-services/backups` page may contain additional detail not captured here.

## Further Reading

- Official docs: <https://docs.codesphere.com/managed-services/overview>
- API reference (Swagger/Scalar UI): `https://<instance>/api/scalar-ui`, tag `managed-services`
- Backups: <https://docs.codesphere.com/managed-services/backups>
- Custom REST backend spec: <https://docs.codesphere.com/managed-services/create-custom-rest-backend>
- Full Managed Service resource schema (API response shape): [service-resource-schema.md](./service-resource-schema.md)
- CLI & API companion reference: `../landscape/cli-and-api.md`
