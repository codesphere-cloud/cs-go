---
name: codesphere-run-deployment
description: 'Deploys an existing ci.yml on Codesphere: checks every ${{ vault.NAME }} secret and ${{ workspace.env[...] }} var the file actually references, offers to configure what''s missing (auto-generating vault secrets where safe, asking for values the platform can''t invent), triggers the deployment, and reports pipeline/health status — including an actual reachability check for public services, not just pipeline state. Trigger for "jetzt deployen", "ci.yml ausrollen", "landscape syncen", "auf Codesphere deployen", or any request to actually run/publish a deployment rather than just generate its ci.yml.'
allowed-tools: Bash Read Write Glob Grep
metadata:
  version: "1.1.0"
  updated: "2026-09-23"
  cost-tier: high
---

> **Process:** When this skill is explicitly/directly invoked by name, execute it immediately — don't ask the user what they want done with it.

## When to use this

Trigger when the user wants an already-generated `ci.yml` actually deployed — "deploy das jetzt", "ci.yml ausrollen", "Landscape syncen". This is the one skill in the `codesphere` family that performs a real deployment action; `codesphere-create-cluster-deployment` / `-container-deployment` / `-reactive-deployment` and `codesphere-add-managed-service` all stop at generating/editing `ci.yml` and hand the deploy step to this skill.

## Reference

Read `references/cli-and-api.md` (CLI commands, API endpoints), `references/secret-management.md` (vault partitions, template syntax, `/generate`), `references/environment-variables.md` (`WORKSPACE_ID` resolution), and `references/landscape.md` — all inside the sibling `codesphere` skill (`references/skill-family-conventions.md` covers locating them without loading `codesphere` itself).

## Process

1. **Preflight.** Confirm `cs` is on `PATH`, `CS_TOKEN` is set and actually valid against the API (e.g. `GET /workspaces/team/{teamId}` — not just non-empty), and `ci.yml` exists at the repo root. Stop with the specific problem (never echo the token) if any of these fail; point at the generation skills if there's no `ci.yml`.
2. **Determine the target workspace.** If running inside a Codesphere workspace, resolve its own ID directly: `tmux show-environment -g WORKSPACE_ID` (`-g` required; `printenv WORKSPACE_ID` is empty over `workspace-ssh`/editor terminals; `HOSTNAME` is a UUID, not the numeric ID — never substitute it). Otherwise reuse a remembered `workspaceId` from a local tracking file if one exists, or list the team's workspaces and match by repo name (`git remote get-url origin`); ask the user to pick on ambiguity or on no match. **Never guess a workspace out of the listing** — an API token is usually team/org-wide, so the listing includes other people's workspaces, and a wrong pick deploys onto one of them. Store the resolved ID in the tracking file for next time.
3. **Parse `ci.yml` for required secrets/env vars.** Scan the entire file content, not just `env:`/`secrets:` blocks — a `${{ vault.NAME }}` or `${{ workspace.env['KEY'] }}` reference can appear inline in any string, including a `command:` line. Build the two distinct-key lists.
4. **Check what's already set.** Determine whether the workspace uses its own vault partition or an assigned shared vault (resolve per `references/secret-management.md`) and list existing key names there (never values), plus existing plain env vars. Diff both against step 3's lists.
5. **Resolve anything missing.** If nothing's missing, skip to step 6. Otherwise name exactly what's missing and ask whether to configure it now.
   - **Yes:** auto-generate each missing vault key with no external meaning via the vault's `/generate` endpoint by default (one call, plaintext never passes through the conversation), or take a specific value from the user if they have one in mind. Ask directly for each missing plain env var — these are real configuration, not generatable.
   - **No:** still auto-generate any vault key needing no external input (declining "configure now" doesn't have to block something that needs nothing from the user) — but **stop here** if a plain env var or an externally-meaningful vault key is still missing; don't proceed with a known-missing required value.
6. **Confirm and trigger the deploy**, with a confirmation **separate from** step 5's configuration one. On yes: wake the workspace if it's on-demand (`cs wake-up`, or check `GET /workspaces/{id}/status` first), ensure it has the latest code (`POST /workspaces/{id}/git/pull[...]`, or the confirmed `cs git` form), then trigger (`POST /workspaces/{id}/landscape/deploy[/{profile}]`, or `cs start --stage prepare`→`test`→`run` in sequence, or `cs sync` if its exact subcommand is confirmed against the target). If the pull or trigger errors, stop and show the error plainly — don't retry silently.
7. **Monitor status.** Poll `GET /workspaces/{id}/status` and/or `.../pipeline/run` until a terminal state or a sensible timeout; report what's actually observed. A `running`/`success` pipeline state is **not sufficient** to call the deployment healthy for any service with `isPublic: true` or a `network.paths` route: check `restricted` on `GET /workspaces/{id}` (defaults `true` unless the workspace was created with `"restricted": false`; when `true` it 303-redirects every request to the IDE sign-in page even with a working route, while `status`/`pipeline` still report fully healthy) and offer to fix it, then make an actual HTTP request against the service's own path on its `devDomain` and confirm a real `2xx` from the app — a sign-in redirect is not a healthy app.
8. **Summarize:** which vault keys/env vars were already set vs. generated vs. user-supplied; whether the deploy was actually triggered and its final status; the actual reachability result per publicly-routed service (not just pipeline state); and, if nothing was deployed, exactly what's still needed.

## Keep in mind

- This is the only skill in the family permitted to perform an active deployment — every sibling explicitly never deploys anything. Don't let that permission bleed into a sibling, and don't apply "never deploys" here.
- Never print a secret or token value in plain text anywhere — status messages, logs, summaries. Mask as `***`, including a value the user pastes into the conversation.
- Use only `cs` CLI top-level commands confirmed in `references/cli-and-api.md`'s Known Commands table. Their exact sub-arguments aren't covered there — run `cs <command> --help` against the real target before relying on precise multi-level syntax, and prefer the documented Public API endpoint when a CLI form isn't confirmed.
- Determine required secrets/env vars purely by parsing the actual `ci.yml` (step 3) — never guess from what a "typical" deployment needs.

## Related

- `codesphere` — reference knowledge this skill reads from (loose coupling, read-only, never auto-invoked): `references/cli-and-api.md`, `references/secret-management.md`, `references/environment-variables.md`, `references/landscape.md`
- `codesphere-create-cluster-deployment` / `-container-deployment` / `-reactive-deployment` / `codesphere-add-managed-service` — generate or edit the `ci.yml` this skill deploys; none of them perform the deploy step themselves
