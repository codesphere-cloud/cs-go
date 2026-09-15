# Object Storage (S3) — Interview Guide

> **Companion to:** `provider-object-storage.md`. If the two disagree, the schema file wins.

> **Used by:** `codesphere-add-managed-service` Phase 3, when the target provider is `s3`.

> **Structural note — this provider doesn't fit the usual config/secrets split cleanly.** `accessKey` is a **plain `config` field** (not a secret) despite being credential-shaped, while only `secretKey` is an actual `secrets` field. Don't reflexively vault-wrap `accessKey` just because it looks like a credential — follow the schema exactly as `provider-object-storage.md` documents it, config vs. secrets split included. This also means the "plan" here is a quota system (buckets/objects/throughput), not compute sizing (cpu/memory/storage) the way every other provider's plan works — don't ask about cpu/memory for this one, it doesn't have those fields at all.

## Purpose

Different shape from every other provider in the catalog: instead of provisioning compute, this creates an S3-compatible user + initial bucket with quota limits. The "plan" question is really "how much traffic/data do you expect," not a compute tier.

## Question flow

### 1. Service name
Same as other providers.

### 2. `accessKey` — generate, don't ask the user to invent one
Must be exactly 20 uppercase letters/digits, cluster-unique. Generate a valid one rather than asking the user to type a random 20-character string themselves — mention what was generated in the summary. This is a plain `config` field, not vault-wrapped (see structural note above).

### 3. `userDisplayName`
Optional, defaults to `My S3 User` — only ask if the user cares about a friendlier label; otherwise apply the default silently.

### 4. `initialBucketName`
Ask — this is genuinely user-specific and cluster-unique. Mention explicitly: if the name is already taken cluster-wide, the *service* still gets created but the *bucket* silently isn't — worth flagging as a known sharp edge in the summary, not just at request time.

### 5. `secretKey`
The one real secret here — exactly 40 alphanumeric characters, generated (never ask the user to type a raw secret), always ends up as `${{ vault.NAME }}`. Suggest a vault key like `<serviceName>SecretKey`.

### 6. Quota plan — ask only if the user signals scale concerns
Default the whole `Generic` plan's quota fields (`maxBuckets: 50`, `maxObjects: 100000`, `maxSizeKb: 10000000`, `maxReadOpsPerS`/`maxWriteOpsPerS: 1000`, `maxReadBytesPerS`/`maxWriteBytesPerS: 100000000`) silently for a typical request. Only surface individual quota fields if the user describes something the defaults clearly wouldn't fit (e.g. "we'll have thousands of buckets" or "high write throughput").

### 7. Backups
**Not offered — genuinely not available yet for this provider** (documented explicitly as "no backups yet" in `provider-object-storage.md`). Mention this proactively rather than letting the user assume it's covered the way Postgres backups are — object storage users backing up their own data is their responsibility until this changes.

### 8. Preview caveat
Mention once: this provider is preview, not enabled by default.

## What NOT to ask

- CPU/memory/storage sizing — this provider doesn't have those plan fields at all, it's quota-based.
- A user-chosen `accessKey` — generate a valid one instead of asking someone to hand-craft a 20-char uppercase string.
- Backups — explicitly not available; don't imply otherwise by asking.

## Defaults quick-reference

| Field | Default | Ask explicitly? |
|---|---|---|
| `config.accessKey` | generated (20 uppercase alphanumeric) | No — generate, don't ask user to invent |
| `config.userDisplayName` | `My S3 User` | Only if the user cares about the label |
| `config.initialBucketName` | — | **Yes** — genuinely user-specific |
| `secrets.secretKey` | generated (40 alphanumeric) | No — generate, always vault-wrapped |
| `plan.parameters.*` (quotas) | `Generic` plan defaults | Only if scale concerns are signaled |
| Backups | not available at all | Never ask — proactively mention the gap |
| Preview status | — | Mention once in the summary |
