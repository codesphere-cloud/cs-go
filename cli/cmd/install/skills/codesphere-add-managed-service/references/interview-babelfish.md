# Babelfish — Interview Guide

> **Companion to:** `provider-babelfish.md`. If the two disagree, the schema file wins.

> **Used by:** `codesphere-add-managed-service` Phase 3, when the target provider is `babelfish`.

## Purpose

Babelfish is a PostgreSQL server with a SQL Server-compatible TDS endpoint bolted on — same underlying platform as `postgres`, so this interview mirrors that one's shape closely, but the config schema itself is much smaller (no `userName`/`databaseName` fields at all — only `version`).

## Question flow

### 1. Service name
Same as Postgres: ask if not stated, suggest from context, must not collide with existing service names.

### 2. Plan / storage
Default `Small` (`id: 0`) silently — `cpu: 1`, `memory: 128`. **Ask about `storage`** (default `1024`, minimum `512`) — same reasoning as Postgres, it's the one capacity field users actually have opinions on.

### 3. Combined PostgreSQL + Babelfish version
Default silently to the current documented default in `provider-babelfish.md`'s Config Schema (re-read live, don't hardcode). Only ask if the user has a specific SQL Server compatibility requirement tied to a particular Babelfish version.

### 4. Secrets
Only one secret here: `superuserPassword` — there's no separate application-user password the way Postgres has `userName`/`userPassword`. Suggest a vault key like `<serviceName>SuperuserPassword`.

### 5. Preview status
Mention once, don't dwell: Babelfish is a **preview** feature and must be enabled by the operator on the target installation before it can actually be used — flag this in the Phase 6 summary as a heads-up, not a blocker for generating the `ci.yml` itself.

### 6. Backups
Same mechanism as Postgres (Barman Cloud, physical backups) since it's the same underlying platform — same bundled opt-in question as Postgres's interview guide.

### 7. Extensions / plugins
**Not applicable.** Babelfish's value is the TDS/T-SQL compatibility layer itself, not a curated extension set the way Postgres has PostGIS/pgvector/etc. — don't ask this question for Babelfish.

## What NOT to ask

- `userName` / `databaseName` — these don't exist in Babelfish's config schema at all, unlike Postgres. Don't offer them.
- HA / replicas — same as Postgres, always a single instance, not a choice.
- Extensions — not a documented concept for this provider (see above).

## Defaults quick-reference

| Field | Default | Ask explicitly? |
|---|---|---|
| `plan.id` | `0` (Small) | No |
| `plan.parameters.cpu` | `1` | No |
| `plan.parameters.memory` | `128` | No |
| `plan.parameters.storage` | `1024` | **Yes** |
| `config.version` (combined PG+Babelfish) | current documented default | Only if compatibility need signaled |
| Backups | disabled | **Yes** (bundled opt-in, same as Postgres) |
| Preview-status heads-up | — | Mention once in the summary, not a question |
