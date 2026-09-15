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

| Name                   | Description                                                                                                                               |
| ---------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `CS_REPLICA`           | Per-replica id — use it to avoid file-write conflicts across replicas.                                                                    |
| `WORKSPACE_ID`         | The workspace's own **numeric** ID — the value `cs -w` / `CS_WORKSPACE_ID` takes. Injected into the tmux session environment (see below). |
| `CODESPHERE_APP_ID`    | The workspace's app id. Injected into the tmux session environment (see below).                                                           |
| `WORKSPACE_DEV_DOMAIN` | The workspace's dev domain. Injected into the tmux session environment (see below).                                                       |
| `NV_LIBCUBLAS_VERSION` | Present when GPU resources are available.                                                                                                 |

### Reading the Built-ins From Inside a Workspace

An agent running *inside* a workspace needs `WORKSPACE_ID` to target that same workspace with `cs -w` / `CS_WORKSPACE_ID`. **Confirmed live** (`cloud.codesphere.com`, `cs` 1.38.0, `workspace-ssh` enabled): `WORKSPACE_ID`, `CODESPHERE_APP_ID` and `WORKSPACE_DEV_DOMAIN` are injected into the workspace's **tmux session environment**, not into the container environment. Whether `printenv` shows them depends entirely on how the current shell was started:

| Where the shell came from                                        | `printenv WORKSPACE_ID`                                                                                                                                                                    |
| ---------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Started by tmux (the IDE terminal)                               | Set — inherited from the tmux environment.                                                                                                                                                  |
| `workspace-ssh`, VS Code Remote, or an editor-extension terminal | **Empty.** The workspace runtime spawns the session shell with a fresh minimal environment (roughly `HOME`, `PATH`, `SSH_*`, `VSCODE_*`). Toolchain vars — `NODE_VERSION`, `FNM_DIR`, `GOPATH` — are missing for the same reason. |
| The container environment (`/proc/1/environ`)                    | **Empty.** Contains none of the three.                                                                                                                                                       |

The tmux **global** environment is the reliable source, readable from any shell in the workspace:

```bash
WORKSPACE_ID=$(tmux show-environment -g WORKSPACE_ID | cut -d= -f2)

export CS_WORKSPACE_ID="$WORKSPACE_ID"
# or pass it per command:
cs start pipeline -w "$WORKSPACE_ID" run
```

- **It must be `-g`.** `tmux show-environment -t <session> WORKSPACE_ID` returns `unknown variable` — the value lives in the global environment, not in any individual session's.
- A command launched via `tmux new-session` inherits `WORKSPACE_ID` and `WORKSPACE_DEV_DOMAIN` from that global environment — an alternative to reading the value out explicitly when the process is being spawned that way anyway.
- **`HOSTNAME` is not a substitute** — it carries a workspace **UUID**, not the numeric ID `-w` takes.

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
- **Concluding from an empty `printenv` that a workspace has no `WORKSPACE_ID`.** Over `workspace-ssh` or in a VS Code Remote / editor-extension terminal the session shell starts with a fresh minimal environment, so none of the tmux-injected built-ins appear. Read `tmux show-environment -g WORKSPACE_ID` before concluding anything.
- **Falling back to `cs list workspaces` to guess which row is "this" workspace.** API tokens are usually team- or org-wide, so the listing contains other people's workspaces and a wrong guess targets one of them — `cs start pipeline ... run` then restarts whatever is running there. Resolve the ID from the tmux global environment instead.

## Known Documentation Discrepancies

> This section may be outdated. Please verify against the official documentation:
> <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/environment-variables>

- Whether plain env vars are ever automatically available to `prepare`/`test` (vs. only `run`) is stated as "not automatically" in the source docs — if this changes per schema version, re-verify against the live docs.

## Further Reading

- Official docs: <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/environment-variables>
- Vault secrets (encrypted, Landscape-scoped): [secret-management.md](./secret-management.md)
- CI pipeline field reference: [ci-pipeline.md](./ci-pipeline.md)
- Targeting a workspace with the CLI (`-w` / `CS_WORKSPACE_ID`): [cli-and-api.md](./cli-and-api.md)
