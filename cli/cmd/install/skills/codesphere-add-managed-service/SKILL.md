---
name: codesphere-add-managed-service
description: Adds a Codesphere managed service (PostgreSQL, Babelfish, DocumentDB/FerretDB, Valkey, RabbitMQ, OpenSearch, S3/Object Storage, Virtual Kubernetes Cluster, or any other provider documented in the codesphere reference set) to a ci.yml — creates ci.yml if none exists, or adds a provider block to an existing one. Guides the user through configuration but only asks for what isn't already stated or safely defaultable, reading the exact config/secrets/plan schema from the matching provider reference file rather than hardcoding it per provider. If a service of the requested provider type already exists, asks whether to replace/edit it or add a parallel second instance rather than picking silently. Trigger for "postgres/rabbitmq/valkey/object storage/... hinzufügen", "add a database/cache/queue/storage to my ci.yml", "managed service konfigurieren", "zweite <provider> instanz", or any request to add/configure a Codesphere managed service — regardless of which specific provider is named.
license: none
allowed-tools: Bash Read Write Glob Grep
metadata:
  version: "1.0.0"
  updated: "2026-07-29"
  cost-tier: medium
---

> **Process:** When this skill is explicitly/directly invoked by name, execute it immediately — don't ask the user what they want done with it. Proceed straight to Phase 0. The only prompts are the Decision Points and Blocking conditions defined within each phase below.

## When to use this

Trigger when the user wants a Codesphere managed service — any provider: PostgreSQL, Babelfish, DocumentDB/FerretDB, Valkey, RabbitMQ, OpenSearch, S3/Object Storage, Virtual Kubernetes Cluster, or a provider added to `codesphere`'s reference set later — added to or configured within a `ci.yml`. E.g. "füge postgres zu meinem ci.yml hinzu", "add a managed valkey cache", "ich brauch noch eine zweite rabbitmq-Instanz parallel", "s3-Bucket als Managed Service konfigurieren", "codesphere-add-managed-service ausführen". This is one generic skill covering every provider — **do not create a separate `codesphere-add-<provider>` skill per provider**; adding support for a new provider means adding its `provider-<name>.md` reference file to `codesphere`, not writing a new skill (see "Adding a new provider" below).

This is narrower than `codesphere-create-cluster-deployment` / `codesphere-create-container-deployment` / `codesphere-create-reactive-deployment`, which generate a whole deployment from a chart/Dockerfile/source — this skill only ever adds or edits one managed-service block within whatever `ci.yml` already exists (or creates a minimal one if none does).

## Hard Gate

- **Shared family conventions apply — see `references/skill-family-conventions.md`** (inside `codesphere`'s directory; this skill has no `references/` folder of its own — `Glob` for `**/codesphere/references/*.md` if the install path isn't already known). Covers locating/reading `codesphere`'s other `references/*.md` files (never requires `codesphere` itself to be loaded as an active skill) and the repo-root-only `ci.yml` placement used below. If `codesphere` can't be located at all, stop and tell the user it needs to be installed alongside this skill — don't guess at provider field content from memory.
- **MUST determine the provider from `references/providers.md`'s provider table before doing anything else** — never hardcode a provider's field names/defaults from memory, and never assume only the providers named in this file's own description are the complete list; `references/providers.md` is authoritative for which providers exist and which `references/provider-<name>.md` file documents each one.
- **MUST be purely additive/non-destructive to an existing `ci.yml`.** Adding or editing the requested service must never touch, reorder, or remove any other `run.<serviceName>` block, `prepare`/`test` step, or top-level field that isn't the one being added/edited. There is no overwrite option — only "add" or "edit the one existing service the user pointed at."
- **MUST NOT silently change an existing file's top-level `schemaVersion`.** If the file already has `schemaVersion: v0.2`/`v0.3`, add the new block using `provider.version` (matching that file's convention) — do not upgrade the whole file to `v0.4` as a side effect. Only a brand-new `ci.yml` (nothing existed before) defaults to `schemaVersion: v0.4` with `provider.schemaVersion`. A `v0.1` file (no `schemaVersion`, flat `run.steps`) cannot host a Managed Service at all in its current shape — see Phase 0's blocking condition.
- **MUST use `${{ vault.NAME }}` for every secret field the target provider's `references/provider-<name>.md` documents — never a plaintext value**, regardless of whether the user typed an actual credential into the conversation. If they did, treat the vault key name as what gets recorded in `ci.yml`, and tell them the literal value still needs to be set through the vault separately — never write the plaintext they gave you into the file.
- **MUST ask only for information the user hasn't already stated or that isn't safely defaultable** — don't re-ask something already given, and don't ask about genuinely optional fields without first offering the documented default as a "press enter to skip" style option.
- **MUST offer the replace-vs-parallel choice whenever an existing service of the same `provider.name` is found in the `ci.yml`** — never silently pick one on the user's behalf.
- **MUST reject any config value the provider's own reference file documents as invalid or reserved** (e.g. PostgreSQL's `userName: postgres`) before finalizing config — check the specific provider file's Common Pitfalls section, don't assume PostgreSQL's rules apply to every provider.
- **MUST NOT perform any active deployment.** No `cs start`, no `POST /workspaces/{id}/landscape/deploy`, no vault API calls to actually store secrets unless the user explicitly asks for that as a separate, clearly-flagged step. This skill's deliverable is `ci.yml` (new or edited) — nothing is deployed or synced.

## Process: 7-Phase Workflow

This section describes 7 phases (Phase 0–6) executed in strict linear order, except where a phase is explicitly skippable per its own Prerequisite.

**Phase [N]: [Action Title]**
- **Prerequisite:** What must be true before this phase
- **Blocker risk:** What could fail and why
- **Action:** What you do (commands, decision points)
- **Output:** What success looks like
- **Blocking conditions:** If X or Y, stop and inform user

### Phase 0: Locate and inspect `ci.yml`

**Prerequisite:** Skill invoked.

**Blocker risk:** Ambiguous repo root in a monorepo; a `v0.1`-shaped file with no `schemaVersion` and a flat `run.steps` that can't host a Managed Service in its current form.

**Action:**
1. Confirm the repository root — `ci.yml` always belongs there.
2. Check whether a `ci.yml` already exists at the root.
   - **No** → note that a new file will be created in Phase 5, defaulting to `schemaVersion: v0.4`. Continue to Phase 1.
   - **Yes** → read it. Note its `schemaVersion` (determines `provider.version` vs. `provider.schemaVersion` later). Note every existing `provider.name` service found, for Phase 2. **Separately, note any existing `provider.name: s3` service(s) specifically** — these matter for Phase 3b regardless of what provider is actually being added/edited this run, since an existing Object Storage service is a reuse candidate for a new backup target.

**Output:** Repo root confirmed; existing `ci.yml` (if any) parsed; list of existing managed-service blocks by provider name (for Phase 2); list of existing `s3` services specifically (for Phase 3b); target `schemaVersion` convention determined.

**Blocking conditions:**
- Existing `ci.yml` is syntactically invalid YAML → stop, ask the user to fix it or confirm proceeding by rewriting just the `run:` block additively without being able to fully validate the rest.
- Existing `ci.yml` is `v0.1`-shaped → stop. Explain it can't host a Managed Service block without first being migrated to `v0.2`+; ask whether the user wants that migration done (out of scope for this skill — point at `codesphere-create-reactive-deployment` or manual migration) or wants to proceed by hand.

### Phase 1: Determine the requested provider

**Prerequisite:** Phase 0 completed.

**Blocker risk:** Guessing the wrong provider from an ambiguous request (e.g. "cache" could mean Valkey, or in principle any other in-memory store the catalog later adds).

**Action:** If the user already named a specific provider or a Codesphere provider name directly (`postgres`, `babelfish`, `ferretdb`, `valkey`, `rabbitmq`, `opensearch`, `s3`, `virtual-k8s`, or any other listed in `references/providers.md`), use that. If they described a need in generic terms ("Datenbank", "Cache", "Message Queue", "Objektspeicher"), map it against `references/providers.md`'s provider table and confirm the match with the user rather than assuming silently if more than one provider could plausibly fit.

**Output:** Confirmed target provider name and its matching `references/provider-<name>.md` file.

**Blocking conditions:** Do not proceed to Phase 2 with an unconfirmed or ambiguous provider choice.

### Phase 2: Existing-instance decision (only if Phase 0 found an existing service of this provider)

**Prerequisite:** Phase 1 determined the provider; Phase 0 found at least one existing service with that same `provider.name`. Skip this phase entirely (continue straight to Phase 3) if none was found.

**Blocker risk:** Silently picking replace or parallel-add without asking, which either destroys a working config or produces an unwanted duplicate service; offering "add a parallel instance" for a provider that doesn't actually allow more than one, which the platform will reject.

**Action:** First check the provider's own property table in `references/provider-<name>.md` for **"Team singleton"**. If `true` (e.g. `virtual-k8s`), only Option A is real — say so plainly rather than presenting a choice that doesn't exist, and skip straight to editing the existing instance. If `false` (the common case), name the existing service(s) of this provider found (service key, plan, key config values) and ask:

- **Option A — "Bestehende Instanz bearbeiten"**: edit the fields of the existing service. Only the fields the user wants to change are touched; everything else about that service block stays as-is.
- **Option B — "Neue parallele Instanz hinzufügen"**: add a second, independent service of the same provider under a new service name. Proceeds like the no-existing-instance case from here, just alongside the first.

If more than one existing service of this provider is found, first ask which one is meant by "bestehende Instanz" before asking A/B.

**Output:** Confirmed target: edit an existing service (and which one), or add a new one — or, for a team-singleton provider, confirmed edit-only with no A/B question asked.

**Blocking conditions:** Do not proceed to Phase 3 without an explicit A/B answer for a non-singleton provider. For a team-singleton provider with an existing instance, do not proceed without confirming the user actually means to edit that one instance.

### Phase 3: Gather configuration from the provider's own reference file

**Prerequisite:** Phase 1 (and Phase 2, if applicable) completed.

**Blocker risk:** Asking for information already given re-annoys the user; silently defaulting something the user actually cares about (e.g. storage size) produces a config they didn't want; assuming a different provider's field names/defaults apply here; treating a post-provisioning action (e.g. Postgres extensions, enabled via SQL, not a `ci.yml` field) as if it belonged in `config:` and hallucinating a field that doesn't exist in the schema.

**Action:** First, check whether `references/interview-<name>.md` exists for the target provider (e.g. `references/interview-postgresql.md`). If it does, follow its curated question order/defaults/opt-in framing instead of improvising from the raw schema tables — it exists specifically to avoid re-deriving the same interview logic from scratch every time and to keep the questions natural rather than a field-by-field march through the schema. If no interview file exists yet for this provider, fall back to deriving the interview directly from `references/provider-<name>.md` per 3a–3c below; this is not a blocker, just the less-curated path. (Not every provider has an interview file yet — new ones get added incrementally.)

Whichever path is used, cover all three of the following areas — don't stop after the base schema, since backups and provider-specific extras are exactly what tends to get skipped:

**3a — Base config/secrets/plan.** Read `references/provider-<name>.md`'s Config Schema, Secrets Schema, and Plan tables — these define exactly which fields exist, which are required, and their documented defaults for *this* provider. For each field:
- Already stated by the user → use it directly, don't ask again.
- Not stated, has a documented default → mention the default in passing rather than silently applying it if the field materially affects capacity/behavior (e.g. storage size, replica count); apply quietly if it's a low-stakes default (e.g. a version string picking the current documented version).
- Not stated, no sensible default (e.g. the service name / `run.<serviceName>` key) → ask.
- Secret fields → always end up as `${{ vault.NAME }}`; either use a vault key name the user gave, or generate a sensible one from the service name.

**3b — Backups (if the provider supports them).** Check the provider's own property table (the "Backups / PITR" row near the top of `references/provider-<name>.md`) — not every provider supports this (e.g. Valkey/RabbitMQ/OpenSearch currently don't, per their reference files). If supported, ask once, opt-in: "Backups aktivieren?" Don't configure backups unless the user opts in.

If the user opts in, the `backups:` block needs an S3 target — don't assume the user already has one lying around externally. Offer three ways to source it, in this priority order:

- **Option A — dedicated new Object Storage instance (default suggestion).** Create a new `provider.name: s3` service specifically for this backup target, named `backup-<serviceName>` (e.g. `backup-app-db` for a database service named `app-db`) unless the user prefers a different name. Gather its fields using `references/interview-object-storage.md`'s guidance (generate `accessKey`/`secretKey`, suggest `initialBucketName` like `<serviceName>-backups`) — this is the same Object Storage config this skill would gather if `s3` were the primary provider being added, just condensed into this sub-flow rather than a separate full pass through Phase 1. This is the recommended default because it needs nothing external and keeps the backup target's lifecycle visible in the same `ci.yml`.
- **Option B — reuse an existing `s3` service.** Only offer this if Phase 0 found one already in the file. Ask whether to point the new backup target at that service's bucket instead of creating another one.
- **Option C — external/self-managed S3.** If the user already has an S3-compatible target elsewhere (AWS, another cloud, self-hosted), gather `endpointUrl` (regional, not bucket-specific), `destinationPath`, `accessKey`, and `secretKey` directly from them.

Whichever option is chosen, the `backups:` block's exact shape (`enabled`, `intervalH`, `deleteRetentionDays`, `config.endpointUrl`, `config.destinationPath`, `config.accessKey`, `secrets.secretKey`) is documented in `references/ci-pipeline.md`'s Managed Service field reference and mirrored per-provider in `references/provider-<name>.md` — follow that shape exactly, including which of `accessKey`/`secretKey` is plain `config` vs. `secrets` per the schema (don't force both into the vault just because one of them is a credential — only the one the schema actually places under `secrets` gets `${{ vault.NAME }}`). For Option A specifically, the new Object Storage service's own `url` output (always `http://rgw-load-balancer.rook-ceph.svc.cluster.local` for Codesphere-managed S3, per `references/provider-object-storage.md`) becomes the backup's `config.endpointUrl`, and its `initialBucketName` feeds `config.destinationPath` (`s3://<bucket>/`).

**3c — Provider-specific extras beyond the base schema.** Read the rest of `references/provider-<name>.md` (not just the Config/Secrets/Plan tables) for anything else worth surfacing to the user — the concrete example today is PostgreSQL's **Extensions** section (PostGIS, pgvector, pg_trgm, etc.). Ask which extensions the user wants. **These are NOT a `ci.yml` config field — do not invent one.** Per the provider's own docs, extensions are enabled post-provisioning via SQL (`CREATE EXTENSION IF NOT EXISTS "..."`). This skill does not edit some other, unrelated service's `prepare`/`run` steps to inject that SQL — that would reach outside the one managed-service block this skill is scoped to touch. Instead, record which extensions were requested and hand them to the user as an explicit follow-up in Phase 6's summary (the exact `CREATE EXTENSION` statements to run once the database is provisioned and reachable). If a future request needs this automated into an application service's own steps, that's a separate, explicit ask against that service — not something to do silently here.

**Output:** Complete field set for the new/edited service (base config/secrets/plan, backups if opted into, and any provider-specific extras noted for the summary), with every value either user-stated or an explicitly-mentioned default.

**Blocking conditions:** Do not proceed to Phase 4 with a config value the provider's reference file documents as invalid/reserved still set — reject and re-ask that one field specifically.

### Phase 4: Resolve schema-version-dependent field names

**Prerequisite:** Phase 3 completed; Phase 0's target `schemaVersion` convention known.

**Blocker risk:** Using `provider.schemaVersion` against a file that's still on `v0.2`/`v0.3` (or vice versa) produces a block the target Codesphere instance may reject.

**Action:** Per the Load-bearing correction in `codesphere`: for a brand-new file, use `schemaVersion: v0.4` and `provider.schemaVersion`. For an existing file already on `v0.2`/`v0.3`, use `provider.version` to match. For an existing file already on `v0.4`, use `provider.schemaVersion`.

**Output:** Correct field name selected for the `provider:` block about to be written.

**Blocking conditions:** None.

### Phase 5: Write `ci.yml`

**Prerequisite:** Phase 4 completed.

**Action:**
- **New file:** `schemaVersion: v0.4`, `prepare: { steps: [] }`, `test: { steps: [] }` (explicit, even empty), and the new `run.<serviceName>` block per `references/ci-pipeline.md`'s Managed Service field reference and the target provider's own config/secrets schema.
- **Existing file, adding a new service:** insert the new `run.<serviceName>` block alongside the existing ones — every other line of the file stays byte-for-byte identical.
- **Existing file, editing an existing service:** modify only the fields the user chose to change within that one service's block — everything else in the file, including that same block's untouched fields, stays as-is.
- **If Phase 3b's backups were opted into:** include the `backups:` block on the same service, per the exact shape in `references/ci-pipeline.md` and the provider's own reference file — this is a sibling field to `provider`/`plan`/`config`/`secrets` on the same `run.<serviceName>` entry, not a separate service. **If Option A (dedicated new Object Storage instance) was chosen:** also write that as its own additional `run.<serviceName>` block (e.g. `backup-app-db`) with `provider.name: s3` — two new services get written this run, not one; don't drop the backup-target service just because the user's original request was about a different provider.

**Output:** `ci.yml` written (new or edited) at the repository root.

**Blocking conditions:** Before finishing, verify no other part of the file changed beyond the intended addition/edit — if a diff would show unrelated changes, stop and fix before writing.

### Phase 6: Summary

**Prerequisite:** Phase 5 completed.

**Action:** Briefly summarize for the user:
- Which provider was added/edited, whether a new `ci.yml` was created or an existing one was edited, and which service name holds the block.
- The resolved config (plan, key config fields).
- **If backups were enabled: which S3 option was used (new dedicated instance / reused existing / external) and, for a new dedicated instance, its service name and bucket** — call this out as a second new service, not a buried detail of the first.
- Every `${{ vault.* }}` reference that still needs a real value populated before the first sync (name each one explicitly) — this now includes the backup target's own `secretKey` if a new Object Storage instance was created.
- A reminder that this only changes `ci.yml` — nothing is deployed; the next Landscape sync/deploy is what actually provisions the service(s).

**Output:** One consolidated status message the user can act on.

**Blocking conditions:** None.

---

## Hard Blockers (Must Stop)

1. **`ci.yml` is `v0.1`-shaped (no `schemaVersion`, flat `run.steps`)**
   - Cannot host a Managed Service block without first migrating to `v0.2`+. Ask the user how they want to proceed; don't migrate the file unprompted as part of "just adding a service."
2. **Existing `ci.yml` is syntactically invalid YAML**
   - Stop. Ask the user to fix it or confirm proceeding with an additive rewrite that can't be fully validated against the rest of the file.
3. **`codesphere` skill can't be located**
   - Stop. This skill has no reference content of its own and cannot safely generate any provider's config from memory alone.
4. **Requested provider isn't in `references/providers.md`'s catalog**
   - Stop and say so plainly — don't invent a plausible-looking config for a provider that isn't documented. Suggest checking `GET /managed-services/providers` on the target instance in case the catalog reference is out of date.
5. **User insists on a plaintext secret in the file**
   - Stop and explain why that's not done here (see Hard Gate) — offer the vault-reference alternative instead of complying.

## Adding a new provider

When Codesphere adds a new managed-service provider (or `codesphere`'s reference set gains one it didn't have before): add `references/provider-<name>.md` to `codesphere` and a row to its provider table in `references/providers.md` and `references/ci-pipeline.md`'s routing table — **do not write a new `codesphere-add-<provider>` skill.** This skill's Phase 1–3 already read the provider table and the matching reference file generically; a new provider needs no change to this skill's own logic at all, only to `codesphere`'s reference set.

Optionally, once the raw-schema interview for a provider has been used a few times and its natural question order/defaults are clear, add `references/interview-<name>.md` (see `references/interview-postgresql.md` for the format) — Phase 3 picks it up automatically if present, with no code/logic change needed here either. Not having one yet is fine; Phase 3 falls back to deriving the interview from the raw schema tables.

## Related

- `codesphere` — reference knowledge this skill reads from (loose coupling, read-only, never auto-invoked); specifically `references/providers.md` (provider catalog/lookup), the matching `references/provider-<name>.md` per request, `references/ci-pipeline.md`, `references/secret-management.md`
- `codesphere-create-cluster-deployment` / `codesphere-create-container-deployment` / `codesphere-create-reactive-deployment` — generate a whole deployment; each of them already detects database/cache/queue-shaped components as managed-service candidates during their own workflow. This skill is the narrower tool for "just add/edit a managed service," whether standalone or as a manual follow-up after one of those three.
