# Codesphere CI Pipeline (`ci.yml`) Reference

> **Last updated:** 2026-07-28 · **Source:** https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/configuring-ci-pipeline

> The documentation changes regularly. If you are facing issues, please refer to the official documentation:
> <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/configuring-ci-pipeline>

## Overview

`ci.yml` (at the project root, `/home/user/app/`) declares a workspace's pipeline: `prepare` → `test` → `run`, with `run` as a map of named services. Use this reference for the exact field shape of each service type (Reactive, Managed Container, Managed Service) and for schema-version differences. Language/framework-specific `prepare`/`run` recipes live in [runtimes.md](./runtimes.md); Landscape-level concepts (networking, multi-service patterns) live in [landscape.md](./landscape.md).

## Core Concepts

- **File location**: `ci.yml` (default) or `ci.<profilename>.yml` (named profile, e.g. `ci.production.yml`), always at the project root.
- **`schemaVersion`**: required, must be the first meaningful line.
- **Reactive**: a service with `steps:` and no `image:` — runs on Codesphere's pooled, pre-warmed base image, millisecond cold starts.
- **Managed Container**: a service with `image:` — you bring your own Docker image; same orchestration (networking/scaling/monitoring) as a Reactive, but second-scale cold starts. Mutually exclusive with `steps`.
- **Managed Service**: a service with a `provider:` block — a database/queue/etc. from the managed-services catalog, embedded directly in the Landscape. See the managed-services reference set for the full architecture.
- **Stage semantics**: `prepare`/`test` run once, on the main replica only, no auto-restart; `run` executes on every replica and auto-restarts on crash.

## API / Syntax

### `schemaVersion` — Version Differences

| Version | Notes                                                                                                                                                              |
| ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `v0.1`  | Legacy. No `schemaVersion` field; `run` is a single flat stage (`run.steps[]`), not a map of named services. Single-app only — no Landscape/multi-service support. |
| `v0.2`  | Introduces `run.<serviceName>: {steps, plan, replicas, network, ...}`. Widely documented, but superseded — see note below.                                         |
| `v0.3`  | Same shapes as v0.2 but makes `volumeMounts` explicit/required for precise filesystem scoping.                                                                     |
| `v0.4`  | **Current.** Renames `provider.version` → `provider.schemaVersion` for Managed Service blocks. Everything else matches v0.3.                                       |

- **`v0.4` is the current schema version** (confirmed directly with Codesphere) — target it by default for new `ci.yml` files, including the `provider.schemaVersion` field name (not `provider.version`) on Managed Service blocks. **The public docs' framing of `v0.2` as the safe/documented default is outdated** — don't rely on that framing even where it still appears verbatim in scraped source material. If a specific target instance is confirmed to only parse an older version, fall back accordingly, but default to v0.4 otherwise.

### Stage Behavior

| Stage     | Runs on           | Restarts on crash | Shared across services                    |
| --------- | ----------------- | ----------------- | ----------------------------------------- |
| `prepare` | Main replica only | No                | Yes — one prepare for the whole Landscape |
| `test`    | Main replica only | No                | Yes — one test for the whole Landscape    |
| `run`     | All replicas      | Yes, automatic    | No — each named service has its own `run` |

### Step Shape

- **Parameters:**

| Name      | Type   | Required | Description                               |
| --------- | ------ | -------- | ----------------------------------------- |
| `name`    | string | No       | Human-readable description (recommended). |
| `command` | string | Yes      | Any shell command.                        |

- **Example:**

```yaml
steps:
  - name: "Human-readable description"
    command: "shell command to execute"
```

### Reactive Service — Field Reference

- **Description:** A service is a Reactive when it has `steps:` and no `image:`.
- **Parameters:**

| Name                                            | Type                       | Required | Description                                                                                                  |
| ----------------------------------------------- | -------------------------- | -------- | ------------------------------------------------------------------------------------------------------------ |
| `steps`                                         | array                      | Yes      | Sequential startup commands; the last one should be long-running.                                            |
| `image`                                         | string                     | No       | Custom base image (advanced) — becomes the basis for a Managed Container-like setup.                         |
| `healthEndpoint`                                | string                     | No       | Default `http://localhost:3000/`, checked from inside the container.                                         |
| `plan`                                          | integer                    | No       | Resource tier id — resolve via `GET /metadata/workspace-plans` or the IDE's plan picker.                     |
| `replicas`                                      | integer                    | No       | Horizontal scaling, default `1`, max `10` (more on enterprise).                                              |
| `isPublic`                                      | boolean                    | No       | Shorthand; ignored if the advanced `network` block is present.                                               |
| `network.ports[].port` / `.isPublic`            | integer / boolean          | No       | Raw exposed ports; `isPublic: true` also gets a direct public port URL (rarely needed, prefer path routing). |
| `network.paths[].port` / `.path` / `.stripPath` | integer / string / boolean | No       | Workspace Router: map URL path prefixes to ports. `stripPath: true` removes the prefix before forwarding.    |
| `env`                                           | map                        | No       | Key-value env vars, injected at runtime.                                                                     |
| `runAsUser` / `runAsGroup`                      | integer                    | No       | Optional UID/GID.                                                                                            |
| `volumeMounts[].name`                           | string                     | No       | Currently only `_workspace` is supported.                                                                    |
| `volumeMounts[].mountPath`                      | string                     | No       | Destination path inside the runtime.                                                                         |
| `volumeMounts[].workspacePath`                  | string                     | No       | Subdirectory of `/home/user/app` to mount; `""` = whole workspace.                                           |

- **Example:**

```yaml
run:
  <serviceName>:
    steps:
      - name: <string>
        command: <string>
    image: <string>
    healthEndpoint: <string>
    plan: <integer>
    replicas: <integer>
    isPublic: <boolean>
    network:
      ports:
        - port: <integer>
          isPublic: <boolean>
      paths:
        - port: <integer>
          path: <string>
          stripPath: <boolean>
    env:
      KEY: <string or number>
    runAsUser: <integer>
    runAsGroup: <integer>
    volumeMounts:
      - name: _workspace
        mountPath: <string>
        workspacePath: <string>
```

- **Note:** every service needs at least one route (`network.paths` entry, or `isPublic: true`) or the Workspace Router never considers it healthy/reachable.

### Managed Container — Field Reference

- **Description:** A service is a Managed Container when it has `image:`. Same platform orchestration as a Reactive; you bring your own Docker image instead of building on Codesphere's base image. Startup is seconds (image pull + init), not milliseconds.
- **Parameters:**

| Name             | Type             | Required | Description                                                                                                                                            |
| ---------------- | ---------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `image`          | string           | Yes      | e.g. `nginx:1.25-alpine`, `myregistry.com/repo/image:tag`.                                                                                             |
| `command`        | array of strings | No       | Overrides the image's default CMD; entrypoint stays the image's own. Mutually exclusive with `steps`.                                                  |
| `healthEndpoint` | string           | No       | Same as Reactive.                                                                                                                                      |
| `plan`           | integer          | No       | Same as Reactive.                                                                                                                                      |
| `replicas`       | integer          | No       | Same as Reactive.                                                                                                                                      |
| `network`        | object           | No       | Identical shape to Reactives; also supports a simplified single-path form (`network.path` + `network.stripPath`).                                      |
| `env`            | map              | No       | Same as Reactive.                                                                                                                                      |
| `runAsUser`      | integer          | No       | UID `1501` gets the same read/write access to workspace files as a Reactive; any other UID gets read access everywhere + write only at the mount root. |
| `runAsGroup`     | integer          | No       | GID `1010` is always added for workspace FS group ownership.                                                                                           |
| `volumeMounts`   | array            | No       | Same shape as Reactive; `mountPath` can point anywhere in the container, e.g. mount a single config file.                                              |

- **Example:**

```yaml
run:
  <serviceName>:
    image: <string>
    command: <array of strings>
    healthEndpoint: <string>
    plan: <integer>
    replicas: <integer>
    network:
      path: /container
      stripPath: false
    env: { ... }
    runAsUser: <integer>
    runAsGroup: <integer>
    volumeMounts:
      - name: _workspace
        mountPath: <string>
        workspacePath: <string>
```

- **Note:** to override the image's **entrypoint** (not just its command) — not possible for a Managed Container. Use a Reactive with the image set via `image:` as a custom base image instead, driven by `steps`.

### Managed Service (embedded) — Field Reference

- **Description:** A service is a Managed Service when it has a `provider:` block instead of `steps`/`image`. Full architecture: see the managed-services reference set.
- **Parameters:**

| Name                                                      | Type                        | Required                | Description                                                                             |
| --------------------------------------------------------- | --------------------------- | ----------------------- | --------------------------------------------------------------------------------------- |
| `provider.name`                                           | string                      | Yes                     | e.g. `postgres`, `valkey` — verify current names via `GET /managed-services/providers`. |
| `provider.schemaVersion`                                  | string                      | Yes (v0.4, current)     | Field name for the provider's schema version on `v0.4` — **use this by default**.       |
| `provider.version`                                        | string                      | Legacy (v0.2/v0.3 only) | Older field name — only needed if targeting a pre-v0.4 instance.                        |
| `plan.id` / `plan.parameters`                             | integer / object            | Yes                     | Plan selection + parameters (`storage` MB, `cpu`, `memory` MB).                         |
| `config`                                                  | object                      | No                      | Non-secret provider config; schema is provider-specific.                                |
| `secrets`                                                 | object                      | Provider-dependent      | Provider-specific secrets; values may use `${{ vault.NAME }}`.                          |
| `backups.enabled` / `.intervalH` / `.deleteRetentionDays` | boolean / integer / integer | No                      | Only for providers that support backups (e.g. PostgreSQL).                              |
| `backups.config.endpointUrl` / `.destinationPath`         | string                      | No                      | S3-compatible backup target.                                                            |
| `backups.secrets.accessKey` / `.secretKey`                | string                      | No                      | S3 credentials.                                                                         |

- **Example (current, v0.4):**

```yaml
run:
  <serviceName>:
    provider:
      name: <string>
      schemaVersion: <string>
    plan:
      id: <integer>
      parameters:
        storage: <integer>
        cpu: <integer>
        memory: <integer>
    config:
      <key>: <value>
    secrets:
      <key>: <string>
    backups:
      enabled: <boolean>
      intervalH: <integer>
      deleteRetentionDays: <integer>
      config: { endpointUrl: <string>, destinationPath: <string> }
      secrets: { accessKey: <string>, secretKey: <string> }
```

- **Legacy (v0.2/v0.3 — only if the target instance doesn't yet support v0.4):** identical shape, but `provider.version` replaces `provider.schemaVersion`.
- **Warning:** renaming a Managed Service's key in `ci.yml` forces recreation — this can cause data loss. Treat the service name as fixed once deployed.

### Cloud Native Deployment (Virtual Kubernetes Cluster)

- **Description:** Not a `run.<serviceName>` runtime type in the same sense — it's a Managed Service (`virtual-k8s` provider) giving the team a real `kubectl`-accessible control plane. Drive it from `prepare`/`run` steps of any Reactive/Managed Container in the same Landscape. Each team can only have **one** virtual cluster at a time.
- **Example (Helm):**

```yaml
prepare:
  steps:
    - name: Deploy with Helm
      command: |
        helm repo add corp42 https://charts.corp42.net
        helm repo update
        helm install my-release corp42/my-awesome-app -n app --create-namespace -f values-codesphere.yaml
```

- **Example (raw manifests):**

```yaml
steps:
  - name: Deploy Application
    command: |
      kubectl apply -f k8s/deployment.yaml
      kubectl apply -f k8s/service.yaml
```

- **Note:** monitoring and Workspace Router networking integration are currently manual for workloads running inside the virtual cluster — automatic only for Reactives/Managed Containers.

### CI Profiles

- **Description:** Multiple pipeline configurations in one workspace.
- **Workflow:** Setup > CI in the IDE → Add Profile (creates `ci.<name>.yml`) → select it when running the pipeline, or reference it from a Service Provider's `backend.landscape.ciProfile`.
- **Typical layout:** `ci.yml` (dev/default), `ci.staging.yml`, `ci.production.yml`.

### Off-the-Shelf Image Example (Managed Container)

- **Example:**

```yaml
schemaVersion: v0.4
prepare:
  steps: []
test:
  steps: []
run:
  nginx-server:
    image: nginx-unprivileged:1.25-alpine
    command: ["-g", "daemon off;"]
    plan: 8
    healthEndpoint: http://localhost/
    network:
      ports:
        - port: 80
          isPublic: false
      paths:
        - port: 80
          path: /
          stripPath: false
    env:
      NGINX_HOST: example.com
    volumeMounts:
      - name: _workspace
        mountPath: /etc/nginx/nginx.conf
        workspacePath: custom-nginx.conf
    runAsUser: 1000
    runAsGroup: 1000
```

## Common Pitfalls

- Omitting `schemaVersion` as the first line, or leaving `run` as a flat `steps[]` list — both are v0.1 shapes and will fail to parse as v0.2+.
- Mixing `steps` and `image` on the same service — mutually exclusive (Reactive vs. Managed Container).
- Using `provider.version` by default — `provider.schemaVersion` is correct for the current (v0.4) schema; only fall back to `provider.version` for a confirmed pre-v0.4 instance.
- Renaming a service (Reactive, Managed Container, or Managed Service) after deploy — forces recreation, can mean data loss for stateful services.
- Assuming `plan:` integers map to a fixed CPU/RAM table — the mapping is cluster-specific; resolve via `GET /metadata/workspace-plans`.
- Forgetting a `network.paths` entry (or `isPublic: true`) — the service is never marked healthy/reachable by the Workspace Router.

## Known Documentation Discrepancies

> This section may be outdated. Please verify against the official documentation:
> <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/configuring-ci-pipeline>

- **Resolved (confirmed directly with Codesphere, not from public docs):** `v0.4` is the current schema version, not `v0.2`. The publicly scraped source docs for this page recommended defaulting to `v0.2` as the "safest" target — that framing is outdated. This reference now defaults to `v0.4`/`provider.schemaVersion` throughout; treat any surviving `v0.2`/`provider.version` framing found elsewhere (public docs, older scrapes, search results) as stale unless a specific target instance is confirmed not to support `v0.4`.
- The exact `plan:` integer → CPU/RAM/storage mapping is cluster-specific and intentionally not reproduced as a fixed table here — always resolve via `GET /metadata/workspace-plans`.

## Further Reading

- Official docs: <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/configuring-ci-pipeline>
- CI profiles: <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/using-ci-profiles>
- Language/framework runtime recipes: [runtimes.md](./runtimes.md)
- Landscape networking & multi-service patterns: [landscape.md](./landscape.md)
- Managed services architecture: `../managed-services/README.md`
