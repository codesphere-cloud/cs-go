# Valkey — Interview Guide

> **Companion to:** `provider-valkey.md`. If the two disagree, the schema file wins. **Note:** this file's own reference set flags the `provider.version`/`schemaVersion` value as unresolved between two conflicting sources (`v0` vs. `v1`) — verify against `GET /managed-services/providers` before finalizing, don't silently pick one.

> **Used by:** `codesphere-add-managed-service` Phase 3, when the target provider is `valkey` (the user may say "redis" — Valkey is the Redis-protocol-compatible provider on Codesphere, there is no `redis`-named provider; confirm this mapping in Phase 1).

## Purpose

The simplest schema in the catalog — one config field (`version`, currently a single allowed value), one secret, no backups, no HA. The interview is correspondingly short; resist the urge to manufacture questions where the schema doesn't have real decisions to offer.

## Question flow

### 1. Service name
Same as other providers.

### 2. Plan
Three tiers exist (`Small`/`Medium`/`Large`, `id: 0`/`1`/`2`) — ask which, don't silently assume `Small` the way single-tier providers default quietly, since the tier *is* the main decision here (there's no separate `storage` field to fine-tune within a tier the way Postgres has). Mention what each tier roughly means in terms of use case (caching for a small app vs. a larger shared cache) if the user seems unsure rather than just listing IDs.

### 3. Version
Only one documented allowed value at time of writing — don't ask, just note it's applied.

### 4. Secrets
One secret: `superuserPassword`. Suggest a vault key like `<serviceName>SuperuserPassword`.

### 5. Availability caveat
Valkey is **closed testing**, available only on dedicated (self-hosted) installations, not shared cloud. Mention this once if the user's target instance isn't already known to be a dedicated installation — better to surface it now than have the `ci.yml` fail to deploy.

### 6. Backups
**Not offered.** No backup capability is documented for this provider — don't ask, don't include a `backups:` block.

### 7. Extensions / plugins
**Not applicable** — no such concept for this provider.

## What NOT to ask

- Backups — not supported, don't ask.
- Extensions — not a concept here.
- Persistence/durability tuning — not exposed via any documented config field; if the user asks about it, say it isn't configurable per the current schema rather than guessing at a field name.

## Defaults quick-reference

| Field | Default | Ask explicitly? |
|---|---|---|
| `plan.id` (Small/Medium/Large) | — | **Yes** — no safe silent default, this is the main decision |
| `config.version` | current single documented value | No |
| Backups | not available | Never ask |
| Closed-testing/dedicated-only caveat | — | Mention proactively if relevant |
