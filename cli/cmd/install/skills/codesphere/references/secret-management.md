# Codesphere Secret Management (Vault) Reference

> **Last updated:** 2026-07-28 · **Source:** https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/secret-management

> The documentation changes regularly. If you are facing issues, please refer to the official documentation:
> <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/secret-management>

> Not covered by the `cs-go` CLI — use the Public API endpoints below directly. See [cli-and-api.md](./cli-and-api.md) for the full CLI-vs-API coverage map.

## Overview

Vault secrets are encrypted-at-rest, Landscape-scoped credentials, distinct from plain workspace env vars (see [environment-variables.md](./environment-variables.md)). **Preview feature** — backed by OpenBao, may change. Only exposed to services that explicitly reference them, and cleaned up automatically when the workspace is deleted.

## Core Concepts

- **Reference, don't inline**: `ci.yml` only _references_ a vault key (`${{ vault.NAME }}`) — the real value is never written into the YAML.
- **Sync flow**: on **Sync Landscape**, any referenced vault key that doesn't exist yet triggers a modal to enter it once; it's then persisted in OpenBao and injected at deploy time. **A Landscape sync fails if a referenced vault key was never initialized.**
- **Shared vaults**: a team-level vault partition any number of workspaces on the team can point at, instead of each workspace having its own partition.
- **Mutual exclusivity**: a workspace resolves `${{ vault.* }}` against **either** its own partition **or** its assigned shared vault — never both merged.

## API / Syntax

### Referencing a Secret in `ci.yml`

- **Example:**

```yaml
run:
  secret-demo:
    steps:
      - command: echo $SECRET_KEY
    plan: 8
    replicas: 1
    network:
      ports: [{ port: 3000, isPublic: false }]
      paths: []
    env:
      SECRET_KEY: ${{ vault.secretFoo }}
      PLAIN_ENV: foo
```

The env var **name** (`SECRET_KEY`) is whatever the service expects; the **value** references the vault key (`secretFoo`) via `${{ vault.NAME }}`.

### Template Syntax

| Template                      | Resolves to                                                                                                                                                                                                               |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `${{ vault.NAME }}`           | A secret from the vault (workspace's own, or its assigned shared vault)                                                                                                                                                   |
| `${{ workspace.id }}`         | The workspace ID                                                                                                                                                                                                          |
| `${{ workspace.devDomain }}`  | The workspace's dev domain                                                                                                                                                                                                |
| `${{ team.id }}`              | The team ID                                                                                                                                                                                                               |
| `${{ workspace.env['KEY'] }}` | A global workspace environment variable — useful for remapping, e.g. `PG_USER: ${{ workspace.env['BACKEND_PG_USER'] }}` when a backend expects `PG_USER` but the value is stored globally under a collision-avoiding name |

**Confirmed live:** any of the above can appear inside a plain `env:` value concatenated with literal text, not just as the entire value — e.g. building a service-to-service internal DNS URL (see [landscape.md](./landscape.md)'s internal DNS pattern) directly in a Reactive's `env:` block:

```yaml
run:
  frontend:
    env:
      BACKEND_URL: http://ws-server-${{ workspace.id }}-backend.workspaces:3000/api
```

resolves at deploy time to the real workspace ID substituted into the string, and the resulting URL is genuinely reachable from inside the Landscape.

### `POST /vault/teams/{teamId}/workspaces/{workspaceId}`

- **Description:** Store one or more secrets in a workspace's own vault partition.
- **Parameters:**

| Name                     | Type   | Required | Description              |
| ------------------------ | ------ | -------- | ------------------------ |
| `teamId` / `workspaceId` | path   | Yes      | Target team/workspace.   |
| body                     | object | Yes      | `{"KEY": "value", ...}`. |

- **Returns:** `{"KEY": "<revision>"}` per key.

### `GET /vault/teams/{teamId}/workspaces/{workspaceId}/keys`

- **Description:** List secret **keys only** — values are never exposed via the API.

### `DELETE /vault/teams/{teamId}/workspaces/{workspaceId}`

- **Description:** Delete secrets.
- **Parameters:**

| Name | Type             | Required | Description         |
| ---- | ---------------- | -------- | ------------------- |
| body | array of strings | Yes      | `["KEY1", "KEY2"]`. |

### `POST /vault/teams/{teamId}/workspaces/{workspaceId}/generate`

- **Description:** Auto-generate cryptographically random secrets server-side — the plaintext never has to leave the vault.
- **Parameters:**

| Name | Type   | Required | Description                                                      |
| ---- | ------ | -------- | ---------------------------------------------------------------- |
| body | object | Yes      | Policy per key: `length`, `rules[].charset`, `rules[].minChars`. |

- **Returns:** the generated `value` once, plus a `revision`, per key.
- **Example:**

```bash
curl -X POST "https://<instance>/api/vault/teams/123/workspaces/456/generate" \
  -H "Authorization: Bearer $CS_TOKEN" -H "Content-Type: application/json" \
  -d '{
    "DATABASE_PASSWORD": {
      "length": 16,
      "rules": [{"charset": "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", "minChars": 16}]
    }
  }'
```

Then reference it in `ci.yml` as `${{ vault.DATABASE_PASSWORD }}` — no plaintext value ever has to pass through the CI/CD pipeline.

### Shared Vaults — Endpoints

| Method   | Path                                               | Purpose                                                              |
| -------- | -------------------------------------------------- | -------------------------------------------------------------------- |
| `GET`    | `/vault/teams/{teamId}/shared`                     | List shared vault names                                              |
| `POST`   | `/vault/teams/{teamId}/shared`                     | Create an empty shared vault. Body: `{"name": "..."}`                |
| `DELETE` | `/vault/teams/{teamId}/shared/{vaultName}`         | Delete a shared vault and **all** its secrets                        |
| `POST`   | `/vault/teams/{teamId}/shared/{vaultName}/secrets` | Store secrets in it (same semantics as the workspace store endpoint) |
| `GET`    | `/vault/teams/{teamId}/shared/{vaultName}/keys`    | List its secret keys                                                 |
| `DELETE` | `/vault/teams/{teamId}/shared/{vaultName}/secrets` | Delete named secrets (vault itself stays)                            |

### Shared Vaults — Behavior Notes

- Zero-setup: workspaces assigned to it can reference its secrets immediately, no per-workspace init step.
- Single source of truth for credentials common across environments; central rotation — update once, applies to every workspace on next deploy.
- Switching a workspace to a shared vault requires every referenced key to already exist there, or the next sync fails.
- Assign at workspace creation (Advanced Options > Secrets Vault dropdown) or later on an existing workspace.
- Listing a team's shared vaults needs team **read** access; create/delete/modify needs team **write** access.
- Shared vaults are **not** deleted when a workspace using them is deleted — they persist until explicitly deleted.

## Common Pitfalls

- Deploying a Landscape that references `${{ vault.NAME }}` for a key that was never initialized — sync fails; provide the value during sync or pre-populate via the API first.
- Assuming a service can read any vault secret in the workspace — it can only read what it explicitly declares in its own `env:` block.
- Assuming vault and plain env vars merge — a workspace resolves `${{ vault.* }}` against either its own partition or its assigned shared vault, never both.
- Switching a workspace to a shared vault without pre-populating every key it references — the next sync fails.
- Expecting `GET .../keys` to return secret values — it only ever returns key names.
- Deleting a workspace and expecting its shared vault to go with it — shared vaults persist independently until explicitly deleted.

## Known Documentation Discrepancies

> This section may be outdated. Please verify against the official documentation:
> <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/secret-management>

- This entire feature is explicitly marked **preview** in the source docs — field names, endpoint paths, and the sync-flow UX may change before general availability. Re-verify against the live docs/Scalar UI before depending on this for production secret rotation workflows.

## Further Reading

- Official docs: <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/secret-management>
- Plain environment variables: [environment-variables.md](./environment-variables.md)
- CLI & API reference (what the CLI does _not_ cover): [cli-and-api.md](./cli-and-api.md)
