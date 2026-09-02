# OpenSearch (Codesphere Managed Service) Reference

> **Last updated:** 2026-07-28 · **Source:** https://docs.codesphere.com/managed-services/providers/opensearch

> The documentation changes regularly. If you are facing issues, please refer to the official documentation:
> <https://docs.codesphere.com/managed-services/providers/opensearch>

## Overview

A distributed search and analytics engine for full-text search, log analysis, and observability workloads. **Closed testing** — early testing state, only available on dedicated (self-hosted) installations, not on shared cloud. Schema/plans/behavior may still change.

| Property       | Value                                         |
| -------------- | --------------------------------------------- |
| Provider name  | `opensearch`                                  |
| Category       | `Search & Analytics`                          |
| Scope          | `global`                                      |
| Team singleton | `false`                                       |
| Pause support  | `false`                                       |
| Availability   | Dedicated installations only (closed testing) |

## Core Concepts

- **Clustered by default**: the example plan runs `replicas: 3` — this is a multi-node cluster, not a single instance like the database providers.
- **Two access surfaces**: the REST API (`host`/`port`) for application/ingest traffic, and OpenSearch Dashboards (`dashboardsUrl`) for the operator/analyst UI.
- **Readiness is replica-aware**: `ready` reflects overall cluster health; `readyReplicas` vs `replicas` lets you see partial-availability states during rollout/scaling.

## API / Syntax

### Config Schema

| Name      | Type   | Required | Description                                                        |
| --------- | ------ | -------- | ------------------------------------------------------------------ |
| `version` | string | No       | OpenSearch engine version. Default and only allowed value: `2.19`. |

### Secrets Schema

| Name                | Type   | Required | Description                           |
| ------------------- | ------ | -------- | ------------------------------------- |
| `superuserPassword` | string | Yes      | Password for the administrative user. |

### Details / Output Schema

| Name            | Type    | Availability       | Description                                |
| --------------- | ------- | ------------------ | ------------------------------------------ |
| `host`          | string  | after provisioning | Internal service hostname.                 |
| `port`          | integer | after provisioning | REST API port.                             |
| `dashboardsUrl` | string  | after provisioning | URL of the OpenSearch Dashboards endpoint. |
| `replicas`      | integer | after provisioning | Reported number of replicas.               |
| `readyReplicas` | integer | after provisioning | Number of replicas currently ready.        |
| `ready`         | boolean | after provisioning | Whether the cluster is ready.              |

### Plans

| Plan        | id  |
| ----------- | --- |
| Small       | 0   |
| Medium      | 1   |
| Large       | 2   |
| Extra Large | 3   |

Parameters for the `Small` (`id: 0`) example plan:

| Name       | Type    | Default | Min    | Max | Static | Description                        |
| ---------- | ------- | ------- | ------ | --- | ------ | ---------------------------------- |
| `storage`  | integer | `5120`  | `5120` | -   | No     | Priced as `storage-mib`.           |
| `cpu`      | number  | `5`     | -      | -   | No     | Priced as `cpu-tenths`.            |
| `memory`   | integer | `1024`  | `1024` | -   | No     | Priced as `ram-mib`.               |
| `replicas` | integer | `3`     | -      | -   | Yes    | Fixed replica count for this plan. |

### Landscape Example

- **Example:**

```yaml
schemaVersion: v0.4
run:
  search:
    provider:
      name: opensearch
      schemaVersion: v0
    plan:
      id: 0
      parameters:
        storage: 5120
        cpu: 5
        memory: 1024
        replicas: 3
    config:
      version: "2.19"
    secrets:
      superuserPassword: "${{ vault.opensearchSuperuserPassword }}"
```

### Connecting

- **Description:** Application runtimes connect over the REST API using `host`, `port`, and the stored superuser credentials. Operators/analysts open `dashboardsUrl` for the Dashboards UI. The same endpoint is reachable from other Codesphere runtimes (reactives, managed containers, Virtual Cluster workloads).

## Common Pitfalls

- Trying to deploy this provider on shared/cloud Codesphere — currently closed-testing on dedicated installations only.
- Assuming `ready: true` means every replica is up — check `readyReplicas` vs `replicas` for partial-availability states.
- Hardcoding `version: "2.19"` assuming other versions are selectable — it is currently the only allowed value.
- Trying to pause the cluster — not supported for this provider.

## Known Documentation Discrepancies

> This section may be outdated. Please verify against the official documentation:
> <https://docs.codesphere.com/managed-services/providers/opensearch>

- As a closed-testing provider, plan definitions (`Medium`/`Large`/`Extra Large` parameter ranges) are not fully enumerated in the source docs beyond the `Small` example — confirm exact values via `GET /api/managed-services/providers` on the target dedicated installation before generating configs for the larger plans.
- Availability (dedicated-only) may change if/when this provider graduates from closed testing to general/preview availability.

## Further Reading

- Official docs: <https://docs.codesphere.com/managed-services/providers/opensearch>
- OpenSearch project: <https://opensearch.org/>
