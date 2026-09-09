# OpenSearch — Interview Guide

> **Companion to:** `provider-opensearch.md`. If the two disagree, the schema file wins.

> **Used by:** `codesphere-add-managed-service` Phase 3, when the target provider is `opensearch`.

## Purpose

Four plan tiers (more than any other provider in the catalog — `Small`/`Medium`/`Large`/`Extra Large`), a fixed `replicas: 3` on the example plan, and two exposed endpoints (REST API + Dashboards UI) worth explaining rather than just configuring silently.

## Question flow

### 1. Service name
Same as other providers.

### 2. Plan
Ask which of the four tiers — the spread here is wider than other providers (dev-scale up to a genuinely large search/analytics cluster), so don't silently assume `Small` for anything the user describes as more than a toy/test index. `storage` (default `5120`, minimum `5120`) is worth surfacing alongside the tier choice since it's meaningfully larger than other providers' defaults.

### 3. Version
Single documented allowed value (`2.19`) at time of writing — don't ask, just note it's applied.

### 4. Secrets
One secret: `superuserPassword`. Suggest a vault key like `<serviceName>SuperuserPassword`.

### 5. Two access surfaces — mention both
This provider exposes both a REST API (`host`/`port`, for application/ingest traffic) and an OpenSearch Dashboards URL (`dashboardsUrl`, for humans). Mention both in the Phase 6 summary — a user thinking of this purely as "a database my app talks to" may not realize there's also a browsable UI endpoint they'll want to know about.

### 6. Availability caveat
Closed testing, dedicated installations only — same caveat pattern as Valkey/RabbitMQ.

### 7. Backups
**Not offered.** No backup capability documented for this provider.

### 8. Extensions / plugins
Not offered via `codesphere-add-managed-service` — OpenSearch has its own plugin ecosystem, but nothing in `provider-opensearch.md` documents a `ci.yml`-level config field for it. If asked, say so rather than guessing.

## What NOT to ask

- Backups — not supported.
- Replica count — fixed at `3` on the documented example plan, not exposed as a separately-adjustable field; don't offer it as a choice unless the live schema shows otherwise.
- Plugin configuration — not exposed via a documented `ci.yml` field.

## Defaults quick-reference

| Field | Default | Ask explicitly? |
|---|---|---|
| `plan.id` (Small/Medium/Large/Extra Large) | — | **Yes** |
| `plan.parameters.storage` | `5120` | Mention alongside the tier choice |
| `config.version` | `2.19` (current single documented value) | No |
| Backups | not available | Never ask |
| Dashboards endpoint | — | Mention proactively in the summary, not a question |
| Closed-testing/dedicated-only caveat | — | Mention proactively if relevant |
