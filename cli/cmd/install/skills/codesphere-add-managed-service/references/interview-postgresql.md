# PostgreSQL — Interview Guide

> **Companion to:** `provider-postgresql.md` (the authoritative field schema — this file is opinionated *ordering and defaulting*, not a schema source in itself). If the two ever disagree on what a field is called or requires, `provider-postgresql.md` wins; update this file to match, not the other way around.

> **Used by:** `codesphere-add-managed-service` Phase 3, when the target provider is `postgres`. Not an official Codesphere doc — curated by us, kept in sync manually as `provider-postgresql.md` changes.

## Purpose

`provider-postgresql.md` says *which fields exist*. This file says *what order to ask them in, which ones to ask vs. silently default, and which ones to actively offer rather than bury*. The goal is a short, natural interview — not marching through every schema field as a separate question.

## Question flow

Ask in this order. Skip anything already stated in the user's original request.

### 1. Service name
The `run.<serviceName>` key. Suggest one from context if the user hasn't named it (e.g. `db`, `postgres`, `app-db` — match the app's own naming style if visible in the repo). Must not collide with an existing service name in the file. Always ask if genuinely ambiguous — this becomes part of the internal DNS name and is expensive to rename later (forces recreation).

### 2. Plan / storage
Default to `Small` (`id: 0`) silently — `cpu: 1`, `memory: 128` MiB are low-stakes defaults, mention them only in passing. **Storage is the one plan field worth asking about explicitly** (`storage`, default `1024` MiB, minimum `512`) — it's the field users most often have an actual opinion on and the one most annoying to have wrong later. Don't ask about `cpu`/`memory` unless the user signals a performance-sensitive workload themselves.

### 3. Engine version
Default silently to the current documented version in `provider-postgresql.md`'s Config Schema table (re-read it live, don't hardcode a number here — it drifts). Only ask if the user signals a specific compatibility requirement (an extension needing a minimum version, matching an existing environment, etc.).

### 4. `userName` / `databaseName`
Default both to `app` silently unless the user names their project/app — if they do, offer that as the suggested `databaseName` rather than the generic default. **Always reject `userName: postgres`** (reserved) the moment it comes up, don't wait until Phase 4's validation to catch it.

### 5. Secrets (vault key naming)
Don't ask what the password should *be* — that's never appropriate (see the skill's own Hard Gate on plaintext secrets). Ask or suggest the **vault key names**: a natural pattern is `<serviceName>UserPassword` / `<serviceName>SuperuserPassword` (e.g. `dbUserPassword`, `dbSuperuserPassword` for a service named `db`). Confirm the user understands these need real values populated through the vault before first sync — don't assume that's obvious.

### 6. Backups
One bundled opt-in question, not several separate ones: "Backups aktivieren? Braucht einen regionalen S3-Endpoint (nicht bucket-spezifisch), einen Zielpfad, ein Aufbewahrungsintervall (`intervalH`) und eine Retention-Dauer (`deleteRetentionDays`), sowie ein Access-Key-Paar." If yes, gather all of it in one pass rather than field-by-field ping-pong. If the `ci.yml` already has an `s3`/Object Storage managed service, mention it and offer to reuse that target instead of asking for an external one.

### 7. Extensions
This is **not a `ci.yml` config field** — extensions are enabled post-provisioning via SQL (`CREATE EXTENSION IF NOT EXISTS "..."`), so this question's answer goes into the Phase 6 summary as a follow-up action, never into `config:`. Don't dump the full 50+-entry list from `provider-postgresql.md` on the user — offer the commonly-relevant set and let them ask for more if needed:

- `postgis` — spatial/geographic data
- `vector` (pgvector) — embeddings/similarity search
- `pg_trgm` — fuzzy text search
- `uuid-ossp` — UUID generation
- `citext` — case-insensitive text
- `pgcrypto` — cryptographic functions
- `hstore` — key-value column type

Ask: "Brauchst du bestimmte Postgres-Extensions (z. B. PostGIS für Geodaten, pgvector für Embeddings)? Sonst lass ich das weg — lässt sich jederzeit später per SQL nachrüsten." If the user wants something not in this short list, check `provider-postgresql.md`'s full Extensions table before confirming it's actually available.

## What NOT to ask

- **HA / replica count** — PostgreSQL is always a single instance on Codesphere (no HA available at all). Don't offer this as a choice; it isn't one.
- **The full extension list upfront** — overwhelming and mostly irrelevant per-user; offer the short common set instead (see above).
- **CPU/memory tuning** — unless the user signals a performance concern themselves, these stay at the plan default silently.
- **Minor version pinning specifics** — the default documented version is fine for the overwhelming majority of requests; don't manufacture a decision point out of it.

## Defaults quick-reference

| Field | Default | Ask explicitly? |
|---|---|---|
| `plan.id` | `0` (Small) | No |
| `plan.parameters.cpu` | `1` | No |
| `plan.parameters.memory` | `128` | No |
| `plan.parameters.storage` | `1024` | **Yes** |
| `config.version` | current documented default | Only if compatibility need signaled |
| `config.userName` | `app` | No (unless project name suggests otherwise) |
| `config.databaseName` | `app` | No (unless project name suggests otherwise) |
| Backups | disabled | **Yes** (single bundled opt-in question) |
| Extensions | none | **Yes** (short common-set offer, not the full list) |
