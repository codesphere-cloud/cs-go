# RabbitMQ — Interview Guide

> **Companion to:** `provider-rabbitmq.md`. If the two disagree, the schema file wins.

> **Used by:** `codesphere-add-managed-service` Phase 3, when the target provider is `rabbitmq`.

## Purpose

Similar shape to Valkey — small schema, three plan tiers, one secret — but with a `replicas` plan parameter that's actually worth surfacing, since it affects message durability/availability in a way a cache tier doesn't.

## Question flow

### 1. Service name
Same as other providers.

### 2. Plan, including replicas
Ask which tier (`Small`/`Medium`/`Large`). Unlike Valkey, also surface `replicas` explicitly if the user's use case sounds like it matters (anything described as important/production messaging) — the `Small` example plan defaults to `replicas: 1`, i.e. no redundancy, which is worth a deliberate choice rather than a silent default for anything beyond a dev/test queue.

### 3. Version
Single documented allowed value at time of writing — don't ask, just note it's applied.

### 4. Secrets and the fixed admin username
One secret: `superuserPassword`. **The username is always `admin`** — not configurable, don't ask about it, just mention it's fixed when explaining how to connect. Suggest a vault key like `<serviceName>SuperuserPassword`.

### 5. Availability caveat
Closed testing, dedicated installations only — same caveat pattern as Valkey, mention once if relevant to the target instance.

### 6. Backups
**Not offered.** No backup capability documented for this provider.

### 7. Extensions / plugins
Not offered via `codesphere-add-managed-service` — RabbitMQ has its own plugin system (management UI, shovel, etc.) but nothing in `provider-rabbitmq.md` documents a `ci.yml`-level way to enable plugins. If the user asks, say this isn't currently exposed via a managed-service config field rather than guessing at one.

## What NOT to ask

- The admin username — fixed at `admin`, not a field.
- Backups — not supported.
- Plugin configuration — not exposed via a documented `ci.yml` field.

## Defaults quick-reference

| Field | Default | Ask explicitly? |
|---|---|---|
| `plan.id` (Small/Medium/Large) | — | **Yes** |
| `plan.parameters.replicas` | `1` (Small example) | Ask if the use case sounds production-relevant, else default silently |
| `config.version` | current single documented value | No |
| Admin username | `admin`, fixed | Never ask — just inform |
| Backups | not available | Never ask |
| Closed-testing/dedicated-only caveat | — | Mention proactively if relevant |
