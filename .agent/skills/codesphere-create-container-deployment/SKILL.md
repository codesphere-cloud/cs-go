---
name: codesphere-create-container-deployment
description: Generates a ci.yml for a Codesphere Managed Container deployment (image: field, one per component) from a repository with existing Dockerfile(s) or a docker-compose.yml. Runs on explicit invocation. Flags clearly that Codesphere does not build Dockerfiles itself — the images need to already be built and pushed to a registry. Also checks whether components map to a Codesphere Managed Service (Postgres, Redis/Valkey, RabbitMQ, ...) instead of staying containers. Note: a Helm chart migration is handled by codesphere-create-cluster-deployment directly (it decides Reactive vs. Managed Container per component itself); this skill is for a repo with Dockerfiles/docker-compose.yml and no Helm chart involved.
license: none
allowed-tools: Bash Read Write Glob Grep
metadata:
  version: "1.0.0"
  updated: "2026-07-28"
  cost-tier: medium
---

> **Process:** When this skill is explicitly/directly invoked by name, execute it immediately — don't ask the user what they want done with it. Proceed straight to Step 0. The only prompts are the Decision Points defined within each Step below.

## When to use this

Trigger when the user wants a `ci.yml` generated using existing Docker images/Dockerfiles rather than Helm or a native rebuild — e.g. "Container Deployment aus meinen Dockerfiles erstellen", "ci.yml für die vorhandenen Docker-Images bauen", "codesphere-create-container-deployment ausführen". If the repository has a Helm chart, `codesphere-create-cluster-deployment` is the right entry point instead — it decides Reactive vs. Managed Container per component itself rather than handing off here. If the user names only the SBOM/publiccode.yml equivalent for a *different* skill family, that's out of scope here.

## Hard Gate

- **Shared family conventions apply — see `references/skill-family-conventions.md`** (inside `codesphere`'s directory; this skill has no `references/` folder of its own — `Glob` for `**/codesphere/references/*.md` if the install path isn't already known). Covers locating/reading `codesphere`'s other `references/*.md` files (never requires `codesphere` itself to be loaded as an active skill) and the repo-root-only `ci.yml` placement used below.
- **MUST NOT build, push, or run anything.** No `docker build`, no `docker push`, no `cs start`, no `POST /workspaces/{id}/landscape/deploy`. Building and pushing the images referenced in the generated `ci.yml` remains the user's own responsibility (their existing CI, or a manual step) — Codesphere pulls a pre-built image, it never builds one from a Dockerfile. Do not let the presence of a `Dockerfile` in the repo create the impression this skill will build it.
- **MUST search both Dockerfiles and `docker-compose.yml`/`.yaml` in Step 2, not just whichever is found first**, and merge the results into one component list rather than treating them as alternatives.
- **MUST NOT contain a `helm` or `kubectl` command anywhere in the generated `ci.yml`** — verify this explicitly before finishing Step 6, even if a Helm chart exists in the same repository.
- **MUST include `prepare:` and `test:` explicitly** in the generated `ci.yml`, even as `steps: []`, matching every worked example in the `codesphere` reference set — don't drop keys just because they'd be empty.

## Process: 8-Step Workflow

### Step 0: Determine repo root

**Prerequisite:** Skill invoked.

**Action:** `ci.yml` always belongs at the repository root — never in a subdirectory, even in a monorepo with multiple components.

**Output:** Confirmed target path for `ci.yml` (repo root).

**Blocking conditions:** None.

### Step 1: Check for an existing `ci.yml`

**Prerequisite:** Step 0 completed.

**Action:** Check whether a `ci.yml` already exists at the root.

- **No** → continue to Step 2.
- **Yes** → **Decision Point**: ask the user whether the existing `ci.yml` should be overwritten or updated. "Update" means preserving the existing services/structure as much as possible and only adding/replacing the containerized service(s). "Overwrite" means rebuilding it from scratch.

**Output:** Overwrite-vs-update decision recorded.

**Blocking conditions:** Do not proceed without an answer when a `ci.yml` already exists.

### Step 2: Identify the deployable components

**Prerequisite:** Step 1 completed.

**Action:** Search **both** of the following sources — not just whichever is found first — and merge the results into one component list:

1. **Dockerfiles**: every component with its own `Dockerfile` (e.g. `apps/backend/Dockerfile`, `apps/frontend/Dockerfile`), or a single root-level `Dockerfile` for a non-monorepo.
2. **`docker-compose.yml`** (or `docker-compose.yaml`, `compose.yml`): every service it declares, whether via `build:` (a local Dockerfile context — cross-reference against source 1, likely the same component) or `image:` (an already-published image, which may have no local `Dockerfile` at all).

When both sources describe the same component, treat it as one component, not two. A compose service defined purely with `image:` and no `build:` is still a real component even without a `Dockerfile` anywhere in the repo. A compose service's `image:` value already answers Step 3's "where does this image come from" question — don't ask the user again for that component.

**Output:** Merged component list, each tagged with its known image source (`Dockerfile`+`build:`, or a compose `image:` value already known).

**Blocking conditions:**
- **Neither a Dockerfile nor a compose file found anywhere** → **Blocker, abort.** Tell the user there's nothing to containerize here and this skill can't proceed. If the repository has recognizable application source instead (e.g. `package.json`, `requirements.txt`, `go.mod`), suggest `codesphere-create-reactive-deployment` instead.

### Step 3: Determine where each image actually lives

**Prerequisite:** Step 2 produced a component list.

**Action:** For every component with a `Dockerfile` and a `build:` context (no known `image:` yet), this skill needs to know where its **built** image will come from — Codesphere only pulls, it never builds. Skip this for any component whose `docker-compose.yml` entry already has an `image:` value — use that value directly instead of asking.

**Decision Point:** for the remaining components, ask the user, per component (or once, if the answer is the same for all): is there already a CI pipeline (GitHub Actions, GitLab CI, etc.) that builds and pushes this image to a registry? If yes, get the registry/image reference (e.g. `registry.example.com/myorg/backend:latest`). If no such pipeline exists yet, say so plainly in the Step 8 summary rather than silently generating an `image:` reference that has nothing behind it.

**Output:** A resolved `image:` reference for every component.

**Blocking conditions:** None — a component without an existing pipeline still gets an `image:` reference, just flagged as not-yet-backed in the summary.

### Step 4: Detect managed-service candidates

**Prerequisite:** Step 3 completed.

**Action:** Look for database/cache/queue components — either their own `Dockerfile`/image (e.g. an official `postgres`/`redis`/`rabbitmq` image referenced in `docker-compose.yml`) or a dependency implied by another component's env vars (`DATABASE_URL`, `REDIS_URL`, etc.). Check `references/providers.md` and the individual `references/provider-*.md` files for a match:

| Component runs | Possible replacement | Provider |
|---|---|---|
| PostgreSQL | `references/provider-postgresql.md` | `postgres` |
| MongoDB-compatible | `references/provider-documentdb.md` | `ferretdb` |
| Redis/Valkey | `references/provider-valkey.md` | `valkey` |
| RabbitMQ | `references/provider-rabbitmq.md` | `rabbitmq` |
| Elasticsearch/OpenSearch | `references/provider-opensearch.md` | `opensearch` |
| S3-compatible object storage/MinIO | `references/provider-object-storage.md` | `s3` |
| SQL Server-compatible | `references/provider-babelfish.md` | `babelfish` |

**Decision Point:** present each match individually — component, proposed managed service, and what it means (existing data would need migration, not performed by this skill). The user decides per candidate: replace with a Managed Service, or keep it as its own container.

**Output:** Per-component replace/keep decisions.

**Blocking conditions:** None.

### Step 5: Work out networking for each component

**Prerequisite:** Step 4 completed.

**Action:** For each component that isn't a managed service: read its `Dockerfile`'s `EXPOSE` (and `docker-compose.yml`'s `ports:` if present) to find the port it actually listens on. Decide public vs. internal: a component the end user is meant to reach directly (typically the frontend) gets a `network.paths` entry at `/`; a component only other services should reach (typically the backend) gets `isPublic: false` and its own `path` prefix, or no public route if truly internal-only. If the Dockerfile has a `HEALTHCHECK` instruction, translate its target into `healthEndpoint` — otherwise leave the platform default (`http://localhost:3000/`) and flag if that doesn't match the actual listening port. **`stripPath` depends on whether the component's own routes already include the path prefix** — check the app's actual route definitions (or its source if available) rather than defaulting to either value; the wrong choice produces a `ci.yml` that looks right but 404s at runtime.

**Output:** Networking configuration per component.

**Blocking conditions:** None.

### Step 6: Generate `ci.yml`

**Prerequisite:** Step 5 completed.

**Action:**
- `schemaVersion: v0.4` (current — not `v0.2`).
- Include `prepare:` and `test:` explicitly, even as `steps: []` — a pure Managed Container Landscape usually has nothing to put in either stage, but that's still `steps: []`, not omitting the key.
- One `run.<serviceName>` per component: `image:` (Step 3), `command:` only if it needs to override the image's default `CMD` (rare), `network` per Step 5, `env:` for plain config, and any Step 4 managed-service connection details wired in as `${{ vault.NAME }}` references directly in `env:` — no Helm, no `--set` overrides.
- For each replacement confirmed in Step 4: an additional `run.<serviceName>` Managed Service block following the schema in the matching `references/provider-*.md`, with `secrets` as `${{ vault.NAME }}` references (never plaintext).

**Output:** Draft `ci.yml`.

**Blocking conditions:** Before finishing this step, scan the draft `ci.yml` for `helm`/`kubectl`/`virtual-k8s` — their presence is a sign Step 2 went wrong. Remove/reconsider anything found.

### Step 7: Write the file

**Prerequisite:** Step 6 produced a validated draft `ci.yml`.

**Action:** Place `ci.yml` at the repository root (overwrite/update per the Step 1 decision). No other file needs to change — the Dockerfiles themselves stay as-is, this skill only ever reads them.

**Output:** `ci.yml` written.

**Blocking conditions:** None.

### Step 8: Summary

**Prerequisite:** Step 7 completed.

**Action:** Briefly summarize for the user:
- Every component containerized, and where its image is expected to come from — explicitly flag any component whose build/push pipeline doesn't exist yet.
- Which managed-service replacements were made vs. declined in Step 4.
- The networking decision made per component in Step 5.
- Which `${{ vault.* }}` references still need real values populated before the first sync.

**Output:** One consolidated status message the user can act on.

**Blocking conditions:** None.

## Related

- `codesphere` — reference knowledge this skill reads from (loose coupling, read-only, never auto-invoked)
- `codesphere-create-cluster-deployment` — handles a Helm-chart migration itself (including any Managed Container components); not a hand-off source into this skill
- `codesphere-create-reactive-deployment` — alternative when no Dockerfile exists but application source does
