---
name: codesphere-create-reactive-deployment
description: Generates a ci.yml with genuine Codesphere Reactive services (steps:, native prepare/run, Nix-installed toolchains where needed — never a helm/docker/kubectl command) from a repository's application source. Runs on explicit invocation. Detects each component's language/framework and maps it to the matching runtime recipe. Also checks whether components map to a Codesphere Managed Service (Postgres, Redis/Valkey, RabbitMQ, ...) instead of being deployed at all. Note: a Helm chart migration is handled by codesphere-create-cluster-deployment directly (it decides Reactive vs. Managed Container per component itself); this skill is for a repo with application source and no Helm chart involved.
license: none
allowed-tools: Bash Read Write Glob Grep
metadata:
  version: "1.0.0"
  updated: "2026-07-28"
  cost-tier: medium
---

> **Process:** When this skill is explicitly/directly invoked by name, execute it immediately — don't ask the user what they want done with it. Proceed straight to Step 0. The only prompts are the Decision Points defined within each Step below.

## When to use this

Trigger when the user wants a `ci.yml` that runs the application natively — no Docker, no Helm, no Kubernetes — e.g. "Reactive Deployment aus dem Quellcode erstellen", "die App nativ auf Codesphere starten", "codesphere-create-reactive-deployment ausführen". If the repository has a Helm chart, `codesphere-create-cluster-deployment` is the right entry point instead — it decides Reactive vs. Managed Container per component itself rather than handing off here.

## Hard Gate

- **Shared family conventions apply — see `references/skill-family-conventions.md`** (inside `codesphere`'s directory; this skill has no `references/` folder of its own — `Glob` for `**/codesphere/references/*.md` if the install path isn't already known). Covers locating/reading `codesphere`'s other `references/*.md` files (never requires `codesphere` itself to be loaded as an active skill) and the repo-root-only `ci.yml` placement used below.
- **MUST NOT contain a `helm`, `kubectl`, or `docker` command anywhere in the generated `ci.yml`, and no service may have an `image:` field.** This is not a stylistic preference — a `ci.yml` that still shells out to Helm or runs a container image is a Cloud Native Deployment or a Managed Container wearing a "reactive" label, not an actual Reactive deployment; earlier drafts of this skill family made exactly that mistake. Every component starts via its own native build/run commands instead, matching the patterns in `references/runtimes.md`. Run the Step 6 self-check before delivering, without exception.
- **MUST treat a Dockerfile-pinned runtime version (`FROM node:20-alpine`, `FROM python:3.11-slim`, etc.) as mandatory, not optional** — install the exact pinned version via Nix in `prepare`. Losing this pin silently is a real, previously-observed failure mode.
- **MUST use only `${{ ... }}` template syntax — never bare `{{ ... }}`** — and only documented fields (`${{ vault.NAME }}`, `${{ workspace.id }}`, `${{ workspace.devDomain }}`, `${{ team.id }}`, `${{ workspace.env['KEY'] }}`). Never invent a cross-service accessor like `${{ workspace.postgres.hostname }}` — fall back to a vault secret the user populates instead.
- **MUST only use fields documented in the `codesphere` reference set's Reactive field reference** (`steps`, `image`, `healthEndpoint`, `plan`, `replicas`, `isPublic`, `network`, `env`, `runAsUser`, `runAsGroup`, `volumeMounts`) — do not add fields like `displayName` that aren't confirmed to exist.

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
- **Yes** → **Decision Point**: ask the user whether the existing `ci.yml` should be overwritten or updated. "Update" means preserving the existing services/structure as much as possible and only adding/replacing the reactive service(s). "Overwrite" means rebuilding it from scratch.

**Output:** Overwrite-vs-update decision recorded.

**Blocking conditions:** Do not proceed without an answer when a `ci.yml` already exists.

### Step 2: Identify the deployable components

**Prerequisite:** Step 1 completed.

**Action:** Find every component with recognizable application source: a monorepo layout (e.g. `apps/backend/package.json`, `apps/frontend/package.json`) or a single app at the root. Manifest files to look for per language: `package.json` (Node), `requirements.txt`/`pyproject.toml`/`Pipfile` (Python), `go.mod` (Go), `Gemfile` (Ruby), `composer.json` (PHP), `pom.xml`/`build.gradle` (Java), `Cargo.toml` (Rust).

**Output:** Component list with each component's manifest file located.

**Blocking conditions:**
- **Nothing recognizable found** → **Blocker, abort.** Tell the user there's no application source this skill can work from. If a `Dockerfile` exists instead, suggest `codesphere-create-container-deployment`. If a Helm chart exists, suggest `codesphere-create-cluster-deployment`.

### Step 3: Reverse-check for a Helm chart

**Prerequisite:** Step 2 found at least one component.

**Action:** Check whether a Helm chart also exists in the repository (`Chart.yaml`). If it does, apply the same trivial/non-trivial read used in `codesphere-create-cluster-deployment`'s Step 4 (per `references/migration-guide.md`): CRDs, operators, or a StatefulSet with no managed-service equivalent among the chart's components are a sign real Kubernetes semantics might be needed.

**Decision Point (only if the chart looks non-trivial):** tell the user the repository has a Helm chart that looks like it needs real Kubernetes semantics (name the specific reason), and ask whether they still want a Reactive deployment anyway or would rather use `codesphere-create-cluster-deployment` instead. Proceed only on explicit confirmation.

**Output:** Confirmation to proceed with a Reactive deployment (immediate if no chart or a trivial chart; explicit if the chart is non-trivial).

**Blocking conditions:** For a non-trivial chart, do not proceed to Step 4 without explicit confirmation.

### Step 4: Detect each component's runtime

**Prerequisite:** Step 3 completed.

**Action:** Determine each component's language/framework from its manifest file and start script (`package.json`'s `scripts.start`/`scripts.build`, a Python `pyproject.toml`'s entry point, etc.). If the component also has a `Dockerfile` (even though it won't be used to build an image), its `RUN`/`CMD`/`ENTRYPOINT` lines are a useful secondary source of truth for build/start commands.

If the Dockerfile's `FROM` line pins a specific runtime version, that pin is mandatory — record the exact version now. `prepare` in Step 6 must install it via Nix (e.g. `nix-env -iA nixpkgs.nodejs_20`); Nix-installed packages persist automatically across `prepare`/`run` and don't need repeating, unlike a non-Nix switch such as `sudo n <version>`.

Match each component against the recipes in `references/runtimes.md`: Node/Next.js, Python (Pipenv or pip), Go, Ruby on Rails, PHP (Laravel), Vue.js are documented there directly; for anything else, follow the general Nix pattern in that file's "Other / unlisted languages" section.

**Decision Point (only if a component's start command can't be determined):** ask the user directly for the command that starts this component in production. Do not guess silently.

**Output:** Per-component runtime recipe (language, Nix package if version-pinned, build command, start command).

**Blocking conditions:** Do not proceed to Step 5 for a component whose start command is unresolved without asking the user first.

### Step 5: Detect managed-service candidates

**Prerequisite:** Step 4 completed.

**Action:** Look for components that are actually a database/cache/queue rather than application code (rare for source-only repos, but possible — e.g. a `docker-compose.yml` alongside the source declaring a `postgres` service the app's `DATABASE_URL` points at). Check `references/providers.md` and the individual `references/provider-*.md` files for a match:

| Component is / connects to | Possible replacement | Provider |
|---|---|---|
| PostgreSQL | `references/provider-postgresql.md` | `postgres` |
| MongoDB-compatible | `references/provider-documentdb.md` | `ferretdb` |
| Redis/Valkey | `references/provider-valkey.md` | `valkey` |
| RabbitMQ | `references/provider-rabbitmq.md` | `rabbitmq` |
| Elasticsearch/OpenSearch | `references/provider-opensearch.md` | `opensearch` |
| S3-compatible object storage/MinIO | `references/provider-object-storage.md` | `s3` |
| SQL Server-compatible | `references/provider-babelfish.md` | `babelfish` |

**Decision Point:** present each match individually — component, proposed managed service, what it means (existing data would need migration, not performed by this skill). The user decides per candidate.

**Output:** Per-component replace/keep decisions.

**Blocking conditions:** None.

### Step 6: Work out networking, then generate `ci.yml`

**Prerequisite:** Step 5 completed.

**Action:**
- For each component: decide public vs. internal — a component the end user reaches directly (typically the frontend) gets a `network.paths` entry; a component only reached internally gets `isPublic: false` and its own path prefix, or no public route if truly internal-only. Every service still needs at least one `network.paths` entry or `isPublic: true`.
- **`stripPath` depends on whether the component's own routes already include the path prefix — check its actual route definitions, don't default to either value.** `stripPath: true` forwards `/api/users` to the app as `/users`; `stripPath: false` forwards it unchanged. Picking the wrong one for how the app itself is written produces a working-looking `ci.yml` that 404s at runtime.
- `schemaVersion: v0.4` (current — not `v0.2`).
- One `run.<serviceName>` per component: `steps:` built from Step 4's findings (Nix install for any Dockerfile-pinned version, then build, then start — including repeating any non-Nix version pin like `sudo n <version>` in **both** `prepare` and `run`), `network` per above, `env:` for plain config, and any Step 5 managed-service connection details wired in as `${{ vault.NAME }}` references directly in `env:`.
- For each replacement confirmed in Step 5: an additional `run.<serviceName>` Managed Service block following the schema in the matching `references/provider-*.md`, with `secrets` as `${{ vault.NAME }}` references (never plaintext).
- Only use fields that actually appear in the Reactive field reference — see Hard Gate above.

**Output:** Draft `ci.yml`.

**Blocking conditions — run this 4-point self-check line by line before writing the file; every one of these has actually gone wrong in a real generation, they are not hypothetical:**
1. No `helm`, `kubectl`, or `docker` in any `command:` string; no service has an `image:` key.
2. Every component whose `Dockerfile` pinned a runtime version has a matching Nix install in `prepare`.
3. Every template reference uses `${{ ... }}` — never bare `{{ ... }}`.
4. Every template reference uses a real, documented field — no invented cross-service accessor.

If any of the 4 points turned up a problem, fix it now — do not proceed to Step 7 with a known issue still in the draft.

### Step 7: Write the file

**Prerequisite:** Step 6's self-check passed cleanly.

**Action:** Place `ci.yml` at the repository root (overwrite/update per the Step 1 decision). No other file needs to change — a Helm chart or Dockerfile that might exist alongside the source stays untouched.

**Output:** `ci.yml` written.

**Blocking conditions:** None.

### Step 8: Summary

**Prerequisite:** Step 7 completed.

**Action:** Briefly summarize for the user:
- Every component and which runtime recipe it was mapped to (or that the start command came from the user directly).
- Which managed-service replacements were made vs. declined in Step 5.
- The networking decision made per component in Step 6.
- Which `${{ vault.* }}` references still need real values populated before the first sync.
- Explicit confirmation that no Helm/Docker/Kubernetes is involved anywhere in the generated file.

**Output:** One consolidated status message the user can act on.

**Blocking conditions:** None.

## Related

- `codesphere` — reference knowledge this skill reads from (loose coupling, read-only, never auto-invoked)
- `codesphere-create-cluster-deployment` — handles a Helm-chart migration itself (including any Reactive components); not a hand-off source into this skill
- `codesphere-create-container-deployment` — alternative when the user prefers existing Docker images over a native rebuild
