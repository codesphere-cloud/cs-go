# Codesphere Deployment Guide Reference

> **Last updated:** 2026-07-28 · **Source:** https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/landscape-lifecycle

> The documentation changes regularly. If you are facing issues, please refer to the official documentation:
> <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/landscape-lifecycle>

## Overview

Covers the platform concepts around a deployed workspace/Landscape: resource plans, always-on vs. off-when-unused, custom domains, zero-downtime releases, horizontal scaling, and Landscape lifecycle operations (deploy/scale/teardown). Use this reference for "how does the platform behave" questions, as distinct from [ci-pipeline.md](./ci-pipeline.md) (the `ci.yml` syntax itself) and [landscape.md](./landscape.md) (multi-service architecture).

## Core Concepts

- **Workspace**: the fundamental unit — a cloud environment with code, IDE, terminal, CI pipeline, and deployment. Maps to a repository/branch. A workspace running a multi-service `ci.yml` is a **Landscape**.
- **Resource plans**: named plans (Micro/Boost/Pro) are used in the API and GitHub Actions integration; integer `plan:` ids are used in `ci.yml`. Higher id ≈ more resources, but the exact id→CPU/RAM/storage mapping is cluster-specific.
- **Always On vs. Off When Unused**: two deployment modes with very different pricing/availability tradeoffs (see below).
- **Zero-downtime release**: achieved via 2+ workspaces + a custom domain + a connection swap, not via any built-in blue/green primitive on a single workspace.

## API / Syntax

### Deployment Modes

| Mode                          | Behavior                                                                                                                                                                                                     |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Always On**                 | Dedicated resources continuously available; full price; required for production and for any workspace/service using replicas.                                                                                |
| **Off When Unused (standby)** | Resources released to the pool after ~60 min idle; ~10% of always-on pricing; reactivates in ~1s on domain visit or IDE access (auto re-runs the `run` stage); free-plan workspaces are always in this mode. |

### `POST /workspaces/{id}/landscape/deploy[/{profile}]`

- **Description:** Deploys the Landscape, optionally a specific CI profile.
- **Parameters:**

| Name      | Type | Required | Description                                              |
| --------- | ---- | -------- | -------------------------------------------------------- |
| `id`      | path | Yes      | Workspace ID.                                            |
| `profile` | path | No       | CI profile name (`ci.<profile>.yml`); omit for `ci.yml`. |

- **Example:**

```bash
curl -X POST "https://codesphere.com/api/workspaces/456/landscape/deploy/production" \
  -H "Authorization: Bearer <token>"
```

### `PATCH /workspaces/{id}/landscape/scale`

- **Description:** Scale named services' replica counts programmatically.
- **Parameters:**

| Name | Type   | Required | Description                                  |
| ---- | ------ | -------- | -------------------------------------------- |
| `id` | path   | Yes      | Workspace ID.                                |
| body | object | Yes      | Map of `{"serviceName": replicaCount, ...}`. |

- **Example:**

```bash
curl -X PATCH "https://codesphere.com/api/workspaces/456/landscape/scale" \
  -H "Authorization: Bearer <token>" -H "Content-Type: application/json" \
  -d '{"myServer1": 3, "myServer2": 2}'
```

### `DELETE /workspaces/{id}/landscape/teardown`

- **Description:** Tears down the Landscape and its Landscape-scoped Managed Services.
- **Parameters:**

| Name | Type | Required | Description   |
| ---- | ---- | -------- | ------------- |
| `id` | path | Yes      | Workspace ID. |

- **Example:**

```bash
curl -X DELETE "https://codesphere.com/api/workspaces/456/landscape/teardown" \
  -H "Authorization: Bearer <token>"
```

### Horizontal Scaling — Rules

- **Description:** Constraints and behavior when using `replicas` on a Reactive/Managed Container.

| Rule                 | Detail                                                                                                                                               |
| -------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| Max replicas         | Up to 10 per service (more on enterprise); requires Always On.                                                                                       |
| Filesystem           | Shared `/home/user/app` across replicas — SSD storage isn't duplicated per replica.                                                                  |
| Stage execution      | `prepare`/`test` run once on the main replica; `run` executes on every replica.                                                                      |
| Pricing              | Per-replica.                                                                                                                                         |
| Autoscaling          | Set min/max replicas; Codesphere scales on CPU usage.                                                                                                |
| Per-replica identity | Use the `CS_REPLICA` env var to segregate per-replica writes (e.g. log file names); avoid concurrent writes to the same file from multiple replicas. |

### Custom Domains — Workflow

1. Add an **A record** pointing at Codesphere's IP.
2. Add a **TXT record** for verification.
3. Verify: UI, or `POST /domains/team/{teamId}/domain/{domainName}/verify`.
4. Connect one or more workspaces: `PUT /domains/team/{teamId}/domain/{domainName}/workspace-connections`.

- **Note:** connecting multiple workspaces to one domain enables automatic load balancing, A/B testing, and blue/green deployments. Path-based routing under a custom domain works the same as within a Landscape — apps must serve from their assigned path prefix unless the router strips it; regex paths are supported via a UI checkbox.

### Zero-Downtime Release — Procedure

1. Deploy the new version to a staging workspace.
2. Test it.
3. Swap the domain connection from production to staging.
4. Instant rollback by reconnecting the previous workspace.

- **Requires:** 2+ workspaces + a custom domain.

## Common Pitfalls

- Using replicas on an Off-When-Unused workspace — replicas require Always On.
- Writing to the same log/output file from multiple replicas without using `CS_REPLICA` to disambiguate — causes concurrent-write conflicts on the shared filesystem.
- Assuming a fixed `plan:` id → CPU/RAM table — it's cluster-specific; resolve via `GET /metadata/workspace-plans`.
- Expecting a single workspace to support blue/green natively — zero-downtime release requires 2+ workspaces plus a custom domain and a connection swap.
- Forgetting `.codesphere-internal/` in `.gitignore` — causes deployment/Landscape scan failures (see Troubleshooting below).

## Troubleshooting

| Symptom                                 | Likely cause                                                                             | Fix                                                                               |
| --------------------------------------- | ---------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------- |
| App not accessible                      | Not bound to `0.0.0.0:<port>`, or `healthEndpoint` unreachable from inside the container | Check bind address and `healthEndpoint` target (localhost unless bound elsewhere) |
| Packages disappear after restart        | Installed outside persistent paths                                                       | Install to `/home/user/app` or via Nix                                            |
| Permission denied                       | Trying root/sudo                                                                         | Use Nix instead                                                                   |
| Workspace stuck rebooting               | Invalid `ci.yml`                                                                         | Check syntax, `schemaVersion`, and named services under `run`                     |
| GitHub repos missing                    | OAuth token expired                                                                      | Re-grant repository access                                                        |
| Env var changes not applying            | `run` stage not restarted                                                                | Re-run the `run` stage                                                            |
| Deployment fails / Landscape won't scan | Missing `.codesphere-internal/` in `.gitignore`                                          | Add it                                                                            |
| Pipeline not parsing                    | Missing `schemaVersion` as first line, or flat `run.steps[]`                             | Fix the header; use named `run.<service>.steps[]`                                 |
| Landscape sync fails on secrets         | A referenced `${{ vault.NAME }}` was never initialized                                   | Provide the value during sync, or pre-populate via the vault API                  |
| Managed service create/update rejected  | Wrong provider name/version/plan for this installation                                   | `GET /managed-services/providers` first, don't assume                             |

## Known Documentation Discrepancies

> This section may be outdated. Please verify against the official documentation:
> <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/landscape-lifecycle>

- The "Marketplace" framing (fixed MySQL/Redis/PostgreSQL/MongoDB options) referenced in older docs is superseded by the Managed Services module — the current provider catalog is API-discoverable and installation-specific, not a fixed list. Always check `GET /managed-services/providers`.
- Exact resource-plan pricing and the id→CPU/RAM table are intentionally not reproduced here since they're cluster-specific — resolve via `GET /metadata/workspace-plans`.

## Further Reading

- Official docs: <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/landscape-lifecycle>
- CI pipeline field reference: [ci-pipeline.md](./ci-pipeline.md)
- Landscape networking & multi-service patterns: [landscape.md](./landscape.md)
- CLI & API reference: [cli-and-api.md](./cli-and-api.md)
- Managed services architecture: `../managed-services/README.md`
