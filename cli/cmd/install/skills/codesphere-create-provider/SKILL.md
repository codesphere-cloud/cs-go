---
name: codesphere-create-provider
description: Generates or edits a provider.yml manifest that turns an existing Codesphere Landscape (a repo with a working ci.yml) into a one-click catalog entry other teams can deploy. Batches its proposals rather than interviewing field-by-field — derives identity fields (reusing publiccode.yml if present) and configSchema/secretsSchema/detailsSchema candidates from the actual ci.yml content in one pass, presents the whole thing as a reviewable draft, and applies corrections. When a provider.yml already exists, works as an editor — loads it as a baseline and proposes only the delta against the current ci.yml/publiccode.yml, never re-deriving everything from scratch. Never publishes the provider itself (no POST to /api/managed-services/providers) — deliverable is the provider.yml file only, and it always requires ci.yml to already exist. Trigger for "provider.yml erstellen", "provider.yml anpassen/bearbeiten", "unser Repo als Catalog-Eintrag anbieten", "custom Service Provider veröffentlichen", "codesphere-create-provider ausführen".
allowed-tools: Bash Read Write Glob Grep
metadata:
  version: "1.0.0"
  cost-tier: medium
---

> **Process:** When this skill is explicitly/directly invoked by name, execute it immediately — don't ask the user what they want done with it. Proceed straight to Phase 0. The only prompts are the Decision Points and Blocking conditions defined within each phase below.

## When to use this

Trigger when the user wants to turn an existing Codesphere Landscape (a repo that already has a working `ci.yml`) into a publishable custom Service Provider, or wants to edit one that already exists — e.g. "provider.yml erstellen", "provider.yml bearbeiten", "unseren Landscape als Catalog-Eintrag anbieten", "codesphere-create-provider ausführen". This is not for creating the underlying `ci.yml` itself — that's `codesphere-create-cluster-deployment` / `-container-deployment` / `-reactive-deployment` / `codesphere-add-managed-service`. This skill wraps an *already-working* Landscape into a manifest others can one-click deploy, and can be re-run against the same repo later as the ci.yml or publiccode.yml evolve.

## Hard Gate

- **Shared family conventions apply — see `references/skill-family-conventions.md`** (inside `codesphere`'s directory; this skill has no `references/` folder of its own — `Glob` for `**/codesphere/references/*.md` if the install path isn't already known, specifically `references/custom-provider.md` and `references/ci-pipeline.md` for this skill). Covers locating/reading `codesphere`'s other `references/*.md` files (never requires `codesphere` itself to be loaded as an active skill). If `codesphere` can't be located, stop and say so — don't invent the manifest schema from memory.
- **MUST require a working `ci.yml` to already exist — no exceptions.** This is the one non-negotiable precondition: without `ci.yml`, there is no `provider.yml`, full stop. This skill generates a manifest *for* an existing Landscape, it never generates the Landscape itself and never proceeds on the assumption that one will show up later. If no `ci.yml` exists, point the user at the appropriate `codesphere-create-*-deployment` skill and stop — this applies in both new and edit mode.
- **MUST batch proposals rather than interviewing field-by-field.** Phases 2 and 4 derive their entire proposal in one pass and present it as one reviewable draft with one correction round — not a question per field. The output is explicitly a starting proposal for the user to sanity-check, not a series of individually-confirmed facts.
- **MUST place `provider.yml` only at the repository root** — matching where Codesphere expects it for the Git-URL publishing method.
- **MUST reuse `publiccode.yml` if one exists in the repo, rather than re-asking for overlapping fields.** `name`/`displayName`, `description`, and `author` have direct or near-direct counterparts between the two manifests — don't make the user restate what's already on record. Only ask for what `publiccode.yml` doesn't cover or states differently enough to need confirmation (see Phase 2).
- **MUST derive `configSchema`/`secretsSchema` candidates from the actual `ci.yml` content, never invent plausible-sounding fields.** Every candidate must trace back to a real `env:`/`config:`/`secrets:` entry in the file being wrapped.
- **MUST NOT publish the provider.** No `POST /api/managed-services/providers`, no call against a live Codesphere instance. This skill's deliverable is `provider.yml` (and, if the inline-spec publishing method is chosen instead of Git-URL, the equivalent JSON payload as a reference snippet in the summary) — nothing gets registered anywhere.
- **MUST validate `name` against `^[-a-z0-9_]+$` and `version` against `v[0-9]+`** before finalizing — these are hard pattern requirements per `references/custom-provider.md`, not soft suggestions.

## Process: 7-Phase Workflow

**Phase [N]: [Action Title]**
- **Prerequisite:** What must be true before this phase
- **Blocker risk:** What could fail and why
- **Action:** What you do (commands, decision points)
- **Output:** What success looks like
- **Blocking conditions:** If X or Y, stop and inform user

### Phase 0: Preflight

**Prerequisite:** Skill invoked.

**Blocker risk:** No `ci.yml` present at all — there's nothing to wrap into a provider yet.

**Action:** Confirm the repository root. Check for `ci.yml` (or a named `ci.<profile>.yml`) at the root.

**Output:** Confirmed target Landscape file to wrap.

**Blocking conditions:** No `ci.yml`/`ci.<profile>.yml` found anywhere at the root → stop. Point at `codesphere-create-cluster-deployment` / `-container-deployment` / `-reactive-deployment` depending on what the repo actually contains (Helm chart / Dockerfiles / source), or `codesphere-add-managed-service` if only a database-style service is needed. This skill does not generate the Landscape itself.

### Phase 1: Check for an existing `provider.yml` — new vs. edit mode

**Prerequisite:** Phase 0 completed.

**Blocker risk:** Treating "update" as "regenerate from scratch and hope it matches" instead of an actual edit; silently changing `name` on an already-published provider.

**Action:** Check whether `provider.yml` already exists at the root.

- **No** → **new mode**: continue to Phase 2, deriving everything fresh.
- **Yes** → **edit mode**: read the existing `provider.yml` in full and use it as the baseline for every subsequent phase — Phase 2 and Phase 4 don't re-derive from scratch, they compute a *diff* against this baseline (see those phases). This is what makes the skill usable as an ongoing editor, not just a one-time generator. If Phase 2 would produce a different `name` than the existing file has, flag that specifically and separately — renaming an already-published provider is a real breaking action, confirm it explicitly rather than folding it into the general batch review.

**Output:** Mode determined (new vs. edit); existing `provider.yml` content loaded as baseline if in edit mode.

**Blocking conditions:** In edit mode, do not proceed with a `name` change until it's been separately confirmed (see above).

### Phase 2: Gather identity fields — batched proposal

**Prerequisite:** Phase 1 completed.

**Blocker risk:** Turning this into a field-by-field interview instead of one reviewable proposal; picking a `name` that doesn't match the required pattern.

**Action:** Derive all identity fields in one pass, without asking per field:
- Check for `publiccode.yml` at the root. If present, derive `name` (slugified to satisfy `^[-a-z0-9_]+$`), `description` (from `description.<lang>.shortDescription`/`longDescription`), and `author` (from maintenance/legal contact info, if a single clear value exists) from it directly.
- Fill in whatever `publiccode.yml` doesn't cover (or everything, if it doesn't exist) with sensible defaults: `version` → `v1` for a first publish; `displayName` → title-cased `name`; `category` → best guess from the Landscape's contents (e.g. a `postgres`/`valkey`/etc. managed-service-heavy `ci.yml` suggests `databases`, a messaging-heavy one suggests `messaging`) rather than leaving it blank, but flag it clearly as a guess in the proposal below.
- In edit mode: compute this the same way, then diff against the loaded baseline — only surface fields that would actually change.

Present the whole set as **one proposal** — "Hier ist mein Vorschlag für die Identitätsfelder: ..." — and ask for one confirmation/correction pass, not a question per field. The result is explicitly a draft the user reviews and adjusts, not a series of individually-confirmed facts.

**Output:** `name`, `version`, `author`, `displayName`, `iconUrl` (optional, only if the user has one ready), `category`, `description` — proposed as a batch (new mode) or as a diff against the existing file (edit mode).

**Blocking conditions:** Do not proceed to Phase 3 with a `name` that doesn't match `^[-a-z0-9_]+$` or a `version` that doesn't match `v[0-9]+` — fix before moving on.

### Phase 3: Determine the Landscape backend

**Prerequisite:** Phase 2 completed.

**Blocker risk:** Pointing at the wrong branch/profile; a private repo the publishing mechanism can't actually reach.

**Action:** Derive `backend.landscape.gitUrl` from `git remote get-url origin`. Determine `backend.landscape.ciProfile` — if the repo's `ci.yml` is the default (no profile suffix), note that explicitly; if a named profile file was used (`ci.<profile>.yml`, e.g. from a prior `codesphere-create-*-deployment` run), use that profile name. If more than one `ci.<profile>.yml` exists, ask which one this provider should deploy.

**Output:** `backend.landscape.gitUrl` and `backend.landscape.ciProfile` resolved.

**Blocking conditions:** None, beyond confirming the right profile when more than one exists.

### Phase 4: Derive configSchema / secretsSchema / detailsSchema — batched proposal

**Prerequisite:** Phase 3 completed.

**Blocker risk:** Turning this into a long field-by-field interview instead of one reviewable proposal; exposing something as user-configurable that should stay fixed (or vice versa); inventing a field that doesn't correspond to anything real in the `ci.yml`.

**Action:** Read the target `ci.yml` in full. For every `env:` entry, `config:` entry, and `${{ vault.NAME }}`-templated `secrets:` entry across all services, classify it using this heuristic **without asking per field**:

- Looks internal/structural (a port number, an internal hostname/DNS reference, a value derived from another service, anything that's clearly plumbing rather than something a *different* team deploying this provider would want to change) → propose leaving it fixed, not exposed anywhere in the manifest.
- Looks like real application configuration a consuming team would plausibly want to set themselves (a display name, a feature flag, a numeric limit, anything resembling `SITE_NAME`/`MAX_USERS`-style examples) → propose it as a `configSchema` property, with a JSON Schema `type` matching its apparent shape (string/integer/boolean).
- Is a `${{ vault.NAME }}`-templated secret → propose it as a `secretsSchema` property (`format: password` where credential-shaped).
- A numeric field that looks like it should only grow (storage-like) → propose `x-update-constraint: increase-only`. A field that looks like a version/engine pin → propose `x-update-constraint: immutable`. Apply these as part of the same proposal, don't ask separately per field.
- Any output/detail Codesphere itself reports post-provisioning (hostname, URL, port) that a consuming team would plausibly want surfaced → propose it for `detailsSchema`.

In edit mode (Phase 1), compute this the same way, then diff against the baseline: new `ci.yml` fields not yet in the existing `provider.yml`'s schemas → propose adding; existing schema entries whose underlying `ci.yml` field no longer exists → flag for possible removal, don't drop them silently.

Present the **entire schema as one proposal** — every field, its proposed classification, and a one-line reason ("`SITE_NAME` sieht nach Nutzer-Konfiguration aus → `configSchema`"; "`internal_port` wird intern zwischen Services verwendet → bleibt fest") — and ask for **one** review/correction pass covering everything at once, not a question per field. This is explicitly a draft: the user corrects whatever the heuristic got wrong in one go, rather than confirming each item individually.

**Output:** `configSchema`, `secretsSchema`, `detailsSchema` — each only containing fields traced back to something real in the `ci.yml`, presented as one batched proposal (new mode) or one batched diff (edit mode), with the user's corrections applied.

**Blocking conditions:** Do not add a schema property that doesn't correspond to an actual `ci.yml` field.

### Phase 5: Write `provider.yml`

**Prerequisite:** Phase 4 completed.

**Action:** Assemble the manifest per `references/custom-provider.md`'s documented shape (`name`, `version`, `author`, `displayName`, `iconUrl`, `category`, `description`, `backend.landscape.gitUrl`/`ciProfile`, `configSchema`, `secretsSchema`, `detailsSchema`), using Phase 2 and Phase 4's confirmed proposals. In edit mode, this means applying the confirmed diff on top of the loaded baseline — fields the user didn't touch keep their existing values verbatim, not re-derived from scratch. Write the result to the repository root.

**Output:** `provider.yml` written or updated. The whole file is still a **draft for the user to review** — this skill's job is to produce a well-reasoned starting point, not a final, authoritative manifest; say so plainly in Phase 6 rather than implying it's ready to publish as-is.

**Blocking conditions:** Before finishing, re-check `name`/`version` against their required patterns one more time — this is the last point before the file exists on disk.

### Phase 6: Summary

**Prerequisite:** Phase 5 completed.

**Action:** Summarize for the user:
- Whether this was a fresh generation or an edit of an existing `provider.yml`, and — in edit mode — exactly what changed vs. the prior version.
- The resolved identity fields, and which ones were reused from `publiccode.yml` vs. proposed by default.
- Every `configSchema`/`secretsSchema`/`detailsSchema` field, which `ci.yml` field it traces back to, and the one-line reasoning from Phase 4's proposal.
- The Git-URL publishing method this manifest supports out of the box (`POST /api/managed-services/providers` with `{"gitUrl": ..., "scope": {...}}`, per `references/custom-provider.md`) — provide the exact `curl` example with the repo's own `gitUrl` filled in, but do not execute it. Mention that `scope.type: global` needs cluster admin permissions, while `team` scope only needs `teamIds`.
- Explicit reminder: this is a **draft** — nothing has been published anywhere, and the heuristic classifications in Phase 4 are a starting proposal the user should sanity-check, not a final answer.

**Output:** One consolidated status message the user can act on.

**Blocking conditions:** None.

## Related

- `codesphere` — reference knowledge this skill reads from (loose coupling, read-only, never auto-invoked); specifically `references/custom-provider.md`, `references/ci-pipeline.md`
- `codesphere-create-cluster-deployment` / `-container-deployment` / `-reactive-deployment` / `codesphere-add-managed-service` — produce the `ci.yml` this skill wraps; run one of these first if no Landscape exists yet
