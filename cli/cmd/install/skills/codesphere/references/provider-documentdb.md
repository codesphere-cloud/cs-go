# Codesphere DocumentDB (Managed Service) Reference

> **Last updated:** 2026-07-28 · **Source:** https://docs.codesphere.com/managed-services/providers/documentdb

> The documentation changes regularly. If you are facing issues, please refer to the official documentation:
> <https://docs.codesphere.com/managed-services/providers/documentdb>

## Overview

A MongoDB-compatible document database powered by PostgreSQL, the DocumentDB extension, and a FerretDB proxy. For document-oriented applications that want MongoDB protocol compatibility inside Codesphere. **Preview** — not enabled by default, must be turned on by the operator; schema/plans/behavior may still change.

| Property                 | Value                                                |
| ------------------------ | ---------------------------------------------------- |
| Provider name            | `ferretdb`                                           |
| Category                 | `Database`                                           |
| Scope                    | `global`                                             |
| Team singleton           | `false`                                              |
| Pause support            | `true`                                               |
| Backups / PITR           | `true` / `true` (PostgreSQL-format, not `mongodump`) |
| High availability        | Not available — single instance only                 |
| Automatic upgrades       | Never                                                |
| Automatic storage growth | Never — bump `storage` plan parameter manually       |

## Core Concepts

- **Two-component architecture**: (1) PostgreSQL + the open-source [DocumentDB extension](https://github.com/documentdb/documentdb) — BSON storage/processing inside PostgreSQL, on a CloudNativePG-managed server; (2) a stateless [FerretDB](https://www.ferretdb.io/) proxy speaking the MongoDB wire protocol (MongoDB 5.0+ driver compatible) in front of it.
- **Two endpoints**: `dsn` (FerretDB/MongoDB protocol, port `27017`) for application traffic; `postgresDSN` for direct PostgreSQL access (migrations, maintenance, advanced extensions).
- **`ready` covers both components**: only `true` once both FerretDB and PostgreSQL are up.
- **No write buffering**: the proxy does not queue requests — if PostgreSQL is unavailable, reads/writes fail immediately rather than being retried.
- **Partial MongoDB compatibility**: FerretDB supports most but not all commands — e.g. change streams and multi-document transactions are **not** available. See the [FerretDB compatibility docs](https://docs.ferretdb.io/migration/compatibility/).
- **Pinned versions**: `version` config pins PostgreSQL major + DocumentDB extension + FerretDB versions together, e.g. `17-0.107.0-ferretdb-2.7.0`.

## API / Syntax

### Config Schema

| Name      | Type   | Required | Description                                                                                                       |
| --------- | ------ | -------- | ----------------------------------------------------------------------------------------------------------------- |
| `version` | string | No       | Combined Postgres/DocumentDB/FerretDB stack version. Default and only allowed value: `17-0.107.0-ferretdb-2.7.0`. |

### Secrets Schema

| Name                | Type   | Required | Description                           |
| ------------------- | ------ | -------- | ------------------------------------- |
| `superuserPassword` | string | Yes      | Password for the administrative user. |

### Details / Output Schema

| Name               | Type    | Availability       | Description                                              |
| ------------------ | ------- | ------------------ | -------------------------------------------------------- |
| `hostname`         | string  | after provisioning | Internal hostname of the MongoDB-compatible endpoint.    |
| `port`             | integer | after provisioning | MongoDB-compatible port (`27017`).                       |
| `dsn`              | string  | after provisioning | MongoDB connection string for the admin user.            |
| `postgresHostname` | string  | after provisioning | Internal hostname of the underlying PostgreSQL instance. |
| `postgresDSN`      | string  | after provisioning | PostgreSQL connection string for direct access.          |
| `ready`            | boolean | after provisioning | True once both FerretDB and PostgreSQL are ready.        |

### Plan: `Small` (`id: 0`)

| Name             | Type    | Default | Min    | Max   | Static | Description                                        |
| ---------------- | ------- | ------- | ------ | ----- | ------ | -------------------------------------------------- |
| `cpu`            | number  | `1`     | -      | -     | Yes    | PostgreSQL CPU, priced as `cpu-tenths`.            |
| `memory`         | integer | `512`   | -      | -     | Yes    | PostgreSQL memory (MiB), priced as `ram-mib`.      |
| `storage`        | integer | `1024`  | `1024` | -     | No     | Persistent storage (MiB), priced as `storage-mib`. |
| `ferretdbCpu`    | number  | `1`     | `1`    | `1`   | Yes    | FerretDB proxy CPU, priced as `cpu-tenths`.        |
| `ferretdbMemory` | integer | `64`    | `64`   | `128` | Yes    | FerretDB proxy memory (MiB), priced as `ram-mib`.  |

### Landscape Example

- **Example:**

```yaml
schemaVersion: v0.4
run:
  docs-db:
    provider:
      name: ferretdb
      schemaVersion: v0
    plan:
      id: 0
      parameters:
        storage: 2048
        memory: 512
        cpu: 1
        ferretdbMemory: 64
        ferretdbCpu: 1
    config:
      version: "17-0.107.0-ferretdb-2.7.0"
    secrets:
      superuserPassword: "${{ vault.documentDbSuperuserPassword }}"
```

### Constructing Hostnames

- **Description:** Deterministic pattern: `ms-{providerName}-{providerVersion}-{teamId}-landscape-{workspaceId}-{serviceName}.ms-postgres`, lowercased, invalid chars → hyphens. For `ferretdb`/`v0`, service `db`, team `42`, workspace `100`: `ms-ferretdb-v0-42-landscape-100-db.ms-postgres`. The underlying `postgres`/`v1` endpoint follows the same pattern.
- **Example (shell):**

```bash
mongosh "mongodb://postgres:${DOCDB_PASSWORD}@ms-ferretdb-v0-$TEAM_ID-landscape-$WORKSPACE_ID-db.ms-postgres:27017/ferretdb?authSource=postgres"
```

- **Example (landscape template, direct PostgreSQL access):**

```yaml
env:
  DOCDB_POSTGRES_HOST: ms-postgres-v1-${{ team.id }}-landscape-${{ workspace.id }}-db.ms-postgres
  DOCDB_PASSWORD: "${{ vault.documentDbSuperuserPassword }}"
```

### Connecting — Terminal (`mongosh`)

- **Note:** No TLS encryption between client and the FerretDB proxy.
- **Example:**

```bash
# Install mongosh
nix-env -iA nixpkgs.mongosh
source ~/.nix-profile/etc/profile.d/nix.sh
# Syntax
mongosh "mongodb://<username>:<password>@<hostname>:27017/ferretdb?authSource=postgres"
# Example
db.myCollection.insertOne({ message: "Hello Codesphere!" })
```

### Connecting — Node.js (`mongodb` driver)

- **Example:**

```javascript
const { MongoClient } = require("mongodb");
const client = new MongoClient(
  "mongodb://postgres:secure-password@ms-ferretdb-v0-4005-my-server.ms-postgres:27017/ferretdb",
);
await client.connect();
const db = client.db("ferretdb");
const collection = db.collection("myCollection");
const result = await collection.insertOne({ message: "Hello Codesphere!" });
console.log(result.insertedId);
await client.close();
```

### Backups

- **Description:** Data engine is PostgreSQL, so backups use the same CloudNativePG Barman Cloud plugin mechanism as the `postgres` provider (physical backups + WAL archiving, PITR-capable). They are **PostgreSQL backups, not MongoDB dumps** — restorable into a new DocumentDB service, but not interchangeable with `mongodump` archives. For native Mongo-format dumps, run `mongodump`/`mongorestore` manually against the service endpoint — this is not automated by Codesphere. Config/recovery payload shape is identical to [postgresql.md](./postgresql.md#patch-managed-servicesid--enable-backups).

## Common Pitfalls

- Relying on change streams or multi-document transactions — not supported by the FerretDB compatibility layer.
- Expecting backups to be `mongodump`-compatible — they are PostgreSQL physical backups.
- Expecting writes to queue during a PostgreSQL outage — they fail immediately, no buffering.
- Using the `postgresDSN` for application traffic instead of the MongoDB-compatible `dsn` (or vice versa for migrations).
- Assuming TLS is active on the client↔FerretDB hop when connecting via `mongosh` from a workspace terminal.

## Known Documentation Discrepancies

> This section may be outdated. Please verify against the official documentation:
> <https://docs.codesphere.com/managed-services/providers/documentdb>

- `version` currently has exactly one allowed value (`17-0.107.0-ferretdb-2.7.0`); expect this list to grow as the preview matures — re-check before hardcoding.
- FerretDB's MongoDB-compatibility surface changes across FerretDB releases; the "not available" list (change streams, multi-doc transactions) reflects FerretDB 2.7.0 at time of writing — verify against the linked FerretDB compatibility docs for the pinned version in use.

## Further Reading

- Official docs: <https://docs.codesphere.com/managed-services/providers/documentdb>
- PostgreSQL provider reference (shared backup mechanism): [postgresql.md](./postgresql.md)
- DocumentDB extension: <https://github.com/documentdb/documentdb>
- FerretDB: <https://www.ferretdb.io/>
- FerretDB compatibility matrix: <https://docs.ferretdb.io/migration/compatibility/>
- `mongodb` Node.js driver: <https://www.npmjs.com/package/mongodb>
