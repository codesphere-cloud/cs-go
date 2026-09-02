# Virtual Kubernetes Cluster (Codesphere Managed Service) Reference

> **Last updated:** 2026-07-28 · **Source:** https://docs.codesphere.com/managed-services/providers/virtual-kubernetes-cluster

> The documentation changes regularly. If you are facing issues, please refer to the official documentation:
> <https://docs.codesphere.com/managed-services/providers/virtual-kubernetes-cluster>

## Overview

Gives a team a managed Kubernetes control plane inside Codesphere for cloud-native workloads, Helm-based deployments, and advanced orchestration. Listed as an `Advanced Compute` category service with a single fully-customizable plan. **Preview** — not enabled by default, must be turned on by the operator.

| Property       | Value                     |
| -------------- | ------------------------- |
| Provider name  | `virtual-k8s`             |
| Category       | `Advanced Compute`        |
| Scope          | `global`                  |
| Team singleton | `true` — **one per team** |
| Pause support  | `false`                   |

## Core Concepts

- **Team singleton**: unlike every other provider in this catalog, a team can only ever have one Virtual Kubernetes Cluster instance.
- **No provider-specific config schema**: access is entirely via the injected kubeconfig and Codesphere's team networking — there's nothing to set under `config`.
- **Indirect access pattern**: Workspaces/Landscapes use the injected `~/.kube/config` with `kubectl`/`helm` to deploy into the cluster; workloads running _inside_ the cluster reach other Codesphere resources over the same team network.

## API / Syntax

### Config Schema

- **Description:** None exposed — this provider has no provider-specific `config` fields.

### Secrets Schema

- **Description:** None documented for this provider.

### Plan: `Custom` (`id: 0`)

- **Description:** All resource limits are adjustable within provider-defined ranges.

| Name               | Type    | Default | Min     | Max      | Static | Description                                              |
| ------------------ | ------- | ------- | ------- | -------- | ------ | -------------------------------------------------------- |
| `cpu`              | integer | `20`    | `20`    | `160`    | No     | vCPU limit, priced as `cpu-tenths`.                      |
| `memory`           | integer | `5120`  | `5120`  | `32768`  | No     | RAM limit (MiB), priced as `ram-mib`.                    |
| `storage`          | integer | `20000` | `20000` | `120000` | No     | Persistent storage limit (MiB), priced as `storage-mib`. |
| `ephemeralStorage` | integer | `30000` | `30000` | `120000` | No     | Ephemeral storage limit (MiB), priced as `storage-mib`.  |

### Accessing the Cluster

- **Description:** Other runtimes typically access this provider indirectly rather than through details fields. Workspaces/Landscapes get `~/.kube/config` injected automatically and use standard `kubectl`/`helm` tooling; in-cluster workloads reach other Codesphere resources over the shared team network.
- **Further detail:** the runtime-specific URL patterns and full kubeconfig workflow are documented separately under Virtual Clusters — see Further Reading.

## Common Pitfalls

- Trying to create a second Virtual Kubernetes Cluster for the same team — blocked by the team-singleton constraint.
- Trying to pause the cluster to save cost — `pause` is not supported for this provider.
- Looking for a `config`/`secrets` schema to fill in on creation — there isn't one; everything is access-pattern based (kubeconfig injection).
- Setting `cpu`/`memory`/`storage`/`ephemeralStorage` outside the documented min/max ranges when scripting service creation.

## Known Documentation Discrepancies

> This section may be outdated. Please verify against the official documentation:
> <https://docs.codesphere.com/managed-services/providers/virtual-kubernetes-cluster>

- As a preview `Advanced Compute` service, plan ranges and the team-singleton constraint are explicitly called out as subject to change — re-verify via `GET /api/managed-services/providers` before relying on the min/max values above.
- Details/output schema for this provider is not documented in the scraped source (unlike the database providers, which expose `hostname`/`port`/`dsn`/`ready`); confirm whether any `details` fields are exposed via the public API before assuming there are none.

## Further Reading

- Official docs: <https://docs.codesphere.com/managed-services/providers/virtual-kubernetes-cluster>
- Virtual Clusters (runtime-specific URL patterns, kubeconfig workflow): <https://docs.codesphere.com/runtimes/virtual-clusters>
