---
name: codesphere-create-reactive-deployment
description: 'Generates a ci.yml with genuine Codesphere Reactive services (steps:, native prepare/run, Nix-installed toolchains — never a helm/docker/kubectl command) from a repository''s application source. Only run on explicit invocation. Detects each component''s language/framework, maps it to a runtime recipe, and checks whether components map to a Managed Service instead of being deployed at all. A Helm chart migration is handled by codesphere-create-cluster-deployment directly, not here — this skill is for application source with no Helm chart involved.'
license: none
allowed-tools: Bash Read Write Glob Grep
metadata:
  version: "1.0.0"
  updated: "2026-07-28"
  cost-tier: medium
---

> **Process:** When this skill is explicitly/directly invoked by name, execute it immediately — don't ask the user what they want done with it.

## When to use this

Trigger when the user wants a `ci.yml` that runs the application natively — no Docker, no Helm, no Kubernetes — e.g. "Reactive Deployment aus dem Quellcode erstellen", "die App nativ auf Codesphere starten", "codesphere-create-reactive-deployment ausführen". If the repository has a Helm chart, `codesphere-create-cluster-deployment` is the right entry point instead — it decides Reactive vs. Managed Container per component itself rather than handing off here.

## Reference

This skill has no `references/` folder of its own. See `references/skill-family-conventions.md` (inside the sibling `codesphere` skill's directory — `Glob` for `**/codesphere/references/*.md` if the install path isn't already known) for how to locate/read `codesphere`'s other reference files without loading `codesphere` itself, and for the repo-root-only `ci.yml` placement convention. Before generating, read: `references/runtimes.md` (per-language recipes and the Reactive field reference), `references/providers.md` + the individual `references/provider-*.md` files (Managed Service schemas), `references/migration-guide.md` (the trivial/non-trivial read used in the Helm reverse-check below), and `references/ci-pipeline.md` (the networking Common Pitfalls this skill's Keep-in-mind section restates). This skill doesn't repeat their content — it's the process wrapper.

## Process

1. `ci.yml` always goes at the repo root, even in a monorepo with multiple components. If one already exists, ask whether to **update** (keep existing services, add/replace this one) or **overwrite** — don't proceed without an answer.
2. Find every component with recognizable application source via its manifest file: `package.json` (Node), `requirements.txt`/`pyproject.toml`/`Pipfile` (Python), `go.mod` (Go), `Gemfile` (Ruby), `composer.json` (PHP), `pom.xml`/`build.gradle` (Java), `Cargo.toml` (Rust). Nothing found → stop; suggest `codesphere-create-container-deployment` if a Dockerfile exists, or `codesphere-create-cluster-deployment` if a Helm chart exists.
3. Reverse-check for a Helm chart (`Chart.yaml`). If one exists and looks non-trivial (CRDs, operators, or a StatefulSet with no managed-service equivalent among its components — per `references/migration-guide.md`'s trivial/non-trivial read), tell the user the specific reason and ask whether they still want Reactive anyway or would rather use `codesphere-create-cluster-deployment`; proceed only on explicit confirmation. No chart, or a trivial one, → proceed without asking.
4. Detect each component's runtime from its manifest and start script (`package.json`'s `scripts.start`, a Python entry point, etc.); if it also has a `Dockerfile`, its `RUN`/`CMD`/`ENTRYPOINT` lines are a useful secondary source even though it won't be built. A pinned `FROM` version is mandatory to record now — it gets installed via Nix in `prepare` (step 6). Match against `references/runtimes.md`'s recipes (Node/Next.js, Python, Go, Rails, Laravel, Vue documented directly; anything else follows the general Nix pattern). Can't determine a start command → ask the user directly, don't guess silently.
5. Check for managed-service candidates — a component that's actually a DB/cache/queue rather than application code (rare in source-only repos, but possible, e.g. a `docker-compose.yml` alongside the source declaring a `postgres` service the app's `DATABASE_URL` points at) — against this table: PostgreSQL → `provider-postgresql.md`/`postgres`; MongoDB-compatible → `provider-documentdb.md`/`ferretdb`; Redis/Valkey → `provider-valkey.md`/`valkey`; RabbitMQ → `provider-rabbitmq.md`/`rabbitmq`; Elasticsearch/OpenSearch → `provider-opensearch.md`/`opensearch`; S3-compatible/MinIO → `provider-object-storage.md`/`s3`; SQL Server-compatible → `provider-babelfish.md`/`babelfish`. Present each match individually — component, proposed service, what changes (existing data needs migration, not performed by this skill) — and let the user decide replace-or-keep per candidate.
6. Work out networking per component (public vs. internal reach, `stripPath` from actual routes — see Keep in mind) and generate the draft `ci.yml` (`schemaVersion: v0.4`, current — not `v0.2`): one `run.<serviceName>` per component with `steps:` (Nix install for any pinned version, then build, then start — repeating any non-Nix pin like `sudo n <version>` in both `prepare` and `run`), `network`, `env:` for plain config plus any step-5 connection details as `${{ vault.NAME }}`, and one additional `run.<serviceName>` Managed Service block per confirmed replacement following its `references/provider-*.md` schema, with secrets always as `${{ vault.NAME }}` (never plaintext).
7. Before writing, run the pre-write self-check line by line (see Keep in mind) and fix anything it turns up — do not write with a known issue still in the draft.
8. Write `ci.yml` at the repo root per the step-1 decision. Nothing else needs to change — a Helm chart or Dockerfile alongside the source stays untouched.
9. Summarize for the user: each component and which runtime recipe it was mapped to (or that the start command came from the user directly); which managed-service replacements were made vs. declined; the networking decision per component; which `${{ vault.* }}` references still need real values before the first sync; and explicit confirmation that no Helm/Docker/Kubernetes is involved anywhere in the generated file.

## Keep in mind

- No `helm`, `kubectl`, or `docker` command anywhere in the generated `ci.yml`, and no service may have an `image:` field. This is the core identity of a Reactive deployment, not a style preference — a `ci.yml` that still shells out to Helm or runs a container image is a Cloud Native Deployment or Managed Container wearing a "reactive" label; earlier drafts of this skill family made exactly that mistake.
- A Dockerfile-pinned runtime version (e.g. `FROM node:20-alpine`) is mandatory, not optional — install the exact version via Nix in `prepare`. Losing this pin silently is a real, previously-observed failure mode. Nix-installed packages persist automatically across `prepare`/`run`; a non-Nix version switch does not and must be repeated in both.
- Only `${{ ... }}` template syntax, never bare `{{ ... }}`, and only documented fields: `${{ vault.NAME }}`, `${{ workspace.id }}`, `${{ workspace.devDomain }}`, `${{ team.id }}`, `${{ workspace.env['KEY'] }}`. Never invent a cross-service accessor like `${{ workspace.postgres.hostname }}` — fall back to a vault secret instead.
- Only use fields documented in the Reactive field reference: `steps`, `image`, `healthEndpoint`, `plan`, `replicas`, `isPublic`, `network`, `env`, `runAsUser`, `runAsGroup`, `volumeMounts`. Don't add undocumented fields like `displayName`.
- Networking — confirmed real-platform failures, not hypothetical: every `network` block must include both `ports` and `paths` as keys, never just one — a `paths`-only block (no `ports` key at all) parses fine but fails real validation with a garbled union-type error. Every service needs a `_workspace` `volumeMounts` entry (`mountPath: /home/user/app`, `workspacePath: ""`) — documented as optional but confirmed necessary in practice, and the other confirmed contributor to that same failure. `stripPath` must match whether the component's own routes already include the path prefix — check its actual route definitions, don't default to either value; the wrong choice produces a working-looking `ci.yml` that 404s at runtime.
- Pre-write self-check — run this line by line before writing the file; every point has actually gone wrong in a real generation:
  1. No `helm`, `kubectl`, or `docker` in any `command:` string; no service has an `image:` key.
  2. Every component whose Dockerfile pinned a runtime version has a matching Nix install in `prepare`.
  3. Every template reference uses `${{ ... }}` — never bare `{{ ... }}`.
  4. Every template reference is a real, documented field — no invented cross-service accessor.
  5. Every service's `network` block has both `ports` and `paths` present, and a `_workspace` `volumeMounts` entry.

## Related

- `codesphere` — reference knowledge this skill reads from (loose coupling, read-only, never auto-invoked): `references/runtimes.md` (Reactive recipes and field reference), `references/providers.md` + `references/provider-*.md` (Managed Service schemas), `references/migration-guide.md` (Helm reverse-check), `references/ci-pipeline.md` (networking pitfalls)
- `codesphere-create-cluster-deployment` — handles a Helm-chart migration itself (including any Reactive components); not a hand-off source into this skill
- `codesphere-create-container-deployment` — alternative when the user prefers existing Docker images over a native rebuild
