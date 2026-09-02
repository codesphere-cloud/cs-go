# Virtual Kubernetes Cluster — Interview Guide

> **Companion to:** `provider-kubernetes.md`. If the two disagree, the schema file wins.

> **Used by:** `codesphere-add-managed-service` Phase 3, when the target provider is `virtual-k8s`.

> **Structural note — Phase 2's "replace vs. parallel" question does not apply to this provider.** `virtual-k8s` is a **team singleton**: a team can only ever have one instance, full stop. If Phase 0 finds an existing `virtual-k8s` service, skip straight to "edit the existing one" — do not offer "add a parallel instance" as a real option, it will be rejected by the platform. Say this plainly to the user rather than presenting a choice that doesn't actually exist.

## Purpose

The simplest config schema in the entire catalog — there isn't one. No `config:` fields, no `secrets:` fields documented at all. Access happens entirely through the injected kubeconfig, not through anything this interview would configure. The only real decisions are resource limits.

## Question flow

### 1. Service name
Same as other providers — though note this also becomes the one and only Kubernetes cluster the team will have, so a generic name (`k8s`, `cluster`) is more appropriate here than an app-specific one.

### 2. Team-singleton check (do this before anything else)
Check whether a `virtual-k8s` service already exists anywhere in the file, or — better, if reachable — via `GET /managed-services/providers`-style awareness that the team may already have one provisioned outside this repo entirely (a Landscape-embedded `virtual-k8s` isn't the only way one could already exist for the team). If there's any doubt, say so and ask the user to confirm rather than generating a second one that will fail.

### 3. Resource limits — the only real config
`plan.id: 0` (`Custom`, the only plan) with four adjustable fields: `cpu` (default `20`, range `20`–`160`), `memory` (default `5120`, range `5120`–`32768`), `storage` (default `20000`, range `20000`–`120000`), `ephemeralStorage` (default `30000`, range `30000`–`120000`). Default all four silently for a typical request; ask only if the user describes a workload that clearly needs more (e.g. "we're running several large services in the cluster").

### 4. No config, no secrets — say so explicitly
There's nothing to ask about beyond the plan. Don't manufacture questions about database credentials, versions, or anything else — this provider genuinely doesn't have those fields. Explain in the summary that access is via kubeconfig, injected automatically, not something configured here.

### 5. Preview caveat
Mention once: this provider is preview, not enabled by default.

### 6. What this actually unlocks — worth a one-line reminder
Adding this service doesn't deploy anything into the cluster by itself — it's the prerequisite for a Cloud Native Deployment. If the user's real goal is deploying a Helm chart, point them at `codesphere-create-cluster-deployment` instead of expecting this skill alone to get them there.

## What NOT to ask

- Anything under `config:` or `secrets:` — this provider has neither.
- "Replace vs. parallel" the normal way — team singleton, see the structural note above.
- Database/queue-style fields (version, credentials, storage in the data sense) — not applicable, this is compute, not a data service.

## Defaults quick-reference

| Field | Default | Ask explicitly? |
|---|---|---|
| `plan.parameters.cpu` | `20` | Only if a larger workload is signaled |
| `plan.parameters.memory` | `5120` | Only if a larger workload is signaled |
| `plan.parameters.storage` | `20000` | Only if a larger workload is signaled |
| `plan.parameters.ephemeralStorage` | `30000` | Only if a larger workload is signaled |
| `config:` / `secrets:` | none exist | Never ask |
| Existing instance found | — | Treat as "edit only" — team singleton, no parallel option |
| Preview status | — | Mention once in the summary |
