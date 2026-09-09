# Custom Landscape-Based Service Provider (`provider.yml`) Reference

> **Last updated:** 2026-07-28 · **Source:** https://docs.codesphere.com/managed-services/create-custom-landscape-provider

> The documentation changes regularly. If you are facing issues, please refer to the official documentation:
> <https://docs.codesphere.com/managed-services/create-custom-landscape-provider>

## Overview

Turns an existing Codesphere Landscape into a catalog entry that other teams can one-click deploy, without writing any backend code — the provider backend is the Landscape itself, orchestrated via the Landscape API. Use this reference when a user wants to publish their own app (e.g. Mattermost, Grafana, Nextcloud) as a managed-service provider. **Requires cluster admin permissions to publish globally**; team admins can request a team-scoped provider instead.

## Core Concepts

- **`provider.yml`**: manifest at the repo root describing the provider's identity, backend, and config/secrets/details schemas. Only needed for the Git-URL publishing method.
- **Publishing methods**: (1) Git URL — Codesphere fetches and validates `provider.yml` from the repo; (2) full inline spec — no `provider.yml` file needed, everything passed in the API body.
- **`configSchema`**: JSON Schema of user-configurable options; rendered in the running service's **config** settings tab, editable later, passed to the Landscape as env vars.
- **`secretsSchema`**: JSON Schema of secrets; collected in the create-service dialog, must be filled at creation, injected into the Landscape's vault.
- **`detailsSchema`**: JSON Schema of runtime details; rendered in the service's **details** tab.
- **`x-update-constraint`**: restricts how a `configSchema` property may change _after_ creation — not enforced on initial creation. Values: `increase-only` (new ≥ current), `immutable` (cannot change once set).
- **`x-endpoint`**: on a `detailsSchema` property, fetches its value live from the running service via a GET request that must return JSON matching the property schema.

## API / Syntax

### `provider.yml` — Manifest Fields

| Name                          | Type              | Required | Description                                            |
| ----------------------------- | ----------------- | -------- | ------------------------------------------------------ |
| `name`                        | string            | Yes      | Unique id, must match `^[-a-z0-9_]+$`.                 |
| `version`                     | string            | Yes      | `v[0-9]+`, e.g. `v1`.                                  |
| `author`                      | string            | No       | Org/individual responsible.                            |
| `displayName`                 | string            | Yes      | Shown in the Marketplace UI.                           |
| `iconUrl`                     | string (uri)      | No       | Provider icon.                                         |
| `category`                    | string            | No       | Grouping, e.g. `databases`, `messaging`, `monitoring`. |
| `description`                 | string (markdown) | No       | Rendered provider description.                         |
| `backend.landscape.gitUrl`    | string (uri)      | Yes      | Repo containing the Landscape.                         |
| `backend.landscape.ciProfile` | string            | Yes      | Which `ci.<profile>.yml` to deploy.                    |
| `configSchema`                | JSON Schema       | No       | User-configurable options → Landscape env vars.        |
| `secretsSchema`               | JSON Schema       | No       | Secrets → Landscape vault entries.                     |
| `detailsSchema`               | JSON Schema       | No       | Runtime details shown post-provisioning.               |

- **Example:**

```yaml
name: mattermost
version: v1
author: Your Team
displayName: Mattermost
iconUrl: https://example.com/mattermost-icon.png
category: collaboration
description: |
  Open-source team messaging and collaboration platform.
  Supports channels, direct messaging, and file sharing.

backend:
  landscape:
    gitUrl: https://github.com/your-org/mattermost-landscape
    ciProfile: production

configSchema:
  type: object
  properties:
    SITE_NAME:
      type: string
      description: Display name for your Mattermost instance
    MAX_USERS:
      type: integer
      x-update-constraint: increase-only

secretsSchema:
  type: object
  properties:
    ADMIN_PASSWORD:
      type: string
      format: password

detailsSchema:
  type: object
  properties:
    hostname: { type: string }
    port: { type: integer }
```

### `POST /api/managed-services/providers` — Publish via Git URL

- **Description:** Simplest method — Codesphere fetches/validates `provider.yml` from the repo. Omit `gitRef` to use the default branch.
- **Parameters:**

| Name            | Type             | Required    | Description                                          |
| --------------- | ---------------- | ----------- | ---------------------------------------------------- |
| `gitUrl`        | string           | Yes         | Repo containing `provider.yml` at its root.          |
| `gitRef`        | string           | No          | Branch/tag/ref to use instead of the default branch. |
| `scope.type`    | `global`\|`team` | Yes         | `global` requires cluster admin.                     |
| `scope.teamIds` | integer[]        | Conditional | Required when `scope.type` is `team`.                |

- **Example:**

```bash
curl -X POST "https://<instance>/api/managed-services/providers" \
  -H "Authorization: Bearer $CS_TOKEN" -H "Content-Type: application/json" \
  -d '{
    "gitUrl": "https://github.com/your-org/mattermost-landscape",
    "gitRef": "my-branch",
    "scope": { "type": "global" }
  }'
```

### `POST /api/managed-services/providers` — Publish via Inline Spec

- **Description:** No `provider.yml` file needed in the repo — the entire manifest is passed in the request body.
- **Parameters:** same fields as the `provider.yml` manifest table above, passed as top-level JSON body fields, plus `plans` (array, can be empty) and `scope`.
- **Example:**

```bash
curl -X POST "https://<instance>/api/managed-services/providers" \
  -H "Authorization: Bearer $CS_TOKEN" -H "Content-Type: application/json" \
  -d '{
    "name": "mattermost", "version": "v1", "author": "Your Team",
    "displayName": "Mattermost", "category": "collaboration",
    "description": "Open-source team messaging platform.",
    "backend": { "landscape": { "gitUrl": "https://github.com/your-org/mattermost-landscape", "ciProfile": "production" } },
    "configSchema": { "type": "object", "properties": { "SITE_NAME": { "type": "string" } } },
    "secretsSchema": { "type": "object", "properties": { "ADMIN_PASSWORD": { "type": "string", "format": "password" } } },
    "detailsSchema": { "type": "object", "properties": { "hostname": { "type": "string" }, "port": { "type": "integer" } } },
    "plans": [],
    "scope": { "type": "team", "teamIds": [42, 43] }
  }'
```

### Consuming Config/Secrets Inside the Landscape

- **Description:** `configSchema` properties become workspace env vars (`workspace.env['NAME']`); `secretsSchema` properties are injected into the vault and referenced with `${{ vault.NAME }}`.
- **Example:**

```yaml
schemaVersion: v0.4
run:
  my-service:
    steps:
      - command: ./start.sh
    env:
      APP_VERSION: ${{ workspace.env['APP_VERSION'] }}
      ADMIN_PASSWORD: ${{ vault.ADMIN_PASSWORD }}
```

### `x-update-constraint` — Example

- **Example:**

```yaml
configSchema:
  type: object
  properties:
    storage:
      type: integer
      x-update-constraint: increase-only # new value must be >= current
    version:
      type: string
      enum: ["17.6", "16.10", "15.14"]
      x-update-constraint: immutable # cannot change once set
```

### `x-endpoint` — Example

- **Example:**

```yaml
detailsSchema:
  type: object
  properties:
    hostname: { type: string }
    port: { type: integer }
    status:
      type: object
      properties: { state: { type: string }, uptime: { type: number } }
      x-endpoint: "https://{{hostname}}:{{port}}/status"
```

### Supported JSON Schema `format` values

`int32`, `int64`, `float`, `double`, `byte`, `binary`, `date`, `date-time`, `password`, `uri`, `hostname`

## Common Pitfalls

- Publishing with `scope.type: global` without cluster admin permissions — request a `team`-scoped provider instead.
- Assuming `x-update-constraint` is enforced at creation time — it only restricts _updates_ after the resource exists.
- Forgetting that `configSchema` values are editable post-creation by default (unless constrained) but `secretsSchema` values are collected only once, at creation.
- Using a `name` that doesn't match `^[-a-z0-9_]+$`, or a `version` that isn't `v[0-9]+`.
- Pointing `backend.landscape.ciProfile` at a `ci.<profile>.yml` that doesn't exist in the target repo.

## Known Documentation Discrepancies

> This section may be outdated. Please verify against the official documentation:
> <https://docs.codesphere.com/managed-services/create-custom-landscape-provider>

- The custom **REST-based** backend contract (for providers pointing at non-Landscape infrastructure) is a separate, larger spec not reproduced here — fetch `https://docs.codesphere.com/managed-services/create-custom-rest-backend` live if a user needs to implement one.
- The exact validation regex for `name` (`^[-a-z0-9_]+$`) and `version` (`v[0-9]+`) should be re-confirmed against the current API schema (`/api/scalar-ui`) before being relied on for client-side validation.

## Further Reading

- Official docs: <https://docs.codesphere.com/managed-services/create-custom-landscape-provider>
- Custom REST backend spec: <https://docs.codesphere.com/managed-services/create-custom-rest-backend>
- Managed Services overview: [README.md](./README.md)
- API reference (Swagger/Scalar UI): `https://<instance>/api/scalar-ui`, tag `managed-services`
