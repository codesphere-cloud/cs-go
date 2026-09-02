# Codesphere CLI (`cs-go`) & Public API Reference

> **Last updated:** 2026-07-28 · **Source:** https://docs.codesphere.com (public API + `cs-go` docs)

> The documentation changes regularly. If you are facing issues, please refer to the official documentation:
> <https://docs.codesphere.com>

## Overview

Two ways to drive Codesphere programmatically: the open-source `cs-go` CLI (pre-installed in every workspace as `cs`) for day-to-day workspace/pipeline operations, and the Public API for everything else (managed services, vault, domains, teams, landscape lifecycle, ...). The full interactive schema for any endpoint is always at `https://<instance>/api/scalar-ui` (cloud: `https://cloud.codesphere.com/api/scalar-ui`) — treat that as the source of truth over this file, since the API evolves.

## Core Concepts

- **`cs-go`**: open-source Go CLI/SDK (`codesphere-cloud/cs-go`), covers workspace/pipeline day-to-day use only — not the full API surface.
- **The CLI is not a complete API client** — vault/secrets, managed services, domains, teams/organizations, explicit Landscape lifecycle (deploy/scale/teardown as a unit), usage/metrics, cluster/org admin, SSH keys, and metadata lookups all require going straight to the Public API.
- **Auth**: bearer token from Settings > API Keys (`CS_TOKEN` for the CLI).

## API / Syntax

### `cs-go` — Installation (outside a workspace)

- **Example (via GitHub CLI):**

```bash
gh release download -R codesphere-cloud/cs-go -O /usr/local/bin/cs -p "*linux_amd64"
chmod +x /usr/local/bin/cs
```

- **Example (via wget + jq):**

```bash
wget -qO- 'https://api.github.com/repos/codesphere-cloud/cs-go/releases/latest' \
  | jq -r '.assets[] | select(.name | match("linux_amd64")) | .browser_download_url' \
  | xargs wget -O cs
mv cs /usr/local/bin/cs && chmod +x /usr/local/bin/cs
```

Or download the platform binary directly from the GitHub Releases page.

### `cs-go` — Global Flags

| Flag                 | Env var           | Description                                                 |
| -------------------- | ----------------- | ----------------------------------------------------------- |
| `--api` / `-a`       | `CS_API`          | API URL, default `https://codesphere.com/api`               |
| `--team` / `-t`      | `CS_TEAM_ID`      | Team ID                                                     |
| `--workspace` / `-w` | `CS_WORKSPACE_ID` | Workspace ID                                                |
| _(token)_            | `CS_TOKEN`        | API token from Codesphere user settings — required for auth |

If team/workspace aren't passed as flags, the CLI reads them from the env vars; commands that need them fail without either.

### `cs-go` — Commands

| Command      | Description                                    |
| ------------ | ---------------------------------------------- |
| `cs create`  | Create a Codesphere resource (workspace, etc.) |
| `cs delete`  | Delete resources                               |
| `cs exec`    | Run a command in a workspace                   |
| `cs list`    | List resources (workspaces, teams)             |
| `cs log`     | Retrieve run logs from services                |
| `cs monitor` | Monitor a command and report health            |
| `cs open`    | Open the IDE in a browser                      |
| `cs set-env` | Set env vars on a workspace                    |
| `cs start`   | Start a pipeline stage (prepare/test/run)      |
| `cs update`  | Self-update the CLI                            |
| `cs version` | Print CLI version                              |

- **Example:**

```bash
export CS_TOKEN="your-api-token"
export CS_TEAM_ID="your-team-id"
export CS_WORKSPACE_ID="your-workspace-id"

cs start --stage prepare
cs start --stage test
cs start --stage run

cs exec --command "npm install"
cs set-env --key "DATABASE_URL" --value "postgres://..."
```

- **Note:** run `cs --help` for the live, authoritative command list — it's open source and can add commands between releases.

### What the CLI Does NOT Cover

| Area                                                                          | Use instead                                                                           |
| ----------------------------------------------------------------------------- | ------------------------------------------------------------------------------------- |
| Vault / secrets                                                               | `/vault/...` — see [secret-management.md](./secret-management.md)                     |
| Managed Services (create/update/pause/delete a DB, publish a custom provider) | `/managed-services/...` — see `../managed-services/README.md`                         |
| Domains                                                                       | `/domains/...`                                                                        |
| Teams & organizations                                                         | `/teams/...`, `/organizations/...`                                                    |
| Landscape lifecycle (explicit deploy/scale/teardown as a unit)                | `/workspaces/{id}/landscape/...`                                                      |
| Usage/metrics                                                                 | `/usage/...`                                                                          |
| Cluster/org admin (private cloud / enterprise)                                | `/clusters/...`                                                                       |
| SSH keys                                                                      | `/ssh/keys`                                                                           |
| Metadata lookups (datacenters, base images, plan→resource mapping)            | `/metadata/...` — useful for resolving a `plan:` integer id before writing a `ci.yml` |

### Public API — Base URL & Auth

|                      |                                    |
| -------------------- | ---------------------------------- |
| Cloud base URL       | `https://codesphere.com/api`       |
| Self-hosted base URL | `https://<your-base-url>/api`      |
| Auth                 | Bearer token, Settings > API Keys  |
| Interactive docs     | `https://<instance>/api/scalar-ui` |

- **Example:**

```bash
curl -H "Authorization: Bearer <token>" https://codesphere.com/api/workspaces/team/<teamId>
```

### Endpoint Catalog — `workspaces`

| Method             | Path                                                           | Purpose                                                             |
| ------------------ | -------------------------------------------------------------- | ------------------------------------------------------------------- |
| POST               | `/workspaces`                                                  | Create a workspace                                                  |
| GET                | `/workspaces/team/{teamId}`                                    | List workspaces for a team                                          |
| GET                | `/workspaces/{workspaceId}`                                    | Get workspace details                                               |
| PATCH              | `/workspaces/{workspaceId}`                                    | Update workspace                                                    |
| DELETE             | `/workspaces/{workspaceId}`                                    | Delete workspace                                                    |
| GET / PUT / DELETE | `/workspaces/{workspaceId}/env-vars`                           | List / set / delete plain env vars                                  |
| POST               | `/workspaces/{workspaceId}/execute`                            | Run a command in the workspace                                      |
| GET                | `/workspaces/{workspaceId}/git/head`                           | Current git HEAD                                                    |
| POST               | `/workspaces/{workspaceId}/git/pull[/{remote}[/{branch}]]`     | Git pull                                                            |
| POST               | `/workspaces/{workspaceId}/landscape/deploy[/{profile}]`       | Deploy the Landscape (optionally a specific CI profile)             |
| PATCH              | `/workspaces/{workspaceId}/landscape/scale`                    | Scale named services: body `{"serviceName": replicaCount, ...}`     |
| DELETE             | `/workspaces/{workspaceId}/landscape/teardown`                 | Tear down the Landscape (and its Landscape-scoped Managed Services) |
| GET                | `/workspaces/{workspaceId}/logs/{stage}/{step}`                | Stage/step logs                                                     |
| GET                | `/workspaces/{workspaceId}/logs/run/{step}/replica/{replica}`  | Per-replica run logs                                                |
| GET                | `/workspaces/{workspaceId}/logs/run/{step}/server/{server}`    | Per-server (landscape service) run logs                             |
| GET                | `/workspaces/{workspaceId}/pipeline/{stage}`                   | Stage status                                                        |
| POST               | `/workspaces/{workspaceId}/pipeline/{stage}/start[/{profile}]` | Start a stage                                                       |
| POST               | `/workspaces/{workspaceId}/pipeline/{stage}/stop`              | Stop a stage                                                        |
| GET                | `/workspaces/{workspaceId}/status`                             | Overall workspace status                                            |

### Endpoint Catalog — `managed-services`

| Method               | Path                                                 | Purpose                                                                                    |
| -------------------- | ---------------------------------------------------- | ------------------------------------------------------------------------------------------ |
| POST                 | `/managed-services`                                  | Create a standalone managed service                                                        |
| GET                  | `/managed-services`                                  | List managed services                                                                      |
| GET                  | `/managed-services/providers`                        | List available providers — **always check this before hardcoding a provider name/version** |
| POST / PUT           | `/managed-services/providers`                        | Create / upsert a custom provider                                                          |
| PATCH / DELETE       | `/managed-services/providers/{name}/{schemaVersion}` | Update / delete a custom provider                                                          |
| GET / PATCH / DELETE | `/managed-services/{id}`                             | Get details / update / delete a service                                                    |
| POST                 | `/managed-services/{id}/backups`                     | Trigger a manual backup                                                                    |

### Endpoint Catalog — `vault`

| Method              | Path                                                      | Purpose                                                |
| ------------------- | --------------------------------------------------------- | ------------------------------------------------------ |
| GET / POST / DELETE | `/vault/teams/{teamId}/workspaces/{workspaceId}`          | List keys (`/keys`) / store / delete workspace secrets |
| POST                | `/vault/teams/{teamId}/workspaces/{workspaceId}/generate` | Server-side generate + store secrets                   |
| GET / POST / DELETE | `/vault/teams/{teamId}/shared`                            | List / create / delete shared vaults                   |
| GET / POST / DELETE | `/vault/teams/{teamId}/shared/{vaultName}/...`            | Keys / store / delete secrets in a shared vault        |

Full detail: [secret-management.md](./secret-management.md).

### Endpoint Catalog — `domains`

| Method                      | Path                                                               | Purpose                                                                                                                              |
| --------------------------- | ------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------ |
| GET                         | `/domains/team/{teamId}`                                           | List domains                                                                                                                         |
| GET / POST / PATCH / DELETE | `/domains/team/{teamId}/domain/{domainName}`                       | Get / create / update / delete a domain                                                                                              |
| POST                        | `/domains/team/{teamId}/domain/{domainName}/verify`                | Trigger DNS verification                                                                                                             |
| PUT                         | `/domains/team/{teamId}/domain/{domainName}/workspace-connections` | Set which workspace IDs are connected to this domain. Body: map of route key → array of workspace IDs (for A/B / blue-green routing) |

### Endpoint Catalog — `teams`

| Method       | Path                                    | Purpose                                                                                 |
| ------------ | --------------------------------------- | --------------------------------------------------------------------------------------- |
| GET / POST   | `/teams`                                | List / create teams. Create body: `{name, dc, organizationId?}` (`dc` = datacenter id)  |
| GET / DELETE | `/teams/{teamId}`                       | Get / delete a team                                                                     |
| GET / POST   | `/teams/{teamId}/members`               | List / invite members. Invite body: `{userEmail, role}` (`role`: `0`=Admin, `1`=Member) |
| DELETE       | `/teams/{teamId}/members/{userId}`      | Remove a member                                                                         |
| PUT          | `/teams/{teamId}/members/{userId}/role` | Change a member's role                                                                  |
| POST         | `/teams/{teamId}/migrate`               | Migrate a team into an organization                                                     |

### Endpoint Catalog — `organizations`

| Method | Path                                                    | Purpose             |
| ------ | ------------------------------------------------------- | ------------------- |
| GET    | `/organizations`                                        | List organizations  |
| GET    | `/organizations/{organizationId}/members`               | List members        |
| POST   | `/organizations/{organizationId}/members`               | Add a member        |
| DELETE | `/organizations/{organizationId}/members/{userId}`      | Remove a member     |
| PUT    | `/organizations/{organizationId}/members/{userId}/role` | Change role         |
| GET    | `/organizations/{organizationId}/teams`                 | List an org's teams |

### Endpoint Catalog — `clusters` (private cloud / cluster admin)

| Method     | Path                      | Purpose                                          |
| ---------- | ------------------------- | ------------------------------------------------ |
| POST       | `/clusters/admins`        | Add a cluster admin                              |
| GET / POST | `/clusters/organizations` | List / create organizations at the cluster level |

### Endpoint Catalog — `ssh`

| Method              | Path        | Purpose                            |
| ------------------- | ----------- | ---------------------------------- |
| GET / POST / DELETE | `/ssh/keys` | List / upload / delete public keys |

### Endpoint Catalog — `metadata`

| Method | Path                              | Purpose                                                                            |
| ------ | --------------------------------- | ---------------------------------------------------------------------------------- |
| GET    | `/metadata/datacenters`           | Available datacenters                                                              |
| GET    | `/metadata/workspace-base-images` | Available Reactive base images                                                     |
| GET    | `/metadata/workspace-plans`       | Plan id → resource mapping — resolve a `ci.yml` `plan:` integer before writing one |

### Endpoint Catalog — `usage`

| Method | Path                                                                    | Purpose                            |
| ------ | ----------------------------------------------------------------------- | ---------------------------------- |
| GET    | `/usage/teams/{teamId}/resources/landscape-service/summary`             | Team-wide landscape resource usage |
| GET    | `/usage/teams/{teamId}/resources/landscape-service/{resourceId}/events` | Usage events for one service       |

### `POST /workspaces` — Create Workspace

- **Parameters:**

| Name                | Type                     | Required | Description                                        |
| ------------------- | ------------------------ | -------- | -------------------------------------------------- |
| `teamId`            | integer                  | Yes      | Owning team.                                       |
| `name`              | string                   | Yes      | Workspace name.                                    |
| `planId`            | integer                  | Yes      | Resource plan id.                                  |
| `isPrivateRepo`     | boolean                  | Yes      | Whether the source repo is private.                |
| `replicas`          | integer                  | Yes      | Initial replica count.                             |
| `baseImage`         | string                   | No       | Reactive base image.                               |
| `gitUrl`            | string                   | No       | Source repository URL.                             |
| `initialBranch`     | string                   | No       | Branch to check out.                               |
| `cloneDepth`        | integer                  | No       | Git clone depth.                                   |
| `sourceWorkspaceId` | integer                  | No       | Clone an existing workspace instead of a git repo. |
| `vpnConfig`         | string                   | No       | VPN config name.                                   |
| `restricted`        | boolean                  | No       | **Defaults to `true` (confirmed live) when omitted.** A `restricted: true` workspace's dev domain 303-redirects every request to the IDE sign-in page instead of serving the app — even for a service that's `isPublic: true` with a working `network.paths` route, and even while `GET /workspaces/{id}/status` and the pipeline both report fully healthy. Pass `"restricted": false` explicitly (CLI: `cs create workspace --public-dev-domain`) for a workspace whose app is meant to be publicly reachable. Can also be changed after creation via `PATCH /workspaces/{workspaceId}` with `{"restricted": false}`. |
| `storageMib`        | integer                  | No       | Storage size override.                             |
| `sharedVaultName`   | string                   | No       | Assign a shared vault at creation.                 |
| `env`               | array of `{name, value}` | No       | Initial plain env vars.                            |
| `welcomeMessage`    | string                   | No       | Custom IDE welcome message.                        |

- **Example:**

```bash
curl -X POST https://codesphere.com/api/workspaces \
  -H "Authorization: Bearer <token>" -H "Content-Type: application/json" \
  -d '{
    "teamId": 123, "name": "my-workspace", "planId": 8,
    "isPrivateRepo": false, "replicas": 1,
    "gitUrl": "https://github.com/user/repo.git",
    "restricted": false
  }'
```

- **Fixing a workspace created without `restricted: false`** (dev domain redirects to IDE sign-in instead of serving the app):

```bash
curl -X PATCH https://codesphere.com/api/workspaces/456 \
  -H "Authorization: Bearer <token>" -H "Content-Type: application/json" \
  -d '{"restricted": false}'
```

### Deploy / Scale / Teardown a Landscape

- **Example:**

```bash
curl -X POST "https://codesphere.com/api/workspaces/456/landscape/deploy/production" \
  -H "Authorization: Bearer <token>"

curl -X PATCH "https://codesphere.com/api/workspaces/456/landscape/scale" \
  -H "Authorization: Bearer <token>" -H "Content-Type: application/json" \
  -d '{"myServer1": 3, "myServer2": 2}'

curl -X DELETE "https://codesphere.com/api/workspaces/456/landscape/teardown" \
  -H "Authorization: Bearer <token>"
```

### Set Env Vars

- **Example:**

```bash
curl -X PUT "https://codesphere.com/api/workspaces/456/env-vars" \
  -H "Authorization: Bearer <token>" -H "Content-Type: application/json" \
  -d '[{"name": "DATABASE_URL", "value": "postgres://..."}]'
```

### GitHub Actions Integration

- **Description:** `codesphere-cloud/gh-action-deploy@main` — creates a preview Landscape per PR, tears it down when the PR closes.
- **Parameters:**

| Name        | Required | Default        | Description               |
| ----------- | -------- | -------------- | ------------------------- |
| `email`     | Yes      | —              | Codesphere account email  |
| `password`  | Yes      | —              | Account password          |
| `team`      | Yes      | —              | Team name                 |
| `plan`      | No       | Boost          | Micro, Boost, or Pro      |
| `onDemand`  | No       | false          | Standby when unused       |
| `env`       | No       | —              | Dotenv-formatted env vars |
| `vpnConfig` | No       | —              | VPN config name           |
| `apiUrl`    | No       | codesphere.com | Custom instance URL       |

- **Required secrets:** `CODESPHERE_EMAIL`, `CODESPHERE_PASSWORD`.
- **Supported PR triggers:** `opened`, `reopened`, `synchronize`, `closed`.
- **Example:**

```yaml
name: Deploy to Codesphere
on:
  pull_request:
    types: [opened, reopened, synchronize, closed]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: codesphere-cloud/gh-action-deploy@main
        with:
          email: ${{ secrets.CODESPHERE_EMAIL }}
          password: ${{ secrets.CODESPHERE_PASSWORD }}
          team: "MyTeam"
          plan: "Boost"
          onDemand: true
          env: |
            API_KEY=${{ secrets.API_KEY }}
            DATABASE_URL=${{ secrets.DATABASE_URL }}
```

### GitLab CI Integration

- **Example:**

```yaml
deploy:
  stage: deploy
  script:
    - |
      curl -X POST "https://codesphere.com/api/workspaces/${CS_WORKSPACE_ID}/git/pull/origin" \
        -H "Authorization: Bearer ${CS_TOKEN}"
    - |
      curl -X POST "https://codesphere.com/api/workspaces/${CS_WORKSPACE_ID}/pipeline/prepare/start" \
        -H "Authorization: Bearer ${CS_TOKEN}"
  variables:
    CS_TOKEN: $CODESPHERE_API_TOKEN
    CS_WORKSPACE_ID: $WORKSPACE_ID
```

### Bitbucket Pipelines Integration

- **Example:**

```yaml
pipelines:
  default:
    - step:
        name: Deploy to Codesphere
        script:
          - |
            curl -X POST "https://codesphere.com/api/workspaces/${CS_WORKSPACE_ID}/git/pull/origin" \
              -H "Authorization: Bearer ${CS_TOKEN}"
          - |
            curl -X POST "https://codesphere.com/api/workspaces/${CS_WORKSPACE_ID}/pipeline/prepare/start" \
              -H "Authorization: Bearer ${CS_TOKEN}"
```

## Common Pitfalls

- Looking for a `cs secret`/`cs managed-service`/`cs domain` CLI subcommand — none exist; go straight to the Public API for these areas.
- Confusing `cs start --stage run` (single pipeline stage on the current workspace) with the Landscape lifecycle endpoints (`/workspaces/{id}/landscape/deploy|scale|teardown`) — they operate at different levels.
- Hardcoding a `plan:` integer without checking `GET /metadata/workspace-plans` first.
- Assuming the CLI's command list in this file is exhaustive — it's open source and can add commands between releases; run `cs --help` for the live list.
- **Confirmed live:** `nix`/`nix-env` is not on `PATH` for a command run via `POST /workspaces/{workspaceId}/execute` (and therefore `cs exec`) — that endpoint invokes a plain non-login shell, which never sources `/home/user/.nix-profile/etc/profile.d/nix.sh` the way an interactive IDE terminal does. A `nix-env -iA nixpkgs.<pkg>` command that works fine as a `ci.yml` `prepare`/`run` step (confirmed to have Nix on `PATH` there) fails with "command not found" when run ad hoc via `cs exec`/`/execute` unless prefixed with `source /home/user/.nix-profile/etc/profile.d/nix.sh &&`.

## Known Documentation Discrepancies

> This section may be outdated. Please verify against the official documentation:
> <https://docs.codesphere.com>

- This file summarizes the endpoint catalog from a bundled OpenAPI spec (`codesphere-public-api.yaml`, `info.version: 0.1.0`) plus docs — the live Scalar UI (`https://<instance>/api/scalar-ui`) is the source of truth for exact request/response schemas, since the API evolves faster than this reference.
- `cs-go` is open source and can add commands between releases — the command table above is a snapshot, not guaranteed current.

## Further Reading

- Official docs: <https://docs.codesphere.com>
- Interactive API schema: `https://<instance>/api/scalar-ui`
- `cs-go` repository: <https://github.com/codesphere-cloud/cs-go>
- Vault/secrets detail: [secret-management.md](./secret-management.md)
- Managed services detail: `../managed-services/README.md`
