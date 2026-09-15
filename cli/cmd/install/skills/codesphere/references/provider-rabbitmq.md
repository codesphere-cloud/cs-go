# RabbitMQ (Codesphere Managed Service) Reference

> **Last updated:** 2026-07-28 · **Source:** https://docs.codesphere.com/managed-services/providers/rabbitmq

> The documentation changes regularly. If you are facing issues, please refer to the official documentation:
> <https://docs.codesphere.com/managed-services/providers/rabbitmq>

## Overview

A message broker for queueing, pub/sub, and streaming-style application integration — asynchronous processing, worker pipelines, service decoupling. **Closed testing** — early testing state, only available on dedicated (self-hosted) installations. Schema/plans/behavior may still change.

| Property       | Value                                         |
| -------------- | --------------------------------------------- |
| Provider name  | `rabbitmq`                                    |
| Category       | `Message Queue`                               |
| Scope          | `global`                                      |
| Team singleton | `false`                                       |
| Pause support  | `false`                                       |
| Availability   | Dedicated installations only (closed testing) |

## Core Concepts

- **Fixed admin username**: the administrative user is always `admin`; only the password is configurable (as a secret).
- **AMQP + Management UI dual ports**: `port` (AMQP, typically `5672`) for client connections, `managementPort` (typically `15672`) for the management UI/API.
- **Node count via `replicas`**: cluster size is a plan parameter, not a separate scaling operation.

## API / Syntax

### Config Schema

| Name      | Type   | Required | Description                                                     |
| --------- | ------ | -------- | --------------------------------------------------------------- |
| `version` | string | No       | RabbitMQ server version. Default and only allowed value: `4.3`. |

### Secrets Schema

| Name                | Type   | Required | Description                                                       |
| ------------------- | ------ | -------- | ----------------------------------------------------------------- |
| `superuserPassword` | string | Yes      | Password for the administrative user. Username is always `admin`. |

### Details / Output Schema

| Name             | Type    | Availability       | Description                            |
| ---------------- | ------- | ------------------ | -------------------------------------- |
| `host`           | string  | after provisioning | Internal service hostname.             |
| `port`           | integer | after provisioning | AMQP port, typically `5672`.           |
| `managementPort` | integer | after provisioning | Management UI port, typically `15672`. |
| `replicas`       | integer | after provisioning | Reported number of replicas.           |
| `readyReplicas`  | integer | after provisioning | Number of replicas currently ready.    |
| `ready`          | boolean | after provisioning | Whether the cluster is ready.          |

### Plans

| Plan   | id  |
| ------ | --- |
| Small  | 0   |
| Medium | 1   |
| Large  | 2   |

Parameters for the `Small` (`id: 0`) example plan:

| Name       | Type    | Default | Min    | Max | Static | Description                        |
| ---------- | ------- | ------- | ------ | --- | ------ | ---------------------------------- |
| `storage`  | integer | `1024`  | `1024` | -   | No     | Priced as `storage-mib`.           |
| `cpu`      | number  | `5`     | -      | -   | No     | Priced as `cpu-tenths`.            |
| `memory`   | integer | `512`   | -      | -   | No     | Priced as `ram-mib`.               |
| `replicas` | integer | `1`     | -      | -   | No     | RabbitMQ node count for this plan. |

### Landscape Example

- **Example:**

```yaml
schemaVersion: v0.4
run:
  broker:
    provider:
      name: rabbitmq
      schemaVersion: v0
    plan:
      id: 0
      parameters:
        storage: 1024
        cpu: 5
        memory: 512
        replicas: 1
    config:
      version: "4.3"
    secrets:
      superuserPassword: "${{ vault.rabbitmqSuperuserPassword }}"
```

### Connecting

- **Description:** Other runtimes use `host` + `port` for AMQP client connections, and `managementPort` for the management UI/API. Authenticate as `admin` with the stored `superuserPassword`. Reachable from other Codesphere runtimes (reactives, managed containers, Virtual Cluster workloads).

## Common Pitfalls

- Trying to deploy this provider on shared/cloud Codesphere — currently closed-testing on dedicated installations only.
- Assuming the admin username is configurable — it's always `admin`; only the password is a secret field.
- Hardcoding `version: "4.3"` assuming other versions are selectable — it is currently the only allowed value.
- Confusing `port` (AMQP, `5672`) with `managementPort` (`15672`) when wiring up client libraries vs. the management UI.
- Trying to pause the broker — not supported for this provider.

## Known Documentation Discrepancies

> This section may be outdated. Please verify against the official documentation:
> <https://docs.codesphere.com/managed-services/providers/rabbitmq>

- As a closed-testing provider, `Medium`/`Large` plan parameter values are not fully enumerated beyond the `Small` example in the source docs — confirm via `GET /api/managed-services/providers` on the target dedicated installation.
- Default AMQP/management port numbers are documented as "typically" `5672`/`15672` — treat the `port`/`managementPort` detail fields as authoritative over the hardcoded defaults if they ever diverge.

## Further Reading

- Official docs: <https://docs.codesphere.com/managed-services/providers/rabbitmq>
- RabbitMQ project: <https://www.rabbitmq.com/>
