# DocumentDB (FerretDB) — Interview Guide

> **Companion to:** `provider-documentdb.md`. If the two disagree, the schema file wins.

> **Used by:** `codesphere-add-managed-service` Phase 3, when the target provider is `ferretdb` (the user will likely say "documentdb" or "mongo(db)" — the provider name itself is `ferretdb`, confirm that mapping in Phase 1, don't silently assume the user knows this).

## Purpose

A two-component service (PostgreSQL + DocumentDB extension underneath, a FerretDB proxy in front speaking the MongoDB wire protocol) — the plan has two separate resource groups (Postgres side, FerretDB proxy side), which is the one thing genuinely unusual to walk a user through compared to a single-component provider.

## Question flow

### 1. Service name
Same as Postgres.

### 2. Plan — two resource groups, not one
Default silently: `cpu: 1` / `memory: 512` (Postgres side), `ferretdbCpu: 1` / `ferretdbMemory: 64` (proxy side). **Ask about `storage`** (default `1024`, and unlike Postgres this one has `min: 1024` too — no smaller option). Only mention the FerretDB-side resources exist at all if the user asks about performance/scaling — for the overwhelming majority of requests the proxy defaults are fine and don't need surfacing as a decision.

### 3. Stack version
At time of writing `provider-documentdb.md` documents exactly **one** allowed value for `config.version` (the combined Postgres/DocumentDB/FerretDB stack version) — there is no real choice to offer here yet. Confirm against the live reference file in case that's changed, but don't manufacture a question out of a single-option field.

### 4. Secrets
One secret: `superuserPassword` for the administrative user. Suggest a vault key like `<serviceName>SuperuserPassword`.

### 5. Compatibility caveat — surface this proactively
FerretDB does **not** support the full MongoDB command set — change streams and multi-document transactions are notably absent. If the user's request mentions either of those (or just generally "do I need to know anything"), mention this explicitly rather than letting them discover it later. This isn't a question to ask, it's a heads-up to volunteer.

### 6. Backups
Same mechanism as Postgres underneath (physical PostgreSQL backups, not `mongodump`-compatible) — mention this distinction plainly if the user opts into backups, since someone coming from a MongoDB background might otherwise assume `mongodump`/`mongorestore` compatibility that doesn't exist here.

### 7. Extensions / plugins
**Not applicable** in the Postgres-extension sense — FerretDB's compatibility surface is fixed by its own version, not something the user configures per-instance.

## What NOT to ask

- Anything implying real MongoDB feature parity (change streams, transactions) — these aren't configurable options, they're just unsupported. Don't present them as a choice.
- FerretDB-side resource tuning — not worth surfacing unless the user asks.
- Extensions — not a concept for this provider.

## Defaults quick-reference

| Field | Default | Ask explicitly? |
|---|---|---|
| `plan.parameters.cpu` (Postgres side) | `1` | No |
| `plan.parameters.memory` (Postgres side) | `512` | No |
| `plan.parameters.storage` | `1024` (min `1024`) | **Yes** |
| `plan.parameters.ferretdbCpu` | `1` | No, unless performance concern signaled |
| `plan.parameters.ferretdbMemory` | `64` | No, unless performance concern signaled |
| `config.version` | current single documented value | No — only one option exists today |
| Backups | disabled | **Yes** (opt-in; mention it's Postgres-format, not `mongodump`) |
| MongoDB compatibility gaps | — | Proactively mention if relevant, not a question |
