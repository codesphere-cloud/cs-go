# Codesphere Migration Guide Reference

> **Last updated:** 2026-07-28 · **Source:** https://docs.codesphere.com (migration guidance, cross-referenced from ci-and-deploy docs)

> The documentation changes regularly. If you are facing issues, please refer to the official documentation:
> <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/configuring-a-landscape>

## Overview

Field-by-field mapping guidance for moving an existing deployment (`docker-compose.yml`, a bare `Dockerfile`, a Helm chart, raw Kubernetes manifests, or another PaaS's config) onto Codesphere. Use this reference when a user asks "how do I move X to Codesphere" rather than starting a `ci.yml` from scratch.

## Core Concepts

- **No in-platform image builds**: Codesphere does not build Dockerfiles for you. Either build+push the image elsewhere and reference it with `image:` (Managed Container), or rewrite the Dockerfile's `RUN`/`CMD` as Reactive `prepare`/`run` steps.
- **DB/cache/queue services should become Managed Services**, not hand-rolled containers, whenever the current catalog has an equivalent — check `GET /managed-services/providers` before assuming a provider name (e.g. Redis-_protocol_ compatibility is via "Valkey", not a "redis" provider).
- **No native dependency graph across `run.<serviceName>` entries**: Compose's `depends_on:` has no direct equivalent — add a wait/retry loop in the dependent app's own `run` step if true startup ordering matters.
- **No named/external volumes**: only `/home/user/app` (network FS) persists and is shareable via `volumeMounts`. Anything that needs to be a real durable blob store goes to Object Storage (S3-compatible Managed Service) instead.
- **Prefer a Landscape (Reactive/Managed Container/Managed Service) over a Cloud Native Deployment for production Helm migrations, whenever the chart allows it.** A self-managed Virtual Kubernetes Cluster shifts real operational and compliance responsibility onto the customer that Codesphere's own Landscape primitives otherwise cover automatically — e.g. confidential-compute/isolation guarantees have to be built and proven by the customer themselves on a self-managed cluster, instead of being a platform guarantee. Reserve Cloud Native Deployment for the part of a chart that genuinely needs Kubernetes semantics (CRDs, operators, StatefulSets with no managed-service equivalent) after managed-service replacement — not as the default reflex for "it's a Helm chart."

## API / Syntax

### Decision Tree

```
What are you migrating?
│
├─ docker-compose.yml
│   └─ One Landscape, one Managed Container per compose service
│      (or Reactive, if you'd rather rebuild from source than ship the image)
│      DB/cache/queue services with a Codesphere-managed equivalent →
│      replace with a Managed Service instead of containerizing them yourself
│
├─ Dockerfile only (no compose)
│   └─ Single Managed Container — needs a pre-built, pushed image
│      (Codesphere doesn't build Dockerfiles for you), OR
│      reimplement the Dockerfile's RUN/CMD as Reactive prepare/run steps
│
├─ Helm chart → prefer decomposing into a Codesphere Landscape; Cloud Native
│   │  Deployment (Virtual Kubernetes Cluster) is a supported fallback, not
│   │  the production default — see "Helm Chart → Codesphere Landscape" below
│   ├─ Non-trivial remainder (multiple resources, CRDs, operators,
│   │   StatefulSets not covered by a managed-service replacement) →
│   │   Cloud Native Deployment, `helm install` in prepare/run steps,
│   │   chart stays as-is — only when a Landscape genuinely can't fit
│   └─ Trivial, or made trivial once DB/cache/queue components become
│       Managed Services → decompose into a Landscape: per remaining
│       component, Reactive (own Dockerfile/code) or Managed Container
│       (unmodified vendor image), never both from a single yes/no —
│       see "Per-Component: Reactive vs. Managed Container vs. Ask" below
│
├─ Raw Kubernetes manifests (no Helm)
│   └─ Same as Helm: Cloud Native Deployment, `kubectl apply -f` in steps
│
└─ Another PaaS's config (Heroku Procfile, Railway, Render, Fly.io, etc.)
    └─ Usually maps cleanly to a single Reactive: the start command becomes
       the last `run` step; buildpacks become `prepare` steps (npm/pip/etc.)
```

### `docker-compose.yml` → Landscape `ci.yml` — Field Mapping

| Compose field                                         | Codesphere equivalent                                                                                                     | Notes                                                                                                                                                                                                   |
| ----------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| top-level `services.<name>`                           | `run.<serviceName>`                                                                                                       | Landscape service names have the same naming rules as YAML keys — lowercase-hyphenated is safest                                                                                                        |
| `image:`                                              | `image:` (Managed Container)                                                                                              | Direct match                                                                                                                                                                                            |
| `build:`                                              | _(no direct equivalent)_                                                                                                  | Either build+push the image elsewhere and reference with `image:`, or rewrite as a Reactive with `prepare`/`run` steps replicating the Dockerfile's `RUN`/`CMD`                                         |
| `command:`                                            | `command:` (Managed Container, array form)                                                                                | Direct match                                                                                                                                                                                            |
| `environment:` (list or map)                          | `env:` (map)                                                                                                              | Direct match, normalize list `KEY=value` entries into map form                                                                                                                                          |
| `env_file:`                                           | Inline the values into `env:`, or use `${{ workspace.env['KEY'] }}` / `${{ vault.KEY }}` templates                        | No native `env_file` support in `ci.yml`                                                                                                                                                                |
| `ports: ["8080:80"]`                                  | `network.ports` (raw port) + `network.paths` (routed path)                                                                | Compose's _host_ port is irrelevant on Codesphere — only the _container_ port matters, routed via the Workspace Router. Give each exposed service a `paths` entry.                                      |
| `depends_on:`                                         | _(implicit)_                                                                                                              | All services share one private network already, and `prepare`/`test` already run before any `run` stage starts. For true startup-order dependencies, add a wait/retry loop in the app's own `run` step. |
| `volumes:` (bind mount `./x:/y`)                      | `volumeMounts` scoped under `/home/user/app`                                                                              | Map the bind source into a subdirectory of the repo and mount via `volumeMounts.workspacePath`.                                                                                                         |
| `volumes:` (named volume)                             | _(rework required)_                                                                                                       | Move the data into `/home/user/app`, or use Object Storage for anything needing a real durable blob store.                                                                                              |
| `networks:`                                           | _(ignored — not needed)_                                                                                                  | Every service in a Landscape is already reachable from every other via internal DNS.                                                                                                                    |
| `restart: always` / `unless-stopped`                  | _(automatic)_                                                                                                             | The `run` stage always auto-restarts on crash — no config needed.                                                                                                                                       |
| `healthcheck:` (HTTP)                                 | `healthEndpoint:`                                                                                                         | Direct match if it's an HTTP GET — only HTTP checks are supported, not arbitrary shell `test: ["CMD", ...]`.                                                                                            |
| `healthcheck:` (shell command)                        | _(no direct equivalent)_                                                                                                  | Expose a lightweight HTTP `/health` endpoint from the app instead.                                                                                                                                      |
| `deploy.replicas:`                                    | `replicas:`                                                                                                               | Direct match                                                                                                                                                                                            |
| `deploy.resources.limits:`                            | `plan:` (integer tier id)                                                                                                 | Not a direct CPU/mem number — pick the closest `plan` id via `GET /metadata/workspace-plans`.                                                                                                           |
| Service referencing another by name (`postgres:5432`) | Internal DNS: `ws-server-[WorkspaceId]-[serviceName].workspaces:[port]`, or the managed service's own connection env vars | Replace bare compose service-name hostnames accordingly.                                                                                                                                                |
| Compose's own `postgres`/`redis`/`mongo`/etc. service | A **Managed Service**                                                                                                     | Check `GET /managed-services/providers` for the live catalog before assuming a name.                                                                                                                    |

### Worked Example — `docker-compose.yml` Input

- **Example:**

```yaml
services:
  web:
    build: .
    ports: ["8080:8080"]
    environment:
      - NODE_ENV=production
      - DATABASE_URL=postgres://app:pw@db:5432/mydb
    depends_on: [db, cache]
  worker:
    build: .
    command: ["node", "worker.js"]
    environment:
      - REDIS_URL=redis://cache:6379
    depends_on: [cache]
  db:
    image: postgres:16
    environment:
      - POSTGRES_PASSWORD=pw
      - POSTGRES_DB=mydb
    volumes: ["pgdata:/var/lib/postgresql/data"]
  cache:
    image: redis:7
volumes:
  pgdata:
```

### Worked Example — Resulting `ci.yml`

- **Description:** `web`/`worker` share one build (a Node app) and are rebuilt from source instead of shipping the image; `db` and `cache` become Managed Services.
- **Example:**

```yaml
schemaVersion: v0.4
prepare:
  steps:
    - name: Install dependencies
      command: npm install
test:
  steps: []
run:
  web:
    steps:
      - name: Start web
        command: npm run start:web
    plan: 8
    replicas: 1
    env:
      NODE_ENV: production
      DATABASE_URL: "postgresql://app:${{ vault.pgPassword }}@ws-server-${{ workspace.id }}-db.workspaces:5432/mydb"
    network:
      ports:
        - port: 8080
          isPublic: true
      paths:
        - port: 8080
          path: /
          stripPath: false
  worker:
    steps:
      - name: Start worker
        command: node worker.js
    plan: 8
    replicas: 1
    env:
      REDIS_URL: "redis://ws-server-${{ workspace.id }}-cache.workspaces:6379"
    network:
      ports: []
      paths: []
  db:
    provider:
      name: postgres # verify against GET /managed-services/providers
      schemaVersion: v1
    plan:
      id: 0
      parameters: { storage: 10000, cpu: 5, memory: 500 }
    config:
      version: "16"
      userName: app
      databaseName: mydb
    secrets:
      userPassword: "${{ vault.pgPassword }}"
      superuserPassword: "${{ vault.pgSuperuserPassword }}"
  cache:
    provider:
      name: valkey # Redis-protocol compatible; verify current name
      schemaVersion: v1
    plan:
      id: 0
      parameters: { storage: 1000, cpu: 2, memory: 250 }
    config: {}
    secrets: {}
```

- **Note:** confirm with the user whether `worker` truly needs no exposed port (`network.paths: []` is valid — it just means "not reachable from outside the Landscape, but still reachable internally"), and fetch the exact Managed Service `config`/`secrets` schema for the chosen provider versions via the API rather than guessing field names.

### Dockerfile-Only Repo (No Compose) — Two Paths

**A. Ship the built image** (fastest, closest to the existing setup):

```yaml
run:
  app:
    image: registry.example.com/myorg/myapp:latest # built+pushed by existing CI
    plan: 8
    network:
      ports: [{ port: 8080, isPublic: true }]
      paths: [{ port: 8080, path: /, stripPath: false }]
```

Requires the existing CI (GitHub Actions, GitLab CI, etc.) to keep building/pushing the image; Codesphere just pulls and runs it. For private registries, ensure credentials are configured for the image pull.

**B. Rebuild from source as a Reactive** (better platform integration — faster cold starts, no image registry to manage): read the Dockerfile's `RUN` lines → `prepare.steps`; its `CMD`/`ENTRYPOINT` → the last `run.<service>.steps` entry; its `EXPOSE`/`ENV` → `network`/`env`.

### Helm Chart → Codesphere Landscape (recommended default)

Decompose the chart into one `ci.yml` that mixes Reactive, Managed Container, and Managed Service entries under a single `run:` — a Landscape isn't restricted to one service type, so there is no need to force a whole chart into "all Reactive" or "all Managed Container." Steps, in order:

1. **Managed Services first.** Every component that's actually a database/cache/queue with a Codesphere-managed equivalent (Postgres, Redis/Valkey, RabbitMQ, OpenSearch, S3, DocumentDB, Babelfish/SQL Server) becomes a Managed Service, not a hand-rolled container — check `GET /managed-services/providers` before assuming a name. This also shrinks what's left to classify next; a StatefulSet running Postgres is not "Kubernetes complexity" once it's replaced, it's gone.
2. **Non-trivial remainder check.** If what's left still needs real Kubernetes semantics (CRDs, operators, a StatefulSet with no managed-service equivalent), that specific part is the one legitimate case for Cloud Native Deployment — see the fallback section below. Otherwise continue.
3. **Per-component: Reactive vs. Managed Container vs. Ask.** For every remaining component (the chart's own Deployments, not what step 1 already replaced), decide individually — never as one chart-wide choice:

   | Signal | Classification |
   | --- | --- |
   | No local Dockerfile/build context for this component anywhere in the repo, and the chart/`values.yaml` references a recognizable vendor/pre-built image tag (e.g. `grafana/grafana:10.4.2`, a `bitnami/*` image, a cert-manager helper) | **Managed Container** (`image:` = that exact image reference). Nothing to rebuild — the image is used unmodified, so pulling it as-is is both correct and less work than reimplementing someone else's app as Reactive `prepare`/`run` steps. |
   | A local Dockerfile/build context exists for this component in the repo, and it adds the team's own code/program on top of a base image | **Reactive**, rebuilt from source — the base image's genericness is irrelevant (`FROM ubuntu` + custom `COPY`/`RUN` is exactly as much "Reactive" as `FROM node:20-alpine` + custom code). Map the Dockerfile's `RUN` → `prepare.steps`, `CMD`/`ENTRYPOINT` → the last `run.<service>.steps` entry, `EXPOSE`/`ENV` → `network`/`env`, per [runtimes.md](./runtimes.md). A pinned `FROM` version is mandatory to reproduce via Nix, not optional. |
   | Neither signal is clear — an image reference with no local Dockerfile *and* no recognizable vendor name (e.g. a private-registry image some other pipeline builds) | **Ask the user directly**: "is `<image>` a vendor/off-the-shelf image, or does your team build it (just not in this repo)?" Don't guess either way — same pattern already used in `codesphere-create-container-deployment`'s Step 3 for an unconfirmed image source. |

   Present the per-component classification to the user as one batch (component → proposed type → one-line reason), same house style as the managed-service matches in step 1 — the user confirms or overrides individual entries, not the whole batch at once.
4. **Generate one `ci.yml`.** `schemaVersion: v0.4`. Reactive components use the matching recipe from [runtimes.md](./runtimes.md); Managed Container components use `image:` directly from the chart with `network`/`env` translated per the field mapping table above; Managed Service components follow the matching `references/provider-*.md`. Every service still needs at least one `network.paths` entry or `isPublic: true`, and `stripPath` must match whether that specific component's own routes already include the path prefix — check the actual route definitions, don't default to either value (see [landscape.md](./landscape.md)).

#### Worked Example — Mixed Landscape from One Chart

A chart with a custom `frontend` and `backend` (each with their own Dockerfile and application code), a `grafana` subchart dependency (official image, no local Dockerfile), and a `postgres` subchart (→ Managed Service):

```yaml
schemaVersion: v0.4
prepare:
  steps:
    - name: Install frontend dependencies
      command: cd frontend && npm install
    - name: Build frontend
      command: cd frontend && npm run build
    - name: Install backend dependencies
      command: cd backend && npm install
test:
  steps: []
run:
  frontend:
    steps:
      - name: Start frontend
        command: cd frontend && npm start
    plan: 8
    replicas: 1
    network:
      ports:
        - port: 3000
          isPublic: true
      paths:
        - port: 3000
          path: /
          stripPath: false
  backend:
    steps:
      - name: Start backend
        command: cd backend && npm start
    plan: 8
    replicas: 1
    env:
      DATABASE_URL: "postgresql://app:${{ vault.pgPassword }}@ws-server-${{ workspace.id }}-database.workspaces:5432/mydb"
    network:
      ports:
        - port: 3000
          isPublic: false
      paths:
        # stripPath here assumes backend's own routes are NOT already prefixed
        # with /api — verify against the real app before copying this value.
        - port: 3000
          path: /api
          stripPath: true
  grafana:
    image: grafana/grafana:10.4.2 # unmodified vendor image, no local Dockerfile → Managed Container
    plan: 8
    healthEndpoint: http://localhost:3000/api/health
    network:
      ports:
        - port: 3000
          isPublic: false
      paths:
        - port: 3000
          path: /grafana
          stripPath: true
    env:
      GF_SERVER_ROOT_URL: "%(protocol)s://%(domain)s/grafana" # Grafana also needs GF_SERVER_SERVE_FROM_SUB_PATH=true for a non-root path in practice
  database:
    provider:
      name: postgres # verify against GET /managed-services/providers
      schemaVersion: v1
    plan:
      id: 0
      parameters: { storage: 10000, cpu: 5, memory: 500 }
    config:
      version: "16"
      userName: app
      databaseName: mydb
    secrets:
      userPassword: "${{ vault.pgPassword }}"
      superuserPassword: "${{ vault.pgSuperuserPassword }}"
```

`frontend`/`backend` → Reactive (own Dockerfiles, own code). `grafana` → Managed Container (unmodified vendor image, classified via the table above — not asked as a single yes/no for the whole chart). `postgres` → Managed Service (step 1, before the per-component classification even runs).

### Helm Chart → Cloud Native Deployment (fallback — only when Kubernetes semantics are genuinely required)

1. Provision a **Virtual Kubernetes Cluster** managed service for the team (UI: Managed Services tab; API: `POST /managed-services` with `provider.name: "virtual-k8s"`). One per team.
2. Retrieve the kubeconfig via the API once provisioned.
3. Deploy the chart from a Landscape service's steps:

```yaml
schemaVersion: v0.4
prepare:
  steps:
    - name: Deploy with Helm
      command: |
        helm repo add corp42 https://charts.corp42.net
        helm repo update
        helm install my-release corp42/my-awesome-app \
          -n app --create-namespace -f values-codesphere.yaml
```

Or for a local chart already in the repo: `helm install my-release ./chart -n app --create-namespace -f values-codesphere.yaml`.

For raw manifests instead of a chart:

```yaml
steps:
  - name: Deploy Application
    command: |
      kubectl apply -f k8s/deployment.yaml
      kubectl apply -f k8s/service.yaml
```

- **What you keep:** the chart/manifests as the actual source of truth — no rewrite needed; in-cluster Services/ConfigMaps/Secrets all work as normal inside the virtual cluster.
- **What you lose:** automatic off-when-unused, automatic monitoring integration, and automatic Workspace Router networking integration for workloads running _inside_ the virtual cluster (manual wiring if the chart needs to be reachable through Codesphere's own domains/path routing) — **and** the customer takes on responsibility for guarantees the platform's own Landscape primitives otherwise provide automatically, e.g. proving confidential-compute/isolation properties for workloads running inside a self-managed cluster. This is why Cloud Native Deployment is the fallback, not the default reflex for "it's a Helm chart" — reserve it for the part of a chart that a Managed Service replacement plus the per-component Reactive/Managed Container split above genuinely can't cover.

## Common Pitfalls

- Assuming Codesphere builds a Dockerfile automatically because an `image:` field exists — it only _pulls_; building/pushing is still the user's responsibility (their own CI, or a Reactive rewrite).
- Migrating a compose `depends_on:` chain 1:1 and expecting Codesphere to enforce ordering — it doesn't; add retry/wait logic in the app itself.
- Migrating a named Docker volume straight into a `volumeMounts` entry — named/external volumes have no equivalent; the data must move into `/home/user/app` or Object Storage.
- Assuming a shell-command Compose healthcheck (`test: ["CMD", ...]`) has a direct equivalent — only HTTP checks are supported; expose an HTTP `/health` endpoint instead.
- Decomposing a trivial single-Deployment Helm chart into a Virtual Kubernetes Cluster instead of a native Reactive/Managed Container — loses off-when-unused, automatic path routing, and automatic monitoring for no real benefit, and for production hands the customer compliance burden (e.g. confidential-compute proof) the platform would otherwise carry.
- Treating "Reactive vs. Managed Container" as one chart-wide yes/no instead of a per-component decision — a chart with a custom frontend/backend plus a stock Grafana/admin-UI sidecar needs both types in the same Landscape, not a single answer forced onto every component.
- Rebuilding an unmodified vendor image as a Reactive because "Reactive is preferred" — that preference is about avoiding an unnecessary Cloud Native Deployment, not about avoiding Managed Container for a component that has no source to rebuild from in the first place.
- Assuming the compose provider name (`redis`, `mongo`, `mysql`) has a same-named Codesphere Managed Service provider — check `GET /managed-services/providers`; some are protocol-compatible equivalents under different names (e.g. Valkey for Redis protocol).

## Post-Migration Checklist

- [ ] Every previously-exposed port has a `network.paths` entry (or, for a Cloud Native Deployment, manual router wiring)
- [ ] Compose bind mounts / host paths re-homed under `/home/user/app`
- [ ] Compose named volumes replaced with Object Storage or workspace FS
- [ ] DB/cache/queue services replaced by verified Managed Services where a catalog equivalent exists, instead of hand-rolled containers
- [ ] All secrets moved to `${{ vault.NAME }}` — not left as plaintext `env:` values
- [ ] `.codesphere-internal/` added to `.gitignore`
- [ ] Any `docker-compose` service-name hostnames rewritten to Codesphere's internal DNS form or the managed service's connection env vars
- [ ] Healthchecks are HTTP-reachable at `healthEndpoint` (default `/` on port 3000) — shell-based Compose healthchecks need an HTTP equivalent

## Known Documentation Discrepancies

> This section may be outdated. Please verify against the official documentation:
> <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/configuring-a-landscape>

- The Redis-protocol-compatible provider is referred to as `valkey` consistently, but its exact `provider.version` (`v0` per the managed-services scrape vs. `v1` in this migration example) also differs across source documents — re-verify via the API rather than trusting either literal value here.

## Further Reading

- Official docs: <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/configuring-a-landscape>
- Landscape networking & multi-service patterns: [landscape.md](./landscape.md)
- CI pipeline field reference: [ci-pipeline.md](./ci-pipeline.md)
- Managed services architecture: `../managed-services/README.md`
- Virtual Kubernetes Cluster provider: `../managed-services/kubernetes.md`
- Valkey provider: `../managed-services/valkey.md`
