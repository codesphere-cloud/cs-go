# Codesphere Landscape Deployments Reference

> **Last updated:** 2026-07-28 · **Source:** https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/configuring-a-landscape

> The documentation changes regularly. If you are facing issues, please refer to the official documentation:
> <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/configuring-a-landscape>

## Overview

A **Landscape** is a deployment configuration for multiple, independently-scaling services within a single workspace, fully defined by one `ci.yml`. It can mix Reactives, Managed Containers, Cloud Native Deployments (via a Virtual Kubernetes Cluster managed service), and Managed Services. Use this reference for networking/routing, shared-vs-independent resource behavior, and multi-service `ci.yml` patterns. Per-service field shapes live in [ci-pipeline.md](./ci-pipeline.md); Landscape lifecycle (deploy/scale/teardown) API calls live in [cli-and-api.md](./cli-and-api.md).

## Core Concepts

- **Single-service is still a Landscape shape**: a one-service deployment still uses `schemaVersion: v0.4`+ with one named service under `run` — same format, just one entry.
- **Shared across all services**: filesystem (`/home/user/app`, `/nix/store`), the `prepare` stage, the `test` stage, and workspace-level env vars.
- **Independent per service**: CPU/RAM/storage (`plan`), replica count, deployment mode (always-on vs. off-when-unused), network configuration.
- **Internal DNS**: every service is reachable from every other service in the same Landscape at `http://ws-server-[WorkspaceId]-[serviceName].workspaces:[port]` — no explicit `networks:` declaration needed (unlike Docker Compose).
- **Headless / internal-only services**: a service can have `network.paths: []` (and no `isPublic: true` port) — it's simply not reachable from outside the Landscape, but stays reachable internally via the internal DNS name. Useful for workers, backends fronted by another service, or internal model-serving processes.
- **Managed Services in a Landscape are lifecycle-bound to it**: created on first deploy, destroyed when the Landscape is torn down — good for dev/staging/preview, not for data that must outlive the Landscape (use a standalone Managed Service instead).

## API / Syntax

### Private Communication Between Services

- **Description:** Internal DNS pattern for service-to-service calls inside the same Landscape.
- **Example:**

```
http://ws-server-[WorkspaceId]-[serviceName].workspaces:[port]
```

Workspace `abc123`, service `backend`, port `3000`:

```
http://ws-server-abc123-backend.workspaces:3000
```

You can copy a service's exact internal URL from the Landscape Config Editor (**Copy** button on the port row). Managed Services expose their own connection details via environment variables instead — see the relevant provider page.

### Path-Based Routing (Workspace Router)

- **Description:** Maps URL path prefixes to a service's port. Every service needs at least one route (`network.paths` entry, or `isPublic: true`) or the Workspace Router never marks it healthy/reachable.
- **Parameters:**

| Name                        | Type    | Required | Description                                                                                                                                  |
| --------------------------- | ------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `network.paths[].port`      | integer | Yes      | Port the path routes to.                                                                                                                     |
| `network.paths[].path`      | string  | Yes      | URL path prefix, e.g. `/`, `/api`.                                                                                                           |
| `network.paths[].stripPath` | boolean | Yes      | `true` removes the matched prefix before forwarding (`/api/users` → app sees `/users`); `false` keeps the full path (app sees `/api/users`). **Which one is correct depends entirely on whether the backend's own route definitions already include the prefix** — check the app's actual routes, don't default to either value; the "Frontend + Backend" example below uses `stripPath: true` specifically because that example's backend routes are unprefixed, not because `true` is the general default. |

- **Example:**

```yaml
run:
  frontend:
    steps:
      - command: npm run start:frontend
    plan: 8
    replicas: 1
    network:
      paths:
        - port: 3000
          path: /
          stripPath: false
  api:
    steps:
      - command: npm run start:api
    plan: 8
    replicas: 1
    network:
      paths:
        - port: 3000
          path: /api
          stripPath: true
```

- **Note:** `healthEndpoint` (default `http://localhost:3000/`) is checked from _inside_ the service's own container — target `localhost` unless the app binds elsewhere.

### Public Access

- **Description:** Public services with path routing get a shared domain + `workspace-id-port-service-name.codesphere.com`-style URL, or a custom domain connected to the workspace. Direct public port URLs (`isPublic: true` on a `network.ports` entry, without path routing) are supported but discouraged — prefer path-based routing.

## Complete Landscape Examples

### Frontend + Backend

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
  frontend:
    steps:
      - name: Start frontend
        command: npm run start:frontend
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
        command: npm run start:backend
    plan: 8
    replicas: 2
    network:
      ports:
        - port: 3000
          isPublic: false
      paths:
        - port: 3000
          path: /api
          stripPath: true
```

### App + Managed PostgreSQL

- **Example:**

```yaml
schemaVersion: v0.4
prepare:
  steps:
    - name: Install dependencies
      command: npm install
    - name: Run migrations
      command: npm run db:migrate
test:
  steps: []
run:
  app:
    steps:
      - name: Start application
        command: npm start
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
  database:
    provider:
      name: postgres
      schemaVersion: v1
    plan:
      id: 0
      parameters:
        storage: 10000
        cpu: 5
        memory: 500
    config:
      version: "17.6"
      userName: app
      databaseName: mydb
    secrets:
      userPassword: "${{ vault.pgPassword }}"
      superuserPassword: "${{ vault.pgSuperuserPassword }}"
```

- **Note:** verify the provider's actual `name`, supported `version`, and `config` fields with `GET /managed-services/providers` before deploying — plans/versions are installation-specific.

### App + Managed Container (nginx reverse-proxying a Reactive)

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
  app:
    steps:
      - name: Start app
        command: npm start
    plan: 8
    replicas: 1
    network:
      ports:
        - port: 3000
          isPublic: false
      paths: []
  edge:
    image: nginx-unprivileged:1.25-alpine
    plan: 8
    healthEndpoint: http://localhost/
    network:
      ports:
        - port: 80
          isPublic: true
      paths:
        - port: 80
          path: /
          stripPath: false
    volumeMounts:
      - name: _workspace
        mountPath: /etc/nginx/nginx.conf
        workspacePath: nginx.conf
```

`nginx.conf` would `proxy_pass` to `http://ws-server-[WorkspaceId]-app.workspaces:3000`.

### Ollama + Open WebUI (headless model service + public frontend)

- **Example:**

```yaml
schemaVersion: v0.4
prepare:
  steps:
    - name: Install ollama
      command: nix-env -iA nixpkgs.ollama
    - name: Install process-compose
      command: nix-env -iA nixpkgs.process-compose
    - name: Install uv
      command: nix-env -iA nixpkgs.uv
    - name: Install open-webui
      command: uv add open-webui && uv sync
test:
  steps: []
run:
  ollama:
    steps:
      - name: Run ollama
        command: process-compose -t=False -f process-compose.models.yaml
    plan: 22
    replicas: 1
    network:
      ports:
        - port: 11434
          isPublic: false
      paths: []
  open-webui:
    steps:
      - name: Run open-webui
        command: >
          . .env && export OLLAMA_BASE_URL=$OLLAMA_BASE_URL &&
          uv run open-webui serve --port 3000
    plan: 22
    replicas: 1
    network:
      ports:
        - port: 3000
          isPublic: true
      paths:
        - port: 3000
          path: /
          stripPath: false
```

### n8n + Ollama

- **Example:**

```yaml
schemaVersion: v0.4
prepare:
  steps:
    - name: Install Node.js 24
      command: nix-env -iA nixpkgs.nodejs_24
    - name: Install dependencies
      command: npm install
test:
  steps: []
run:
  n8n-frontend:
    steps:
      - name: n8n start
        command: N8N_PORT=3000 ./node_modules/n8n/bin/n8n start
    plan: 8
    replicas: 1
    network:
      ports:
        - port: 3000
          isPublic: false
      paths:
        - port: 3000
          path: /
          stripPath: false
  ollama-service:
    steps:
      - command: process-compose -t=False -f process-compose.models.yaml
    plan: 9
    replicas: 1
    network:
      ports:
        - port: 3000
          isPublic: false
      paths:
        - port: 3000
          path: /model
          stripPath: true
```

## Managed Services in a Landscape

|            | Landscape Managed Service          | Standalone Managed Service         |
| ---------- | ---------------------------------- | ---------------------------------- |
| Definition | In `ci.yml` (Git-tracked)          | UI or API                          |
| Lifecycle  | Created/deleted with the Landscape | Persists until manually deleted    |
| Use case   | Dev/test/preview environments      | Production, long-lived shared data |

## Service Naming

- Match monorepo directory names where possible.
- Lowercase with hyphens (`my-service`, not `myService`).
- The service name becomes part of the internal DNS name — **renaming a service (Reactive/Container or Managed Service) forces recreation**, which can mean data loss for stateful services. Treat names as fixed post-deploy.

## Common Pitfalls

- Assuming `depends_on`-style startup ordering exists between `run.<serviceName>` entries — it doesn't; add a wait/retry loop in the dependent app's own `run` step if ordering matters.
- Forgetting that a headless/internal service still needs `network: { ports: [...], paths: [] }` wired correctly for internal DNS to work, even with no public route.
- Renaming a service key in `ci.yml` expecting an in-place rename — it forces recreation instead.
- Using a Landscape Managed Service for data that must outlive the Landscape — it's destroyed on teardown; use a standalone Managed Service instead.
- Assuming custom Docker `networks:` declarations are needed — every service in a Landscape already shares one private network.

## Known Documentation Discrepancies

> This section may be outdated. Please verify against the official documentation:
> <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/configuring-a-landscape>

- "Headless services" (services with no public route) are documented here as a `network.paths: []` pattern inferred from the Ollama/n8n examples in the source docs, rather than from a dedicated headless-services page — cross-check against <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/headless-services> for any additional constraints (e.g. whether a `network` block can be omitted entirely for a fully internal service, vs. requiring an explicit empty `paths: []`).
- Managed Service `config`/`secrets` field names in the PostgreSQL example (`userName`, `databaseName`) are provider-specific and may differ for other providers/versions — always verify via `GET /managed-services/providers`.

## Further Reading

- Official docs: <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/configuring-a-landscape>
- Headless services: <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/headless-services>
- Landscape lifecycle (deploy/scale/teardown API): [cli-and-api.md](./cli-and-api.md)
- CI pipeline field reference: [ci-pipeline.md](./ci-pipeline.md)
- Language/framework runtime recipes: [runtimes.md](./runtimes.md)
- Managed services architecture: `../managed-services/README.md`
