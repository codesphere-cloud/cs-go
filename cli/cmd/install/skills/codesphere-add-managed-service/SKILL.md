---
name: codesphere-add-managed-service
description: 'Adds a Codesphere managed service (PostgreSQL, Babelfish, DocumentDB/FerretDB, Valkey, RabbitMQ, OpenSearch, S3/Object Storage, Virtual Kubernetes Cluster, or any other documented provider) to a ci.yml — creates one if none exists, or adds/edits a provider block in an existing file. Reads the exact config/secrets/plan schema from the matching provider reference file rather than hardcoding it, asks only for what isn''t already stated or safely defaultable, and asks replace-vs-parallel when a service of that provider already exists. Trigger for "postgres/rabbitmq/valkey/... hinzufügen", "add a database/cache/queue/storage to my ci.yml", "zweite <provider> instanz", or any request to add/configure a managed service, regardless of which provider.'
license: none
allowed-tools: Bash Read Write Glob Grep
metadata:
  version: "1.1.0"
  updated: "2026-09-23"
  cost-tier: medium
---

> **Process:** When this skill is explicitly/directly invoked by name, execute it immediately — don't ask the user what they want done with it.

## When to use this

Trigger when the user wants any Codesphere managed-service provider added to or configured within a `ci.yml` — e.g. "füge postgres zu meinem ci.yml hinzu", "add a managed valkey cache", "ich brauch noch eine zweite rabbitmq-Instanz parallel", "s3-Bucket als Managed Service konfigurieren". This is one generic skill covering every provider — **don't create a separate `codesphere-add-<provider>` skill per provider**; a new provider means adding its `provider-<name>.md` file to `codesphere`, not a new skill (see "Adding a new provider" below). Narrower than `codesphere-create-cluster-deployment` / `-container-deployment` / `-reactive-deployment`, which generate a whole deployment: this skill only ever adds or edits one managed-service block within whatever `ci.yml` already exists.

## Reference

Read `references/providers.md` (provider catalog — authoritative for which providers exist and which file documents each), the target `references/provider-<name>.md` (config/secrets/plan schema, backups support, provider-specific pitfalls), `references/ci-pipeline.md` (Managed Service field shape, including `backups:`), and `references/secret-management.md` — all inside the sibling `codesphere` skill (`references/skill-family-conventions.md` covers locating them without loading `codesphere` itself). This skill's own `references/interview-<name>.md` files, when present for the target provider, are resolved **locally** in this skill's own directory instead.

## Process

1. Confirm the repo root (`ci.yml` always lives there). If none exists, note a new file will be created later (`schemaVersion: v0.4`). If one exists, read it — note its `schemaVersion` (decides `provider.version` vs. `provider.schemaVersion` later), every existing `provider.name` found, and any existing `s3` service specifically (a reuse candidate for backups). Stop if the file is `v0.1`-shaped (no `schemaVersion`, flat `run.steps`) — it can't host a Managed Service without migrating first; ask the user how they want to proceed. Stop if the YAML is invalid — ask to fix it, or confirm proceeding with an additive rewrite that can't be fully validated.
2. Determine the requested provider against `references/providers.md`'s table. Use it directly if named explicitly; if described generically ("Datenbank", "Cache", "Message Queue"), map it and confirm with the user if more than one provider could plausibly fit. Never hardcode a provider's fields from memory.
3. If a service with this same `provider.name` already exists: check that provider's "Team singleton" property first — if true (e.g. `virtual-k8s`), only editing the existing instance is possible, say so rather than presenting a choice. Otherwise name the existing instance(s) and ask edit-existing vs. add-a-parallel-instance; if more than one existing instance is found, ask which one "existing" refers to before that.
4. Gather configuration. If `references/interview-<name>.md` exists for this provider, follow its curated question order/defaults instead of improvising from the raw schema; otherwise derive the interview directly from `references/provider-<name>.md`. Either way, cover:
   - **Base config/secrets/plan** — per field: use what the user already stated; mention (don't silently apply) a default that materially affects capacity/behavior (storage size, replicas); quietly apply a low-stakes default (e.g. a version string); ask when there's no sensible default (e.g. the service name). Secret fields always become `${{ vault.NAME }}`.
   - **Backups**, only if the provider's property table supports them — ask once, opt-in. If yes, source the S3 target: default to a new dedicated Object Storage service (`backup-<serviceName>`), or reuse an existing `s3` service found in step 1, or take external S3 credentials directly from the user. The exact `backups:` shape, and which of `accessKey`/`secretKey` is plain `config` vs. a vault `secrets` entry, comes from `references/ci-pipeline.md` and the provider's own file — follow it exactly, don't force both into the vault.
   - **Provider-specific extras** beyond the base schema (e.g. PostgreSQL's Extensions) — these are **not** a `ci.yml` field. Note what was requested for the Step 7 summary as a post-provisioning follow-up (e.g. the `CREATE EXTENSION` statements to run once reachable) — never invent a config key for them, and never inject that logic into some other service's own `steps`.
5. Resolve the schema-version field name: brand-new file → `schemaVersion: v0.4` + `provider.schemaVersion`; existing file already on `v0.2`/`v0.3` → `provider.version` to match; existing file already on `v0.4` → `provider.schemaVersion`.
6. Write `ci.yml`, purely additively: a new file gets `prepare`/`test` as explicit empty `steps` plus the new block; an existing file gets only the new/edited block — every other line stays byte-for-byte identical (verify no unrelated diff before writing). If backups used a new dedicated Object Storage instance, that's a **second** new `run.<serviceName>` block, not folded into the first.
7. Summarize: provider added/edited and its service name, the resolved config, backups sourcing if used (and the second service name/bucket if a dedicated instance was created), every `${{ vault.* }}` reference still needing a real value, and a reminder that nothing is deployed — only `ci.yml` changed.

## Keep in mind

- Never hardcode a provider's field names/defaults from memory — always read them from `references/providers.md` / `references/provider-<name>.md`. If `codesphere` can't be located at all, stop and say so.
- Stay purely additive/non-destructive: never touch, reorder, or remove any service/stage/field other than the one being added or edited. There's no "overwrite" option here — only "add" or "edit the one instance the user pointed at."
- Never silently upgrade an existing file's top-level `schemaVersion` as a side effect of adding one service.
- Every secret field becomes `${{ vault.NAME }}` — never plaintext, even if the user pasted a real credential into the conversation; tell them the literal value still needs to be set through the vault separately.
- Ask only for what isn't already stated or safely defaultable — don't re-ask, and offer documented defaults as a skippable option rather than silently applying anything that affects capacity or behavior.
- Reject a config value the provider's own reference file marks invalid/reserved (e.g. PostgreSQL's `userName: postgres`) — check that specific provider's Common Pitfalls; don't assume one provider's rules apply to another.
- Never perform an active deployment or a real vault-store call — the deliverable is `ci.yml` only, unless the user explicitly and separately asks to actually store a secret.
- **Hard stops:** a `v0.1`-shaped file (can't host a Managed Service — ask how to proceed, don't migrate unprompted); invalid YAML (ask to fix, or proceed with an unvalidatable additive rewrite); `codesphere` skill not found; a requested provider not in `references/providers.md`'s catalog (say so plainly, don't invent a config, suggest `GET /managed-services/providers` in case the reference is stale); the user insisting on a plaintext secret (refuse, offer the vault-reference alternative instead).

## Adding a new provider

When Codesphere adds a new managed-service provider: add `references/provider-<name>.md` to `codesphere` plus a row in its `references/providers.md` and `references/ci-pipeline.md` tables — **don't write a new `codesphere-add-<provider>` skill.** This skill's steps 2–4 already read the provider table and matching reference file generically, so a new provider needs no change here at all. Optionally, once a provider's raw-schema interview has been used a few times and its natural question order is clear, add `references/interview-<name>.md` (format: `references/interview-postgresql.md`) — step 4 picks it up automatically if present; not having one yet is fine, it just falls back to the raw schema tables.

## Related

- `codesphere` — reference knowledge this skill reads from (loose coupling, read-only, never auto-invoked): `references/providers.md`, the matching `references/provider-<name>.md`, `references/ci-pipeline.md`, `references/secret-management.md`
- `codesphere-create-cluster-deployment` / `-container-deployment` / `-reactive-deployment` — generate a whole deployment and already detect managed-service candidates during their own workflow; this skill is the narrower tool for "just add/edit one managed service," standalone or as a manual follow-up
