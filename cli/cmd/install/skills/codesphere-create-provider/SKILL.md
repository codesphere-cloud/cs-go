---
name: codesphere-create-provider
description: 'Generates or edits a provider.yml manifest that turns an existing Codesphere Landscape (a repo with a working ci.yml) into a one-click catalog entry other teams can deploy, batching field derivation into one reviewable draft rather than interviewing field-by-field. Only run on explicit invocation. Requires ci.yml to already exist — for creating the Landscape itself, use codesphere-create-cluster-deployment / -container-deployment / -reactive-deployment or codesphere-add-managed-service instead. Never publishes the provider (no POST to /api/managed-services/providers) — the deliverable is the provider.yml file only.'
allowed-tools: Bash Read Write Glob Grep
metadata:
  version: "1.0.0"
  cost-tier: medium
---

> **Process:** When this skill is explicitly/directly invoked by name, execute it immediately — don't ask the user what they want done with it.

## When to use this

Trigger when the user wants to turn an existing Codesphere Landscape (a repo that already has a working `ci.yml`) into a publishable custom Service Provider, or wants to edit one that already exists — e.g. "provider.yml erstellen", "provider.yml bearbeiten", "unseren Landscape als Catalog-Eintrag anbieten", "codesphere-create-provider ausführen". Not for creating the underlying `ci.yml` itself — that's `codesphere-create-cluster-deployment` / `-container-deployment` / `-reactive-deployment` / `codesphere-add-managed-service`. This skill wraps an *already-working* Landscape into a manifest others can one-click deploy, and can be re-run against the same repo later as the `ci.yml` or `publiccode.yml` evolve.

## Reference

Shared family conventions apply — see `references/skill-family-conventions.md` inside `codesphere`'s directory (this skill has no `references/` folder of its own; `Glob` for `**/codesphere/references/*.md` if the install path isn't already known). Read `references/custom-provider.md` (manifest schema, `name`/`version` patterns, publishing methods) and `references/ci-pipeline.md` (`ci.yml` field shapes) before generating anything — this skill doesn't repeat their content, it's the process wrapper. If `codesphere` can't be located, stop and say so rather than inventing the manifest schema from memory.

## Process

1. **Preflight.** Find `ci.yml` (or a named `ci.<profile>.yml`) at the repo root. None found → stop, don't generate anything, and point at whichever of `codesphere-create-cluster-deployment` / `-container-deployment` / `-reactive-deployment` / `codesphere-add-managed-service` fits the repo's contents. Applies in both new and edit mode — this skill never generates the Landscape itself.
2. **Determine mode.** Check whether `provider.yml` already exists at the root. None → **new mode**, derive everything fresh. Exists → **edit mode**: read it in full as the baseline; every later step computes a *diff* against it instead of re-deriving from scratch. In edit mode, if the derived `name` would differ from the baseline, flag that specifically and confirm it separately before proceeding — renaming an already-published provider is a real breaking action, not a routine field edit.
3. **Identity fields, batched.** Derive `name` (slugified), `description`, and `author` from `publiccode.yml` if present, rather than re-asking for overlapping fields; fill whatever it doesn't cover with defaults (`version` → `v1`, `displayName` → title-cased `name`, `category` → best guess from the Landscape's contents, flagged as a guess). Edit mode: diff against the baseline, surface only fields that would change. Present the whole set as one proposal with one confirmation/correction pass — not a question per field — and validate `name` against `^[-a-z0-9_]+$` and `version` against `v[0-9]+` before moving on.
4. **Backend.** Derive `backend.landscape.gitUrl` from `git remote get-url origin`. Derive `backend.landscape.ciProfile` from the default `ci.yml` or a named `ci.<profile>.yml`; if more than one profile file exists, ask which one this provider should deploy.
5. **Schemas, batched.** Read the target `ci.yml` in full and classify every `env:`/`config:`/`secrets:` entry across all services, without asking per field: internal/structural (ports, internal DNS, plumbing) → stays fixed, not exposed; real application config a consuming team would want to set → `configSchema`, with a JSON Schema type matching its shape; `${{ vault.NAME }}`-templated → `secretsSchema` (`format: password` where credential-shaped); storage-like numeric field → `x-update-constraint: increase-only`; version/engine pin → `x-update-constraint: immutable`; post-provisioning output (hostname/URL/port) → `detailsSchema`. Every candidate must trace back to a real entry in `ci.yml` — never invent one. Edit mode: diff against the baseline — new `ci.yml` fields not yet in the schemas get proposed as additions, existing schema entries whose field disappeared get flagged for possible removal, never dropped silently. Present the entire schema as one proposal, each field with a one-line reason, and one review/correction pass covering everything at once.
6. **Write `provider.yml`.** Assemble the manifest per `references/custom-provider.md`'s shape, using the confirmed proposals from steps 3–5. Write only to the repository root — that's where Codesphere expects it for Git-URL publishing. Edit mode: apply the confirmed diff on top of the baseline; fields the user didn't touch keep their existing values verbatim. Before finishing, re-check `name`/`version` against their required patterns one last time.
7. **Summary.** Report whether this was a fresh generation or an edit (and exactly what changed, in edit mode); which identity fields were reused from `publiccode.yml` vs. defaulted; every `configSchema`/`secretsSchema`/`detailsSchema` field with its `ci.yml` provenance and reasoning; the Git-URL publishing method's exact `curl` example (`POST /api/managed-services/providers` with `{"gitUrl": ..., "scope": {...}}`, `scope.type: global` needing cluster admin vs. `team` scope only needing `teamIds`) — constructed for reference, not executed; and an explicit reminder that this is a **draft**, nothing has been published.

## Keep in mind

- `ci.yml` must already exist — no exceptions, in new or edit mode. This skill wraps a Landscape, it never creates one.
- `provider.yml` lives only at the repository root, never elsewhere.
- Batch, don't interview: identity fields (step 3) and schemas (step 5) are each one reviewable draft with one correction round, never a question per field.
- `configSchema`/`secretsSchema` candidates must trace back to a real `env:`/`config:`/`secrets:` entry in `ci.yml` — never invented, however plausible-sounding.
- Never publish: no `POST /api/managed-services/providers`, no call against a live instance. If the inline-spec method is chosen instead of Git-URL, the deliverable is the equivalent JSON snippet in the summary, not an executed call.
- `name` must match `^[-a-z0-9_]+$`, `version` must match `v[0-9]+` — hard pattern requirements, validate before finalizing.
- A `name` change in edit mode is a breaking action (recreates an already-published provider's identity) — confirm it separately, don't fold it into the general batch review.
- If `codesphere` and its `references/` can't be located, stop and say so rather than inventing the manifest schema from memory.
- The written file is always a draft for the user to sanity-check, not an authoritative final manifest — say so plainly rather than implying it's ready to publish as-is.

## Related

- `codesphere` — reference knowledge this skill reads from (loose coupling, read-only, never auto-invoked): `references/custom-provider.md`, `references/ci-pipeline.md`
- `codesphere-create-cluster-deployment` / `-container-deployment` / `-reactive-deployment` / `codesphere-add-managed-service` — produce the `ci.yml` this skill wraps; run one of these first if no Landscape exists yet
