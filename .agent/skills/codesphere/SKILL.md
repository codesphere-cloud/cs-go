---
name: codesphere
description: Use this skill for ANYTHING involving Codesphere — a cloud IDE/PaaS built around workspaces, `ci.yml` pipelines, Landscapes (multi-service deployments), and a Managed Services catalog (PostgreSQL, Babelfish, DocumentDB, Object Storage, Virtual Kubernetes Cluster, OpenSearch, RabbitMQ, Valkey). Always trigger this skill when the task is to create, write, edit, review, or debug a `ci.yml` or `ci.<profile>.yml`; create, write, or publish a `provider.yml` (custom Codesphere Service Provider); or migrate/port a `docker-compose.yml`, a bare Dockerfile, a Helm chart, or raw Kubernetes manifests specifically onto Codesphere, into a Reactive, Managed Container, or Cloud Native Deployment. Also trigger for: enabling or version-pinning a language runtime (Node, Python, Go, Ruby, PHP, Java, Rust, or anything else) via Nix inside a Codesphere `prepare`/`run` step; Reactives vs. Managed Containers vs. Managed Services; Landscape networking (internal `ws-server-*.workspaces` DNS, path routing) or scaling/replicas; the `cs-go` CLI or Codesphere's public API (workspaces, managed-services, vault, domains, teams endpoints); or a `${{ vault.NAME }}` / `${{ workspace.env[...] }}` template reference. For the managed-service providers (Postgres, Babelfish, DocumentDB/FerretDB, Valkey, RabbitMQ, OpenSearch, S3/Object Storage, Virtual Kubernetes Cluster), trigger when the question is about *Codesphere's managed version* of them — creating/configuring one as a service, its `provider:` block, its plan/config/secrets schema, connecting to Codesphere's `hostname`/`dsn` — but NOT for generic questions about running, debugging, or operating that same open-source project elsewhere (self-hosted, another cloud, bare Kubernetes, EC2, RDS) with no Codesphere artifact or context in sight. The same restraint applies to `docker-compose.yml` (only migrating it TO Codesphere counts) and vault/secrets questions (only Codesphere's `${{ vault.* }}` template syntax counts, not HashiCorp Vault or other secret managers). Trigger even when the user doesn't say "Codesphere" explicitly but the artifact or context makes it obvious — a `ci.yml`/`provider.yml` file open or pasted, `workspace.env`, `${{ vault.* }}` templates, `ws-server-*.workspaces` hostnames, `schemaVersion`, or a request to "deploy this on Codesphere" / "port this compose file to a landscape" / "turn this Helm chart into a managed service." When in doubt and there's a genuine Codesphere-specific artifact or term in play, trigger; when the query could just as easily be about the open-source project running anywhere else, don't.
license: none
allowed-tools: Read Glob Grep
metadata:
  version: "1.0.0"
  updated: "2026-07-28"
  cost-tier: low
---

> **Process:** Read the matching `references/*.md` file(s) from the routing table below before answering — do not answer from memory alone on anything version-, field-, or provider-name-specific. Several files below already correct real conflicts between Codesphere's own public docs and current reality; treat those corrections as settled, not as open questions.

## When to use this

Trigger for anything involving Codesphere: `ci.yml`/`provider.yml` creation, editing, or debugging; migrating a `docker-compose.yml`, Dockerfile, or Helm chart onto Codesphere; enabling/pinning a language runtime via Nix; Reactive vs. Managed Container vs. Managed Service questions; Landscape networking or scaling; the `cs-go` CLI or public API; vault secrets/env vars; or any managed-service provider **in Codesphere's own managed context** — not generic questions about that same open-source project running elsewhere. See the frontmatter `description` for the full trigger list, including artifact-based cues (`ci.yml` open or pasted, `${{ vault.* }}` templates, `ws-server-*.workspaces` hostnames).

## Hard Gate

- **Shared family conventions: see `references/skill-family-conventions.md`.** It covers why any skill in this family can freely read `codesphere`'s `references/*.md` files without loading `codesphere` itself as an active skill (this is what keeps the family's per-invocation cost flat as it grows — every `codesphere-add-<provider>`/`codesphere-create-<type>-deployment` skill added later needs the same read-only access to this skill's `references/`, never any of the others loaded), plus the shared `ci.yml`-at-repo-root rule every generating/editing skill follows.

- MUST read the matching `references/*.md` file before answering anything version-, field-, or provider-name-specific — never answer from memory alone on these.
- MUST prefer what the user states directly about their own Codesphere instance over anything in these reference files — they are a best-effort snapshot, not a live source of truth. If the user corrects something here, treat that as authoritative for the conversation and flag it back for the reference file to be fixed.
- MUST NOT present a "Load-bearing correction" below as uncertain or optional — these are confirmed fixes to Codesphere's own outdated public docs, not open questions.
- MUST NOT hardcode a provider name/version/plan for an unfamiliar or production target instance without noting that `GET /managed-services/providers` is the authoritative check — the catalog is installation-specific and drifts.

# Codesphere

Codesphere is a cloud IDE/PaaS. A **workspace** runs a `ci.yml` pipeline (`prepare` → `test` → `run`); a workspace with multiple named services in `run` is a **Landscape**. Services are one of three types: **Reactive** (`steps:`, Codesphere's own base image), **Managed Container** (`image:`, your own Docker image), or **Managed Service** (`provider:`, a catalog database/queue/etc.). This skill's `references/` hold the full field-level detail — this file is a router, not a substitute for reading them.

## Quick routing

| The user is asking about... | Read |
|---|---|
| `ci.yml` syntax, `schemaVersion`, Reactive/Managed Container/Managed Service field shapes, CI profiles | `references/ci-pipeline.md` |
| Enabling/pinning a language (Node, Python, Go, Ruby, PHP, Java, Rust, or anything else) via Nix; per-framework `prepare`/`run` recipes | `references/runtimes.md` |
| Multi-service Landscape networking (internal DNS, path routing, headless/internal-only services), shared vs. independent resources, multi-service examples | `references/landscape.md` |
| Deployment modes (always-on vs. off-when-unused), custom domains, zero-downtime releases, horizontal scaling/replicas, troubleshooting table | `references/deployment-guide.md` |
| Moving a `docker-compose.yml`, bare Dockerfile, Helm chart, or raw k8s manifests onto Codesphere | `references/migration-guide.md` |
| Plain workspace environment variables | `references/environment-variables.md` |
| Vault secrets (`${{ vault.NAME }}`), shared vaults — **preview feature** | `references/secret-management.md` |
| `cs-go` CLI commands, what the CLI does *not* cover, the full public API endpoint catalog, GitHub Actions/GitLab CI/Bitbucket integration | `references/cli-and-api.md` |
| Managed Services architecture, lifecycle (deploy/pause/delete/backups), the provider catalog, publishing a custom provider | `references/providers.md` |
| A specific managed-service provider's `config`/`secrets`/`details` schema and connection examples | `references/provider-<name>.md` — see table below |
| The exact JSON shape of a Managed Service API resource (`GET /managed-services`), the `status` state machine | `references/service-resource-schema.md` |
| Publishing a custom Landscape-based provider (`provider.yml`) | `references/custom-provider.md` |
| Shared conventions every skill in this family follows (locating/reading these reference files, `ci.yml` repo-root placement) | `references/skill-family-conventions.md` |

### Provider reference files

| Provider (`provider.name`) | File | Category |
|---|---|---|
| `postgres` | `references/provider-postgresql.md` | Database (GA) |
| `babelfish` | `references/provider-babelfish.md` | Database, SQL Server/TDS-compatible (preview) |
| `ferretdb` | `references/provider-documentdb.md` | Database, MongoDB-compatible (preview) |
| `valkey` | `references/provider-valkey.md` | Key-Value Store (closed testing) |
| `rabbitmq` | `references/provider-rabbitmq.md` | Message Queue (closed testing) |
| `opensearch` | `references/provider-opensearch.md` | Search & Analytics (closed testing) |
| `s3` | `references/provider-object-storage.md` | Storage (preview) |
| `virtual-k8s` | `references/provider-kubernetes.md` | Advanced Compute — team singleton (preview) |

Provider names/versions/plans are installation-specific and drift — always prefer `GET /managed-services/providers` over hardcoding when the user's target instance is known and reachable. Treat the reference files as the best available default, not a guarantee for a specific installation.

## Load-bearing corrections (don't regress these)

These are points where earlier drafts of this reference set were wrong or where Codesphere's own public docs are confirmed outdated. An agent generating a `ci.yml` today should default to the corrected version:

- **`schemaVersion: v0.4` is current — not `v0.2`.** Codesphere's public docs frame `v0.2` as the "safe default"; that's outdated per direct confirmation. Default new `ci.yml` files to `v0.4`, and use `provider.schemaVersion` (not the legacy `provider.version`) inside Managed Service blocks. See `references/ci-pipeline.md`.
- **The Virtual Kubernetes Cluster provider name is `virtual-k8s`** — not `virtual-kubernetes-cluster`, which appeared in some source material and has since been corrected throughout.
- **The Managed Service API resource requires `provider.version` AND `provider.schemaVersion` together**, per the platform's own JSON Schema — this is a different surface than the `ci.yml` *input*, which only uses one field name at a time depending on schema version. Don't conflate the two. See `references/service-resource-schema.md`.
- Java and Rust runtime examples in `references/runtimes.md` are extrapolated from the documented Nix pattern (no dedicated official guide exists for them at time of writing) — flagged in that file's own discrepancy section, verify `nixpkgs` attribute names before relying on them verbatim.
- The Valkey `provider.version`/`schemaVersion` value shows as `v0` in one source and `v1` in another across this reference set — still **unresolved**; verify against `GET /managed-services/providers` before hardcoding it.

## Core concepts an agent should never guess at

- **No root access.** No `apt`/`sudo`. Install OS-level packages/toolchains via Nix (`nix-env -iA nixpkgs.<package>`). Full detail: `references/runtimes.md`.
- **Only `/home/user/app` and `/nix/store` persist** and are shared across every service in a Landscape. Nothing else survives a restart.
- **`prepare`/`test` run once, on the main replica only, no auto-restart. `run` executes on every replica and auto-restarts on crash.**
- **Every Reactive/Managed Container needs at least one route** (`network.paths` entry or `isPublic: true`), or the Workspace Router never marks it healthy/reachable.
- **Renaming a service key (Reactive, Managed Container, or Managed Service) in `ci.yml` forces recreation** — can mean data loss for stateful services. Treat names as fixed post-deploy.
- **A Landscape-embedded Managed Service is destroyed when the Landscape is torn down.** Use a standalone Managed Service (created via UI/API, not `ci.yml`) for anything that must outlive the Landscape.
- **Vault secrets (`${{ vault.NAME }}`) are referenced, never inlined** in `ci.yml`. A Landscape sync fails if a referenced key was never initialized. This is a **preview** feature.

## Related

- `codesphere-create-cluster-deployment` — migrates an existing Helm chart onto Codesphere; decomposes it into a Landscape (Reactive/Managed Container/Managed Service per component) by default, falling back to a Cloud Native Deployment (Virtual Kubernetes Cluster) only for the part of a chart that genuinely needs Kubernetes semantics
- `codesphere-create-container-deployment` — generates a `ci.yml` for a Managed Container deployment from existing Dockerfiles/`docker-compose.yml`
- `codesphere-create-reactive-deployment` — generates a `ci.yml` with genuine Reactive services from application source
