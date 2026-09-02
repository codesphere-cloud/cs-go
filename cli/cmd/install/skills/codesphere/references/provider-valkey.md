# Valkey (Codesphere Managed Service) Reference

> **Last updated:** 2026-07-28 · **Source:** https://docs.codesphere.com/managed-services/providers/valkey

> The documentation changes regularly. If you are facing issues, please refer to the official documentation:
> <https://docs.codesphere.com/managed-services/providers/valkey>

## Overview

A high-performance in-memory key-value store suited for caching, transient state, queues, and lightweight messaging patterns. **Closed testing** — early testing state, only available on dedicated (self-hosted) installations. Schema/plans/behavior may still change.

| Property       | Value                                         |
| -------------- | --------------------------------------------- |
| Provider name  | `valkey`                                      |
| Category       | `Key-Value Store`                             |
| Scope          | `global`                                      |
| Team singleton | `false`                                       |
| Pause support  | `false`                                       |
| Availability   | Dedicated installations only (closed testing) |

## Core Concepts

- **Redis-protocol compatible**: Valkey is a Redis fork; any Redis-compatible or Valkey-native client works against `hostname`/`port`.
- **No AMQP/HTTP layer**: unlike RabbitMQ/OpenSearch, this is a single protocol/port service — no separate management endpoint is exposed in `details`.

## API / Syntax

### Config Schema

| Name      | Type   | Required | Description                                                   |
| --------- | ------ | -------- | ------------------------------------------------------------- |
| `version` | string | No       | Valkey engine version. Default and only allowed value: `9.0`. |

### Secrets Schema

| Name                | Type   | Required | Description                           |
| ------------------- | ------ | -------- | ------------------------------------- |
| `superuserPassword` | string | Yes      | Password for the administrative user. |

### Details / Output Schema

| Name       | Type    | Availability       | Description                               |
| ---------- | ------- | ------------------ | ----------------------------------------- |
| `hostname` | string  | after provisioning | Internal service hostname.                |
| `port`     | integer | after provisioning | Valkey service port.                      |
| `ready`    | boolean | after provisioning | Whether the instance accepts connections. |

### Plans

| Plan   | id  |
| ------ | --- |
| Small  | 0   |
| Medium | 1   |
| Large  | 2   |

Parameters for the `Small` (`id: 0`) example plan:

| Name      | Type    | Default | Min | Max | Static | Description              |
| --------- | ------- | ------- | --- | --- | ------ | ------------------------ |
| `storage` | integer | `1024`  | -   | -   | No     | Priced as `storage-mib`. |
| `cpu`     | number  | `5`     | -   | -   | No     | Priced as `cpu-tenths`.  |
| `memory`  | integer | `512`   | -   | -   | No     | Priced as `ram-mib`.     |

### Landscape Example

- **Example:**

```yaml
schemaVersion: v0.4
run:
  cache:
    provider:
      name: valkey
      schemaVersion: v0
    plan:
      id: 0
      parameters:
        storage: 1024
        cpu: 5
        memory: 512
    config:
      version: "9.0"
    secrets:
      superuserPassword: "${{ vault.valkeySuperuserPassword }}"
```

### Connecting

- **Description:** Other runtimes connect to `hostname`/`port` with any Redis-compatible or Valkey-native client, authenticating with the stored `superuserPassword`. Reachable from other Codesphere runtimes (reactives, managed containers, Virtual Cluster workloads).

## Common Pitfalls

- Trying to deploy this provider on shared/cloud Codesphere — currently closed-testing on dedicated installations only.
- Hardcoding `version: "9.0"` assuming other versions are selectable — it is currently the only allowed value.
- Expecting a persistence/backup capability like the database providers — none is documented for Valkey; treat cached data as ephemeral/reconstructable.
- Trying to pause the instance — not supported for this provider.

## Known Documentation Discrepancies

> This section may be outdated. Please verify against the official documentation:
> <https://docs.codesphere.com/managed-services/providers/valkey>

- As a closed-testing provider, `Medium`/`Large` plan parameter values are not fully enumerated beyond the `Small` example in the source docs — confirm via `GET /api/managed-services/providers` on the target dedicated installation.
- Backup/capabilities table is not present in the source docs for this provider (unlike PostgreSQL/Babelfish/DocumentDB) — confirm whether backups are genuinely unsupported or simply undocumented via `capabilities` on the live API.

## Further Reading

- Official docs: <https://docs.codesphere.com/managed-services/providers/valkey>
- Valkey project: <https://valkey.io/>
