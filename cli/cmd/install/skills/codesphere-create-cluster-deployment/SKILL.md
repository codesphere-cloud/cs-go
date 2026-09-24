---
name: codesphere-create-cluster-deployment
description: 'Generates a ci.yml for migrating an existing Helm chart onto Codesphere: decomposes it into a Landscape (Reactive/Managed Container/Managed Service per component) by default, falling back to a Cloud Native Deployment (Virtual Kubernetes Cluster) only where a component genuinely needs Kubernetes semantics. Only run on explicit invocation. Requires an actual Helm chart — for Dockerfiles/docker-compose.yml use codesphere-create-container-deployment, for application source only use codesphere-create-reactive-deployment.'
license: none
allowed-tools: Bash Read Write Glob Grep
metadata:
  version: "2.1.0"
  updated: "2026-09-22"
  cost-tier: high
---

> **Process:** When this skill is explicitly/directly invoked by name, execute it immediately — don't ask the user what they want done with it.

## When to use this

Trigger when the user wants a `ci.yml` generated from an existing Helm chart — "Helm-Chart auf Codesphere migrieren", "ci.yml aus meinem Helm-Chart erstellen", "codesphere-create-cluster-deployment ausführen", or as a suggested next step when a user has a Helm chart and no `ci.yml` yet. No Helm chart at all → wrong entry point: use `codesphere-create-container-deployment` (Dockerfiles/`docker-compose.yml`) or `codesphere-create-reactive-deployment` (application source only).

## Reference

Read `references/migration-guide.md`'s **"Helm Chart → Codesphere Landscape"** and **"Helm Chart → Cloud Native Deployment"** sections before generating anything — they live in the sibling `codesphere` skill (see `references/skill-family-conventions.md` for how to locate/read them without loading `codesphere` itself). That reference has the decision tree, the per-component Managed-Service/Reactive/Managed-Container classification tables, and two full worked examples (a mixed Landscape, and the Cloud Native fallback). This skill doesn't repeat them — it's the process wrapper, not the schema.

## Process

1. `ci.yml` always goes at the repo root, even in a monorepo with multiple charts. If one already exists, ask whether to **update** (keep existing services, add/replace this one) or **overwrite**.
2. Find the chart (`Chart.yaml`). None found → stop and say so; don't generate anything. Treat each subchart of an umbrella chart (`dependencies:` in `Chart.yaml`) as its own component, not one blob.
3. Per component, check for a Managed Service equivalent (DB/cache/queue) against the migration guide's table. Present each match to the user individually — component, proposed service, what changes (data moves out of the cluster; this skill doesn't migrate existing data) — and let them confirm replace-or-keep. Nothing swaps silently.
4. Assess what's left once confirmed replacements are set aside. Trivial (one Deployment+Service, maybe Ingress/ConfigMap/Secret, no CRDs/operators/StatefulSets) → default to a Landscape, but confirm with the user first, naming the tradeoff: Cloud Native keeps the chart as-is but loses automatic off-when-unused/monitoring/router integration, and shifts things like confidential-compute proof onto them. Non-trivial remainder → Cloud Native Deployment, no need to ask.
5. **Landscape branch:** classify every remaining component as Reactive (own Dockerfile/code in the repo) or Managed Container (unmodified vendor image, nothing to rebuild) per the migration guide's table; ask directly only when the signal is genuinely ambiguous. Present the whole batch as one confirmation pass, not a question per component.
6. **Cloud Native branch:** for every confirmed Managed Service replacement, work out what the chart itself needs — disable the old in-cluster component (`condition: false` in `values.yaml`, or a `--set` override on the `helm install`/`upgrade` line) and rewire whatever pointed at its old Service DNS to the Managed Service via a `${{ vault.NAME }}` secret, never hardcoded. Tell the user which chart files will change and why before editing them.
7. Generate `ci.yml` (`schemaVersion: v0.4`) from the confirmed decisions, following the migration guide's worked examples for either branch.
8. Write `ci.yml`. Cloud Native branch: also apply the confirmed chart edits from step 6, at their existing locations. Landscape branch: the chart stays untouched (kept only as historical reference — Helm never runs on this branch).
9. Summarize: which branch was taken and why, the per-component decisions from steps 3 and 5, every file written or edited, and which `${{ vault.* }}` references still need real values before the first sync.

## Keep in mind

- **Never actively deploy anything.** No `cs start`, no `POST /workspaces/{id}/landscape/deploy`, no real `helm install`/`kubectl apply` against a target. Every Helm/kubectl command this skill produces is a `command:` line inside `ci.yml`, never executed directly by this skill.
- **Cloud Native branch:** the generated `ci.yml` needs its own `run.<serviceName>` block with `provider.name: virtual-k8s` and `provider.schemaVersion` — easy to silently drop since the helm/kubectl `command:` lines alone look sufficient on their own. Check for it explicitly before finishing.
- **Landscape branch:** no `helm`/`kubectl`/`docker` command anywhere in the output, and every `image:` is a straight, unmodified pull from the chart — the whole point of this branch is that nothing gets rebuilt that didn't need rebuilding. Before finishing, run the same self-check `codesphere-create-reactive-deployment` applies to its own output: pinned Dockerfile versions have a matching Nix install, every template reference uses `${{ ... }}` (never bare `{{ ... }}`), every field is real and documented.
- Renaming a service key in `ci.yml` after deploy forces recreation — data loss risk for stateful services. Treat a name as fixed once it's chosen.
- Treat complexity assessment (step 4) as per-component, never chart-wide, and run it *after* the Managed Service pass (step 3) — a StatefulSet running Postgres stops counting as complexity the moment it's replaced.
- Ask each Decision Point (steps 3, 4, 5) once, as a batch — don't re-litigate it later in the same run.

## Related

- `codesphere` — reference knowledge this skill reads from (loose coupling, read-only, never auto-invoked): `references/migration-guide.md` (the decision tree and worked examples this skill's process follows), `references/runtimes.md` (Reactive recipes), `references/landscape.md` (networking/`stripPath`)
- `codesphere-create-container-deployment` / `codesphere-create-reactive-deployment` — not hand-off targets for a Helm migration (this skill's steps 5/7 handle the per-component split inline); still the right entry point for a repo with Dockerfiles/source but no Helm chart
