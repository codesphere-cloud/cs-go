---
name: codesphere-create-container-deployment
description: 'Generates a ci.yml for a Codesphere Managed Container deployment (one image: per component) from a repository''s Dockerfile(s) and/or docker-compose.yml. Codesphere only pulls pre-built images — it never builds or pushes them, so the referenced images must already exist in a registry. Also checks whether components map onto a Codesphere Managed Service (Postgres, Redis/Valkey, RabbitMQ, ...) instead of staying containers. Only run on explicit invocation; a repo with a Helm chart is codesphere-create-cluster-deployment''s job instead, and a repo with only application source (no Dockerfile/compose) is codesphere-create-reactive-deployment''s.'
license: none
allowed-tools: Bash Read Write Glob Grep
metadata:
  version: "2.0.0"
  updated: "2026-09-23"
  cost-tier: medium
---

> **Process:** When this skill is explicitly/directly invoked by name, execute it immediately — don't ask the user what they want done with it.

## When to use this

Trigger when the user wants a `ci.yml` generated from existing Docker images/Dockerfiles rather than Helm or a native rebuild — e.g. "Container Deployment aus meinen Dockerfiles erstellen", "ci.yml für die vorhandenen Docker-Images bauen", "codesphere-create-container-deployment ausführen". If the repository has a Helm chart, use `codesphere-create-cluster-deployment` instead — it decides Reactive vs. Managed Container per component itself rather than handing off here. If there's no Dockerfile/compose file but there is application source, use `codesphere-create-reactive-deployment`.

## Reference

This skill has no `references/` folder of its own. Read `references/skill-family-conventions.md` (inside `codesphere`'s directory — `Glob` for `**/codesphere/references/*.md` if the install path isn't already known) for how to locate/read `codesphere`'s other reference files without loading `codesphere` itself, and for the repo-root-only `ci.yml` placement convention. For managed-service matches, read the matching `references/provider-*.md` for its exact `ci.yml` schema. This skill doesn't repeat their content.

## Process

1. `ci.yml` always goes at the repo root, even in a monorepo with multiple components. If one already exists, ask whether to **update** (keep existing services, add/replace this one) or **overwrite** (rebuild from scratch) — don't proceed without an answer.
2. Find deployable components by searching **both** Dockerfiles (per-component, or a single root one) and `docker-compose.yml`/`.yaml`/`compose.yml`, merging results into one list rather than treating them as alternatives. A component described in both sources is one component, not two. A compose service defined only with `image:` (no `build:`) is still a real component even with no Dockerfile anywhere in the repo, and its `image:` value already answers step 3's question for it. Neither source found anywhere → abort, tell the user there's nothing to containerize, and suggest `codesphere-create-reactive-deployment` if recognizable application source exists instead.
3. For every component with a Dockerfile+`build:` and no known `image:` yet, ask whether an existing CI pipeline (GitHub Actions, GitLab CI, etc.) already builds and pushes it, and get the registry reference if so. If no such pipeline exists, don't invent one — still generate the `image:` reference, but flag it in the step 8 summary as not yet backed by a real build.
4. Detect managed-service candidates: a component running its own DB/cache/queue image, or implied by another component's env vars (`DATABASE_URL`, `REDIS_URL`, etc.). Match against `references/providers.md`/`references/provider-*.md`: PostgreSQL → provider-postgresql.md (`postgres`); MongoDB-compatible → provider-documentdb.md (`ferretdb`); Redis/Valkey → provider-valkey.md (`valkey`); RabbitMQ → provider-rabbitmq.md (`rabbitmq`); Elasticsearch/OpenSearch → provider-opensearch.md (`opensearch`); S3-compatible/MinIO → provider-object-storage.md (`s3`); SQL Server-compatible → provider-babelfish.md (`babelfish`). Present each match individually — component, proposed service, what changes (existing data needs migration, not performed by this skill) — and let the user decide replace-or-keep per candidate. Nothing swaps silently.
5. For each component that isn't a managed service, work out networking: read the Dockerfile's `EXPOSE` (and compose `ports:`) for the port it actually listens on; decide public vs. internal (end-user-facing, typically the frontend → a `network.paths` entry at `/`; internal-only, typically the backend → `isPublic: false` plus its own path prefix, or no public route at all). Translate a Dockerfile `HEALTHCHECK` into `healthEndpoint` if present, otherwise leave the platform default and flag if it doesn't match the real port. `stripPath` must match the component's actual route definitions, not a default guess — the wrong value produces a `ci.yml` that looks right but 404s at runtime.
6. Generate `ci.yml` at `schemaVersion: v0.4`, with `prepare:` and `test:` present explicitly — even as `steps: []` — never omitted. One `run.<serviceName>` per component: `image:` from step 3, `command:` only if it needs to override the image's default `CMD`, `network` from step 5, `env:` for plain config, and any managed-service connection details wired in as `${{ vault.NAME }}` references directly in `env:` — no Helm, no `--set`. Each confirmed step 4 replacement gets its own `run.<serviceName>` Managed Service block per its `references/provider-*.md` schema, with secrets as `${{ vault.NAME }}` references, never plaintext. Before finishing, scan the draft for `helm`/`kubectl`/`virtual-k8s` and remove/reconsider anything found — their presence is a sign step 2 went wrong.
7. Write `ci.yml` to the repo root per the step 1 decision. Nothing else changes — Dockerfiles and compose files are only ever read, never edited.
8. Summarize: every component and where its image is expected to come from (explicitly flag any whose build/push pipeline doesn't exist yet), which managed-service replacements were made vs. declined, the networking decision per component, and which `${{ vault.* }}` references still need real values before the first sync.

## Keep in mind

- Never build, push, or run anything — no `docker build`, no `docker push`, no `cs start`, no `POST /workspaces/{id}/landscape/deploy`. Building and pushing the images referenced in the generated `ci.yml` stays the user's own responsibility (their existing CI, or a manual step); Codesphere pulls a pre-built image, it never builds one from a Dockerfile. Don't let a Dockerfile's presence in the repo imply this skill will build it.
- No `helm`/`kubectl` anywhere in the generated `ci.yml` — verify this explicitly before finishing (step 6), even if a Helm chart happens to exist in the same repository.
- Step 2 must search both Dockerfiles and compose files and merge the results — never just whichever is found first, and never treat the two sources as alternatives.
- `prepare:` and `test:` must always be present explicitly, even as `steps: []` — don't drop keys just because they'd be empty.
- Renaming a service key in `ci.yml` after deploy forces recreation — treat a name as fixed once chosen.

## Related

- `codesphere` — reference knowledge this skill reads from (loose coupling, read-only, never auto-invoked)
- `codesphere-create-cluster-deployment` — handles a Helm-chart migration itself, including any Managed Container components; not a hand-off source into this skill
- `codesphere-create-reactive-deployment` — alternative when no Dockerfile exists but application source does
