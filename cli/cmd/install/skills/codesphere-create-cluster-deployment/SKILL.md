---
name: codesphere-create-cluster-deployment
description: Generates a ci.yml for migrating an existing Helm chart in the repository onto Codesphere — by default decomposing it into a single Landscape that mixes Reactive, Managed Container, and Managed Service entries per component (the production-recommended path), falling back to a Cloud Native Deployment (Virtual Kubernetes Cluster, chart kept as-is) only for the part of a chart that genuinely needs real Kubernetes semantics. Only run on explicit invocation — general Codesphere questions belong to the reference skill codesphere, not this command. Before generating anything, identifies which chart components (including database/cache/queue subcharts like Postgres/Redis/RabbitMQ) map to a Codesphere Managed Service, then classifies every remaining component individually — Reactive for a component with its own Dockerfile/code, Managed Container for an unmodified vendor image — rather than forcing one deployment type onto the whole chart. The standard frontend+backend+database(+off-the-shelf sidecar) chart is the common case this applies to, not an edge case.
license: none
allowed-tools: Bash Read Write Glob Grep
metadata:
  version: "2.0.0"
  updated: "2026-09-01"
  cost-tier: high
---

> **Process:** When this skill is explicitly/directly invoked by name, execute it immediately — don't ask the user what they want done with it. Proceed straight to Step 0. The only prompts are the Decision Points defined within each Step below.

## When to use this

Trigger when the user wants a `ci.yml` generated from an existing Helm chart — this skill decomposes it into a Codesphere Landscape by default (Reactive/Managed Container/Managed Service, chosen per component) and falls back to a Cloud Native Deployment only for the part of the chart that genuinely needs Kubernetes semantics — e.g. "Helm-Chart auf Codesphere migrieren", "ci.yml aus meinem Helm-Chart erstellen", "codesphere-create-cluster-deployment ausführen". Also reached automatically as a suggested next step when a user has a Helm chart and no `ci.yml` yet. If the user has no Helm chart at all, this skill is the wrong entry point — see `codesphere-create-container-deployment` (Dockerfiles/`docker-compose.yml`) or `codesphere-create-reactive-deployment` (application source) instead.

## Production Recommendation

Cloud Native Deployment (a team's Virtual Kubernetes Cluster) is fully supported, but it is **not** the recommended default for production — it's a fallback for when a chart genuinely needs real Kubernetes semantics (CRDs, operators, StatefulSets with no managed-service equivalent) that survive Step 3's managed-service replacement. A self-managed cluster shifts real operational and compliance responsibility onto the customer that Codesphere's own Landscape primitives (Reactive/Managed Container/Managed Service) otherwise cover automatically — for example, confidential-compute/isolation guarantees have to be built and proven by the customer themselves on a self-managed cluster, instead of being a platform guarantee. This is why Step 4 below defaults to decomposing into a Landscape whenever the chart allows it, and why continuing with Cloud Native Deployment requires the chart to actually earn that fallback, not just exist. Full rationale and a worked example: `references/migration-guide.md`'s "Helm Chart → Codesphere Landscape" section.

## Hard Gate

- **Shared family conventions apply — see `references/skill-family-conventions.md`** (inside `codesphere`'s directory; this skill has no `references/` folder of its own — `Glob` for `**/codesphere/references/*.md` if the install path isn't already known). Covers locating/reading `codesphere`'s other `references/*.md` files (never requires `codesphere` itself to be loaded as an active skill — that's the missing-sibling case covered by the next Hard Gate if it truly can't be found) and the repo-root-only `ci.yml` placement used below.
- **MUST classify Reactive vs. Managed Container per remaining component, never as one chart-wide choice.** A chart with a custom frontend/backend plus a stock vendor sidecar (Grafana, a `bitnami/*` image, a cert-manager helper, ...) needs both types in the same generated `ci.yml`, not a single yes/no answer forced onto every component. Use `references/migration-guide.md`'s per-component table (local Dockerfile+own code → Reactive; unmodified vendor image reference, no local build → Managed Container; neither clear → ask) — never guess when the signal is genuinely ambiguous.

- **MUST NOT judge chart complexity before managed-service detection, and MUST NOT judge it chart-wide instead of per-component.** A StatefulSet running Postgres/Redis/RabbitMQ/Mongo/etc. that maps to a Codesphere Managed Service is not a complexity signal — it's a candidate for removal from the chart entirely, and stops counting as chart complexity the moment it's replaced. Judging complexity before checking for a managed-service replacement misclassifies the single most standard use case there is — an app split into a frontend, a backend, and a database — as "too complex for Reactive" purely because the database happens to run as a StatefulSet, when the correct read is: frontend → Reactive, backend → Reactive, database → Managed Service, no cluster needed at all. Steps 3–4 are ordered specifically to avoid this — do not reorder them.
- **MUST only edit the Helm chart's own files (Step 5) when continuing on the Cloud Native Deployment path.** Never when decomposing into a Landscape (Step 4b) — Helm never runs on that path, so there's nothing in the chart to disable or rewire; the chart stays untouched, kept only as historical reference.
- **MUST NOT perform any active deployment.** No `cs start`, no `POST /workspaces/{id}/landscape/deploy`, no `helm install` against a real target. Every Helm/kubectl command generated becomes a `command:` line inside `ci.yml` — it is never executed by this skill itself.
- **MUST include a `run.<serviceName>` block with `provider.name: virtual-k8s` and `provider.schemaVersion`** (not `provider.version`) whenever continuing with Cloud Native Deployment. This is easy to silently drop — a `ci.yml` with `prepare.steps` that just run `helm`/`kubectl` commands but no `virtual-k8s` Managed Service block assumes a cluster that nothing in the file actually provisions. Verify this explicitly before finishing Step 6; "the helm/kubectl commands are present" is not sufficient on its own.
- **MUST NOT re-ask a per-component or managed-service Decision Point already answered earlier in the same run** — Step 3's replace/keep decisions and Step 4b's Reactive/Managed-Container classifications are each asked once, as one batched Decision Point, not re-litigated later in the workflow.

## Process: 10-Step Workflow (Steps 0–4, 4b, 5–8)

### Step 0: Determine repo root

**Prerequisite:** Skill invoked.

**Action:** `ci.yml` always belongs at the repository root — never in a subdirectory, even in a monorepo with multiple Helm charts. In a monorepo, identify the specific chart the request is about, but still place `ci.yml` at the root.

**Output:** Confirmed target path for `ci.yml` (repo root).

**Blocking conditions:** None.

### Step 1: Check for an existing `ci.yml`

**Prerequisite:** Step 0 completed.

**Action:** Check whether a `ci.yml` already exists at the root.

- **No** → continue to Step 2.
- **Yes** → **Decision Point**: ask the user whether the existing `ci.yml` should be overwritten or updated. "Update" means preserving the existing services/structure as much as possible and only adding/replacing the new cluster-deployment service. "Overwrite" means rebuilding it from scratch.

**Output:** Overwrite-vs-update decision recorded.

**Blocking conditions:** Do not proceed without an answer when a `ci.yml` already exists.

### Step 2: Look for a Helm chart

**Prerequisite:** Step 1 completed.

**Action:** Search the repository for a Helm chart (`Chart.yaml`, typical paths like `./chart`, `./helm`, `./deploy/chart`, or a top-level search). This includes umbrella/parent charts with subcharts declared under `dependencies:` in `Chart.yaml` (e.g. a parent chart with `frontend`, `backend`, and `postgres` as local `file://` dependencies) — treat each subchart as its own component in the steps below, not as one undifferentiated blob.

**Output:** Chart location and component list (single chart, or list of subcharts for an umbrella chart).

**Blocking conditions:**
- **No chart found** → **Blocker, abort.** Tell the user no Helm chart was found and this skill can't meaningfully proceed without one. Do not generate a `ci.yml`. (For deployments without a Helm chart: `codesphere-create-reactive-deployment` or `codesphere-create-container-deployment`.)

### Step 3: Detect managed-service candidates, per component

**Prerequisite:** Step 2 found at least one chart component.

**Action:** Before judging complexity, identify every component of the chart that is a database/cache/queue workload with a Codesphere Managed Service equivalent. Look at each component's templates and `values.yaml` for the underlying engine (own templates, e.g. a `StatefulSet` running `postgres:*`/`redis:*`/`rabbitmq:*` images, or a Bitnami-style subchart), then check `references/providers.md` and the individual `references/provider-*.md` files for a match:

| Component runs | Possible replacement | Provider |
|---|---|---|
| PostgreSQL | `references/provider-postgresql.md` | `postgres` |
| MongoDB-compatible | `references/provider-documentdb.md` | `ferretdb` |
| Redis/Valkey | `references/provider-valkey.md` | `valkey` |
| RabbitMQ | `references/provider-rabbitmq.md` | `rabbitmq` |
| Elasticsearch/OpenSearch | `references/provider-opensearch.md` | `opensearch` |
| S3-compatible object storage/MinIO | `references/provider-object-storage.md` | `s3` |
| SQL Server-compatible | `references/provider-babelfish.md` | `babelfish` |

**Decision Point:** present each match to the user individually — the component in the chart, the proposed managed service, and what it means (data moves from an in-cluster pod to a Codesphere-managed service; existing data would need migration, which this skill does not perform). The user decides per candidate: replace or leave it in the chart. Nothing gets swapped automatically without confirmation.

**Output:** Per-component replace/keep decisions. These carry directly into Step 4 (a replaced component is removed from the complexity assessment entirely) and into Step 5 (every confirmed replacement needs real changes to the chart's own files, not just a new block in `ci.yml`).

**Blocking conditions:** None — proceeds regardless of how many (or few) replacements are confirmed.

### Step 4: Assess remaining complexity — cluster or Landscape?

**Prerequisite:** Step 3's replace/keep decisions recorded.

**Action:** Assess complexity using only what's left after Step 3's confirmed replacements are set aside. For each remaining component, check against the decision tree in `references/migration-guide.md`:

- **Trivial** (one Deployment + one Service, maybe an Ingress/ConfigMap/Secret, no CRDs/operators/StatefulSets) → decompose into a Landscape (Step 4b) — automatic off-when-unused, path routing, and monitoring that a Cloud Native Deployment doesn't get automatically, and the production-recommended default per the Production Recommendation above.
- **Non-trivial** (multiple resources, CRDs, operators, StatefulSets that were *not* covered by a Step 3 replacement) → Cloud Native Deployment is the right fit for that component.

For an umbrella chart, look at the components together: if every remaining component is individually trivial, the whole app is trivial — a frontend Deployment+Service+Ingress plus a backend Deployment+Service+Ingress+Secret, with the database already peeled off into a Managed Service in Step 3, is the textbook trivial case, not a borderline one.

**Decision Point (only if the overall result is trivial):** a Landscape is the default whenever the chart allows it, per the Production Recommendation above — Cloud Native Deployment is the fallback, not a coin flip. Name the concrete components and ask:

**"Deploy without the Helm chart/cluster at all?"**
- **No** → proceed to Step 5 (Cloud Native Deployment, chart stays the source of truth). For a trivial result, restate the Production Recommendation's tradeoff (lost off-when-unused/monitoring/routing automation, and the customer taking on compliance burden like confidential-compute proof) before accepting "No" — make sure it's a deliberate choice, not a default.
- **Yes** → proceed to Step 4b. Do not proceed to Step 5 for this branch — Step 5's Helm-chart edits apply only to the Cloud Native Deployment path.

**Output:** Either a confirmed continuation into Step 5 (Cloud Native Deployment), or a confirmed continuation into Step 4b (Landscape).

**Blocking conditions:**
- Do not proceed past this Decision Point without an answer when the result is trivial.
- For a genuinely non-trivial result, skip this Decision Point entirely and continue straight to Step 5 — don't ask a question that has an obvious answer.

### Step 4b: Classify each remaining component — Reactive or Managed Container? (Landscape branch only)

**Prerequisite:** Step 4 resulted in "Yes" (decompose into a Landscape).

**Action:** For every component still remaining after Step 3's managed-service replacements, classify it using `references/migration-guide.md`'s per-component table — never as one chart-wide choice:

- **No local Dockerfile/build context for this component anywhere in the repo, and the chart/`values.yaml` references a recognizable vendor/pre-built image tag** (e.g. `grafana/grafana:10.4.2`, a `bitnami/*` image, a cert-manager helper) → **Managed Container** (`image:` = that exact image reference). Nothing to rebuild — pulling it as-is is correct and less work than reimplementing someone else's app as Reactive steps.
- **A local Dockerfile/build context exists for this component in the repo, adding the team's own code/program on top of a base image** → **Reactive**, rebuilt from source per `references/runtimes.md` (Dockerfile `RUN` → `prepare.steps`, `CMD`/`ENTRYPOINT` → the last `run.<service>.steps` entry, `EXPOSE`/`ENV` → `network`/`env`). The base image's genericness is irrelevant — `FROM ubuntu` plus the team's own `COPY`/`RUN` is exactly as much "Reactive" as a language-specific base image. A pinned `FROM` version is mandatory to reproduce via Nix, not optional.
- **Neither signal is clear** (an image reference with no local Dockerfile *and* no recognizable vendor name — e.g. a private-registry image some other pipeline builds) → **ask the user directly**: "is `<image>` a vendor/off-the-shelf image, or does your team build it (just not in this repo)?" Don't guess either way.

**Decision Point:** present the whole per-component classification as one batch — component, proposed type (Reactive/Managed Container), one-line reason — and ask for one review/correction pass, plus a direct question for any component the signal couldn't resolve. Same house style as Step 3's per-component managed-service matches — not a question per component.

**Output:** Confirmed Reactive-or-Managed-Container classification for every remaining component, feeding directly into Step 6's Landscape branch.

**Blocking conditions:** Do not proceed to Step 6 with any component whose classification is still unresolved from an "unclear" case above.

### Step 5: Determine and apply the Helm chart changes each replacement requires (Cloud Native Deployment branch only)

**Prerequisite:** Step 4 resulted in continuing with Cloud Native Deployment. If Step 4 resulted in the Landscape branch (Step 4b), skip this step — go straight to Step 6's Landscape branch instead.

**Action:** Updating `ci.yml` alone is not enough for a component replaced in Step 3 — the chart will keep deploying it via `helm install` unless the chart's own files change too. For every confirmed replacement, work out and apply:

- **Disable the replaced component in the chart.** If the umbrella `Chart.yaml` declares it as a subchart dependency with a `condition:` (e.g. `condition: postgres.enabled`), the clean fix is setting that flag to `false` — either in the chart's own `values.yaml` (persists as the chart's default, needs a file edit) or via a `--set <component>.enabled=false` flag on the `helm install`/`helm upgrade` line generated in Step 6. Prefer the `values.yaml` edit when the chart already structures things this way, since it keeps the chart itself accurate. If the chart has no such toggle, the component's own template files — and its entry in `Chart.yaml`'s `dependencies:` — need to be removed or commented out.
- **Rewire whatever depended on the replaced component.** Anything that pointed at the old in-cluster Service DNS name (e.g. a backend's `database.url` pointing at `postgres:5432`) needs to point at the Managed Service instead. Treat the connection details as a `${{ vault.NAME }}` value the user populates through the vault (see `references/secret-management.md`), never hardcoded into a chart file checked into git. Pass it in via a `--set` override on the `helm install`/`helm upgrade` line, e.g. `--set backend.database.url="postgresql://<user>:${{ vault.pgPassword }}@<managed-service-hostname>:5432/<db>"`.
  - **Caveat to flag to the user, not to assert silently:** referencing a sibling Managed Service's own generated `hostname`/`dsn` output directly from another service's `command:` line isn't confirmed anywhere in the available reference material. The safer fallback is a vault secret the user populates (or generates via the vault's `/generate` endpoint).

**Decision Point:** before touching any file other than `ci.yml`, tell the user plainly which files need to change and why. Get a quick confirmation, then make the edits — separate from Step 3's "replace or keep" decision.

**Output:** Applied edits to the chart's own files (`values.yaml`, `Chart.yaml`, or specific templates), scoped exactly to the confirmed replacements. A component kept in the chart (declined in Step 3) has its files left untouched.

**Blocking conditions:** Do not touch any file other than `ci.yml` without the Decision Point above having been answered first.

### Step 6: Generate `ci.yml`

**Prerequisite:** Step 5 completed, for the Cloud Native branch — or Step 4b completed, for the Landscape branch.

**Action — Cloud Native Deployment branch:**
- `schemaVersion: v0.4` (current — not `v0.2`).
- `prepare.steps`: `helm repo add`/`helm repo update` (if the chart comes from a repo, otherwise omit) + `helm install <release> <chart> -n <namespace> --create-namespace -f <values-file>`, extended with the `--set` overrides worked out in Step 5.
- For each replacement confirmed in Step 3: an additional `run.<serviceName>` Managed Service block following the schema in the matching `references/provider-*.md`, with `secrets` as `${{ vault.NAME }}` references (never plaintext).
- Add a comment noting which components from the chart were deliberately *not* replaced, so it stays traceable for future readers.

**Action — Landscape branch:**
- `schemaVersion: v0.4`.
- One `run.<serviceName>` per component classified in Step 4b: Reactive components get `steps:` built from `references/runtimes.md`'s matching recipe (Dockerfile `RUN`/`CMD`/`EXPOSE`/`ENV` as the source); Managed Container components get `image:` set to the exact reference from the chart, with `network`/`env` translated from the chart's Service/ConfigMap/env definitions.
- For each replacement confirmed in Step 3: an additional `run.<serviceName>` Managed Service block, exactly as in the Cloud Native branch above.
- Every service needs at least one `network.paths` entry or `isPublic: true`; `stripPath` must match whether that specific component's own routes already include the path prefix (`references/landscape.md`) — check the actual routes, don't default to either value.
- No `helm`, `kubectl`, or `docker` command anywhere in the generated file on this branch, and no service has an `image:` value that isn't a straight pull of an existing vendor image — the whole point of the Landscape branch is that Helm never runs and nothing gets rebuilt that didn't need rebuilding.
- Worked example for this exact shape: `references/migration-guide.md`'s "Worked Example — Mixed Landscape from One Chart".

**Output:** Draft `ci.yml` (either branch).

**Blocking conditions:**
- Cloud Native branch: before finishing, explicitly check the draft `ci.yml` for a `run.<serviceName>` block with `provider.name: virtual-k8s` and `provider.schemaVersion` — if it's missing, add it (see Hard Gate above).
- Landscape branch: before finishing, run the same self-check `codesphere-create-reactive-deployment` uses for its own output on every Reactive component (no helm/kubectl/docker anywhere; every Dockerfile-pinned runtime version has a matching Nix install; every template reference uses `${{ ... }}`, never bare `{{ ... }}`; every template reference uses a real, documented field), plus confirm every Managed Container's `image:` matches the exact reference from the chart.

### Step 7: Write the files

**Prerequisite:** Step 6 produced a validated draft `ci.yml`.

**Action:** Place `ci.yml` at the repository root (overwrite/update per the Step 1 decision). **Cloud Native branch:** also apply the Step 5 edits to the chart's own files at their existing locations — never move or rename chart files as part of this. **Landscape branch:** no other file changes — the chart's own files stay untouched, kept only as historical reference; Helm never runs on this branch, so there's nothing in the chart to disable or rewire.

**Output:** `ci.yml` written; Helm chart files edited where the Cloud Native branch's Step 5 required it.

**Blocking conditions:** No file gets touched without its corresponding Decision Point (Step 1 for `ci.yml`, Step 5 for chart files on the Cloud Native branch) having been answered first.

### Step 8: Summary

**Prerequisite:** Step 7 completed.

**Action:** Briefly summarize for the user:
- Which chart/components were used, how the overall result was classified (trivial/non-trivial) in Step 4, and which branch (Landscape or Cloud Native) was taken.
- Which managed-service replacements were made vs. declined in Step 3, and how that fed into the Step 4 verdict.
- **Landscape branch:** the Step 4b classification for every component (Reactive vs. Managed Container) and the one-line reason for each.
- **Cloud Native branch:** restate the Production Recommendation tradeoff plainly rather than leaving it implicit — what stays manual (Workspace Router wiring, monitoring integration for workloads inside the virtual cluster) and that the customer now carries compliance/operational responsibility (e.g. confidential-compute proof) the Landscape branch would have covered automatically.
- Every file that was written or edited, listed explicitly.
- Which `${{ vault.* }}` references still need real values populated before the first sync.

**Output:** One consolidated status message the user can act on.

**Blocking conditions:** None.

## Related

- `codesphere` — reference knowledge this skill reads from (loose coupling, read-only, never auto-invoked); specifically `references/migration-guide.md` (per-component classification table and the worked mixed-Landscape example this skill's Step 4b/6 follow), `references/runtimes.md` (Reactive recipes), `references/landscape.md` (networking/`stripPath`)
- `codesphere-create-container-deployment` / `codesphere-create-reactive-deployment` — no longer hand-off targets for a Helm migration (Step 4b/6 handle the per-component Reactive/Managed-Container split inline, in one file); still the right entry point for a repo with Dockerfiles/source but *no* Helm chart at all
