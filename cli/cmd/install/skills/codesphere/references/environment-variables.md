# Codesphere Environment Variables Reference

> **Last updated:** 2026-07-28 · **Source:** https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/environment-variables

> The documentation changes regularly. If you are facing issues, please refer to the official documentation:
> <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/environment-variables>

## Overview

Plain, workspace-scoped environment variables — set once, available to every service's `run` stage. For encrypted, Landscape-scoped secrets instead, see [secret-management.md](./secret-management.md).

## Core Concepts

- **Scope**: workspace-scoped, not per-service — available to every service's `run` stage automatically. Not automatically available to `prepare`/`test` unless referenced there too.
- **Visibility**: hidden by default in the UI (eye icon to reveal).
- **Apply on change**: changing a value requires re-running the CI Pipeline `run` stage to take effect — it does not hot-reload.
- **Naming rule**: `name` must match `^[A-Za-z_][A-Za-z0-9_.-]*$`.

## API / Syntax

### UI

- **Description:** Setup > Environment Variables.

### `GET /workspaces/{workspaceId}/env-vars`

- **Description:** List plain env vars.
- **Parameters:**

| Name          | Type | Required | Description       |
| ------------- | ---- | -------- | ----------------- |
| `workspaceId` | path | Yes      | Target workspace. |

### `PUT /workspaces/{workspaceId}/env-vars`

- **Description:** Set env vars.
- **Parameters:**

| Name          | Type                     | Required | Description                                     |
| ------------- | ------------------------ | -------- | ----------------------------------------------- |
| `workspaceId` | path                     | Yes      | Target workspace.                               |
| body          | array of `{name, value}` | Yes      | `name` must match `^[A-Za-z_][A-Za-z0-9_.-]*$`. |

- **Example:**

```bash
curl -X PUT "https://codesphere.com/api/workspaces/456/env-vars" \
  -H "Authorization: Bearer <token>" -H "Content-Type: application/json" \
  -d '[{"name": "DATABASE_URL", "value": "postgres://..."}]'
```

### `DELETE /workspaces/{workspaceId}/env-vars`

- **Description:** Remove env vars.
- **Parameters:**

| Name          | Type | Required | Description       |
| ------------- | ---- | -------- | ----------------- |
| `workspaceId` | path | Yes      | Target workspace. |

### At Workspace Creation

- **Description:** Env vars can also be set at creation time via the `env` array on `POST /workspaces` (array of `{name, value}`).

### Built-in Variables

| Name                   | Description                                                            |
| ---------------------- | ---------------------------------------------------------------------- |
| `CS_REPLICA`           | Per-replica id — use it to avoid file-write conflicts across replicas. |
| `WORKSPACE_ID`         | The workspace's own ID.                                                |
| `NV_LIBCUBLAS_VERSION` | Present when GPU resources are available.                              |

### Inline in Commands

- **Example:**

```yaml
steps:
  - name: Run server
    command: PYTHONPATH=/home/user/app/pipLib PORT=3000 python3 server.py
```

Or source a `.env` file:

```yaml
command: . .env && export MY_VAR=$MY_VAR && npm start
```

## Common Pitfalls

- Expecting a changed env var to apply without re-running the `run` stage — it doesn't hot-reload.
- Referencing an env var in `prepare`/`test` that was only set for `run` — plain env vars are automatically available in `run`, not the other two stages, unless explicitly referenced/sourced there.
- Forgetting framework-specific build-time prefixes (e.g. Create React App needs `REACT_APP_`-prefixed vars) — a var without the expected prefix silently isn't picked up at build time.
- Manually creating a replica and expecting env vars to carry over automatically — they must be copied manually.
- Using a `name` that doesn't match `^[A-Za-z_][A-Za-z0-9_.-]*$` in a `PUT /env-vars` call — rejected.

## Known Documentation Discrepancies

> This section may be outdated. Please verify against the official documentation:
> <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/environment-variables>

- Whether plain env vars are ever automatically available to `prepare`/`test` (vs. only `run`) is stated as "not automatically" in the source docs — if this changes per schema version, re-verify against the live docs.

## Further Reading

- Official docs: <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/environment-variables>
- Vault secrets (encrypted, Landscape-scoped): [secret-management.md](./secret-management.md)
- CI pipeline field reference: [ci-pipeline.md](./ci-pipeline.md)
