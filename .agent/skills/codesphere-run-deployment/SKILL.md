---
name: codesphere-run-deployment
description: Deploys an existing ci.yml on Codesphere — checks whether every ${{ vault.NAME }} secret and ${{ workspace.env[...] }} plain env var the file actually references is set, offers to configure whatever's missing (auto-generating vault secrets where safely possible, asking for values the platform can't invent), then triggers the deployment via Codesphere's public API (or the confirmed cs-go CLI equivalent) and reports pipeline/health status. Trigger for "jetzt deployen", "ci.yml ausrollen", "landscape syncen", "auf Codesphere deployen", or any request to actually run/publish a Codesphere deployment rather than just generate its ci.yml.
allowed-tools: Bash Read Write Glob Grep
metadata:
  version: "1.0.0"
  cost-tier: high
---

> **Process:** When this skill is explicitly/directly invoked by name, execute it immediately — don't ask the user what they want done with it. Proceed straight to Phase 0. The only prompts are the Decision Points and Blocking conditions defined within each phase below.

## When to use this

Trigger when the user wants an already-generated `ci.yml` actually deployed on Codesphere — e.g. "deploy das jetzt", "ci.yml ausrollen", "Landscape syncen", "codesphere-run-deployment ausführen". This is the one skill in the `codesphere` family that performs a real deployment action — `codesphere-create-cluster-deployment` / `-container-deployment` / `-reactive-deployment` and `codesphere-add-managed-service` all explicitly stop at generating/editing `ci.yml` and hand the actual deploy step to this skill.

## Hard Gate

- **This is the one skill in the family permitted to perform an active deployment.** Every sibling skill's Hard Gate says the opposite ("never deploys anything") — don't copy that constraint here, and don't let this skill's permission bleed into any sibling.
- **Shared family conventions apply — see `references/skill-family-conventions.md`** (inside `codesphere`'s directory; this skill has no `references/` folder of its own — `Glob` for `**/codesphere/references/*.md` if the install path isn't already known, specifically `references/cli-and-api.md` and `references/secret-management.md` for this skill). Covers locating/reading `codesphere`'s other `references/*.md` files (never requires `codesphere` itself to be loaded as an active skill).
- **MUST never print a secret or token value in plain text** — `CS_TOKEN`, any vault secret value, any generated password. Mask as `***` in every status message, log excerpt, and summary. This includes not echoing a value the user pastes into the conversation back into a later message.
- **MUST verify `ci.yml` exists before doing anything else.** If it doesn't, stop and point the user at `codesphere-create-cluster-deployment` / `-container-deployment` / `-reactive-deployment` / `codesphere-add-managed-service` instead — this skill deploys an existing file, it does not generate one.
- **MUST determine required secrets/env vars by parsing the actual `ci.yml` content** — every `${{ vault.NAME }}` and `${{ workspace.env['KEY'] }}` reference anywhere in the file, not a guess or a memory of what a "typical" deployment needs.
- **MUST distinguish auto-generatable vault secrets from values only the user can supply.** A vault secret with no external meaning (a random password Codesphere itself will use, e.g. a database password) can be created via the vault's `/generate` endpoint without ever asking the user for a value — this is a single API call, not a two-step "ask then store." A plain `workspace.env` var or a vault secret tied to something external (a third-party API key, an S3 credential for a service outside Codesphere) cannot be invented — only the user can supply it.
- **MUST stop if the user declines to configure a value that has no auto-generatable fallback** — per the user's own framing: if configuration is missing and the user says no to setting it up, and there's no safe way to generate it, the run ends there. Don't proceed to deploy with a known-missing required value.
- **MUST require a separate, explicit confirmation before the actual deploy trigger (Phase 5)** — distinct from the secrets-setup confirmation in Phase 4. Setting up config and pulling the trigger are two different real actions against two different pieces of infrastructure.
- **MUST use only `cs` CLI top-level commands confirmed in `references/cli-and-api.md`'s Known Commands table** (as of 2026-07-29 that's the full set including `wake-up`, `git`, `sync`, `curl`, `generate`, `scale`, `licenses`, `mcp`, `completion` — reconfirmed directly against real `cs --help` output, not just older docs). Their **exact sub-arguments** (e.g. what `cs git`'s or `cs sync`'s own subcommands/flags are) are *not* shown by that table — run `cs <command> --help` against the actual target before relying on precise syntax for a multi-level command, and prefer the documented Public API endpoint from `references/cli-and-api.md` when a CLI form's exact syntax isn't already confirmed.
- **MUST NOT report a deployment as healthy based on `status`/`pipeline` state alone when the `ci.yml` has a publicly-routed service.** `GET /workspaces/{workspaceId}/status` (`isRunning`) and a `running`/`success` pipeline state only confirm the container processes are alive — they say nothing about whether the Workspace Router actually serves the app to an outside caller. **Confirmed live:** a workspace whose `restricted` flag is `true` — which is what `POST /workspaces` defaults to unless the request body explicitly sets `"restricted": false` (CLI equivalent: `cs create workspace` without `--public-dev-domain`) — 303-redirects every request on its dev domain to the IDE sign-in page, even for a service with `isPublic: true` and a working `network.paths` route, while `status`/`pipeline` both report fully healthy throughout. See Phase 6 for the required check.

## Process: 8-Phase Workflow

**Phase [N]: [Action Title]**
- **Prerequisite:** What must be true before this phase
- **Blocker risk:** What could fail and why
- **Action:** What you do (commands, decision points)
- **Output:** What success looks like
- **Blocking conditions:** If X or Y, stop and inform user

### Phase 0: Preflight

**Prerequisite:** Skill invoked.

**Blocker risk:** `cs` CLI missing; no valid `CS_TOKEN`; no `ci.yml` to deploy at all.

**Action:**
1. `command -v cs` — confirm the CLI is available.
2. Confirm `CS_TOKEN` is set (env var, or however the target environment provides it) and validate it against the Codesphere API (e.g. `GET /workspaces/team/{teamId}`, per `references/cli-and-api.md`) rather than just checking it's non-empty.
3. Confirm `ci.yml` exists at the repository root.

**Output:** Confirmed working CLI, valid token, confirmed `ci.yml` to deploy.

**Blocking conditions:**
- `cs` not found → stop: "cs CLI nicht gefunden. Bitte installieren, siehe docs.codesphere.com."
- `CS_TOKEN` missing or invalid → stop: name exactly what's wrong (missing vs. rejected by the API), never echo the token value itself even partially.
- No `ci.yml` at the root → stop, point at the generation skills (see Hard Gate).

### Phase 1: Determine the target workspace

**Prerequisite:** Phase 0 completed.

**Blocker risk:** Multiple workspaces plausibly match the repo; none do; the user has no workspace at all yet.

**Action:** Check for a local tracking file (e.g. `.codesphere/codesphere-deploy.json`) with a remembered `workspaceId` from a prior run of this skill — reuse it if present rather than re-asking. If not present, list the team's workspaces (`GET /workspaces/team/{teamId}`) and try to match by repo name (derived from `git remote get-url origin`).

- **Exactly one match** → use it.
- **No match** → ask the user directly for the workspace ID or name; this skill does not provision a new workspace itself.
- **Multiple matches** → list them and ask the user to pick.

Store the resolved `workspaceId` in the local tracking file for next time.

**Output:** Confirmed `workspaceId`.

**Blocking conditions:** Do not proceed to Phase 2 without a confirmed, single workspace.

### Phase 2: Parse `ci.yml` for required secrets and env vars

**Prerequisite:** Phase 1 completed.

**Blocker risk:** Missing a reference because it's nested somewhere unexpected (a `--set` override string, a `command:` line) rather than in an obvious `env:`/`secrets:` block.

**Action:** Scan the entire file content, not just the obvious `env:`/`secrets:` blocks — a `${{ vault.NAME }}` or `${{ workspace.env['KEY'] }}` reference can appear inline inside any string value, including a `command:` line (see `references/secret-management.md`'s Template Syntax section for the confirmed reference forms). Build two lists:
- Every distinct `${{ vault.NAME }}` key referenced anywhere.
- Every distinct `${{ workspace.env['KEY'] }}` key referenced anywhere.

Plain literal `env:` values (not template references) need nothing set — skip those.

**Output:** Two lists: required vault keys, required plain env var names.

**Blocking conditions:** None.

### Phase 3: Check what's already set

**Prerequisite:** Phase 2 produced the required-keys lists.

**Blocker risk:** Checking the wrong vault partition — a workspace assigned to a shared vault resolves `${{ vault.* }}` against that shared vault, not its own, per `references/secret-management.md`.

**Action:** Determine whether the workspace uses its own vault partition or an assigned shared vault, then list existing keys accordingly: `GET /vault/teams/{teamId}/workspaces/{workspaceId}/keys` (own partition) or `GET /vault/teams/{teamId}/shared/{vaultName}/keys` (shared vault) — either way, only key names are ever returned, never values. Also `GET /workspaces/{workspaceId}/env-vars` for plain env vars. Diff both against Phase 2's required lists.

**Output:** List of missing vault keys, list of missing plain env vars (each may be empty).

**Blocking conditions:** None.

### Phase 4: Resolve missing secrets/env vars

**Prerequisite:** Phase 3 completed.

**Blocker risk:** Silently generating a value for something that actually needed to be a specific external credential; silently proceeding to deploy with a required value still missing.

**Action:** If Phase 3 found nothing missing, skip straight to Phase 5.

Otherwise, name exactly what's missing (both lists) and ask: **"Sollen die fehlenden Secrets/Env-Vars jetzt konfiguriert werden?"**

- **Yes:**
  - For each missing vault key: offer to auto-generate (`POST /vault/teams/{teamId}/workspaces/{workspaceId}/generate`, per `references/secret-management.md` — a single call, the plaintext never has to pass through this conversation) as the default, or let the user supply their own value (`POST /vault/teams/{teamId}/workspaces/{workspaceId}` with `{"KEY": "value"}`) if they have a specific one in mind. Never print whatever value results.
  - For each missing plain env var: ask for the value directly — these are meaningful configuration (a URL, a flag), not generatable secrets, and their absence is not resolvable by auto-generation.
- **No:**
  - For missing vault keys with no indication of needing a specific external value, generate them anyway (this needs no input from the user, so a "no" to manual configuration doesn't have to block it — see Hard Gate).
  - For missing plain env vars, or any vault key that clearly needs a real external value the user declined to supply, **stop here** — deployment cannot proceed with a known-missing required value.

**Output:** Every required vault key and env var either confirmed present, freshly generated, or freshly supplied.

**Blocking conditions:** Do not proceed to Phase 5 with any required key still missing.

### Phase 5: Confirm and trigger deployment

**Prerequisite:** Phase 4 completed with nothing required still missing.

**Blocker risk:** Deploying stale code because the workspace's checkout is behind the branch that was actually pushed; treating an unconfirmed CLI shortcut as reliable.

**Action:** Ask for **explicit confirmation to deploy now** — separate from Phase 4's configuration confirmation. On yes:
1. Wake the workspace if it's an on-demand/off-when-unused one: `cs wake-up` (top-level command confirmed, exact flags not — check `cs wake-up --help` against the target, or use `GET /workspaces/{workspaceId}/status` first to see if this step is even needed).
2. Ensure the workspace has the latest code: `POST /workspaces/{workspaceId}/git/pull[/{remote}[/{branch}]]` (confirmed API endpoint), or the `cs git` CLI form if its exact pull syntax has been confirmed against the target (`cs git --help`).
3. Trigger the deployment: `POST /workspaces/{workspaceId}/landscape/deploy[/{profile}]` (confirmed API endpoint for a multi-service Landscape, which is what this family always generates) — or `cs start --stage prepare` then `test` then `run` in sequence (`cs start`'s exact stage-flag form is confirmed per `references/cli-and-api.md`). `cs sync` also exists as a top-level command and may be the more direct Landscape-specific form — confirm its exact subcommand (e.g. `landscape`) via `cs sync --help` against the target before relying on a specific invocation.

**Output:** Deployment triggered.

**Blocking conditions:**
- User declines the deploy confirmation → stop, report that configuration is ready but nothing was deployed, and how to trigger it later.
- Git pull or deploy trigger returns an error → stop, show the error text plainly, don't retry silently.

### Phase 6: Monitor status

**Prerequisite:** Phase 5 successfully triggered a deployment.

**Blocker risk:** Reporting "done" before the pipeline has actually finished; polling forever if something is genuinely stuck.

**Action:** Poll `GET /workspaces/{workspaceId}/status` and/or `GET /workspaces/{workspaceId}/pipeline/run` (confirmed API endpoints) at a reasonable interval until a terminal state is reached or a sensible timeout elapses. Report the actual state observed — don't infer "healthy" from the absence of an error.

Once the pipeline reports running/success, that alone is **not** sufficient to call the deployment healthy if the `ci.yml` has any service with `isPublic: true` or a `network.paths` route — per the Hard Gate above, `status`/`pipeline` never confirm actual public reachability. For every such service:
1. `GET /workspaces/{workspaceId}` and check `restricted` — if `true`, the dev domain will redirect to the IDE sign-in page instead of serving the app; flag this to the user and ask whether to fix it (`PATCH /workspaces/{workspaceId}` with `{"restricted": false}`, or the `cs-go` CLI equivalent if confirmed) before calling the deployment done.
2. Do an actual HTTP request against the service's own path on the workspace's `devDomain` (from the same `GET /workspaces/{workspaceId}` response, or `cs curl` if its exact invocation is confirmed via `cs curl --help`) and check for a real `2xx` from the application itself — a `3xx` redirect to `/ide/signIn` looks superficially like a response but is not the app. Use this as the application-level health check the Public API's own status endpoints can't provide.

**Output:** Final observed pipeline/workspace status, plus confirmed actual public reachability (or the specific reason it isn't reachable yet) for every publicly-routed service.

**Blocking conditions:** If status can't be determined within a reasonable time, say so plainly and point the user at the Codesphere UI rather than guessing.

### Phase 7: Summary

**Prerequisite:** Phase 6 completed (or Phase 5 stopped before deploying).

**Action:** Summarize for the user:
- Which vault keys/env vars were found already set, generated, or user-supplied this run.
- Whether deployment was actually triggered, and the final status observed.
- For every publicly-routed service: the actual reachability result from Phase 6 (confirmed public, or blocked by `restricted: true`/something else) — not just the pipeline/status state.
- If not deployed (user declined in Phase 5, or a blocker stopped the run earlier): exactly what's still needed before it can be.

**Output:** One consolidated status message the user can act on.

**Blocking conditions:** None.

## Related

- `codesphere` — reference knowledge this skill reads from (loose coupling, read-only, never auto-invoked); specifically `references/cli-and-api.md`, `references/secret-management.md`, `references/environment-variables.md`, `references/landscape.md`
- `codesphere-create-cluster-deployment` / `codesphere-create-container-deployment` / `codesphere-create-reactive-deployment` / `codesphere-add-managed-service` — generate or edit the `ci.yml` this skill deploys; none of them perform the deploy step themselves
