# Babelfish (Codesphere Managed Service) Reference

> **Last updated:** 2026-07-28 · **Source:** https://docs.codesphere.com/managed-services/providers/babelfish

> The documentation changes regularly. If you are facing issues, please refer to the official documentation:
> <https://docs.codesphere.com/managed-services/providers/babelfish>

## Overview

A PostgreSQL-backed database endpoint with Microsoft TDS wire-protocol compatibility, for workloads/tooling that expect SQL Server semantics. **Preview** — not enabled by default, must be turned on by the operator; schema/plans/behavior may still change. Under the hood it is the same managed PostgreSQL setup as the [`postgres` provider](./postgresql.md) (CloudNativePG-powered) with the Babelfish for PostgreSQL extensions installed, exposing a TDS endpoint on port `1433`.

| Property                 | Value                                                  |
| ------------------------ | ------------------------------------------------------ |
| Provider name            | `babelfish`                                            |
| Category                 | `Database`                                             |
| Scope                    | `global`                                               |
| Team singleton           | `false`                                                |
| Pause support            | `true`                                                 |
| Backups / PITR           | `true` / `true`                                        |
| High availability        | Not available — single instance only                   |
| Automatic upgrades       | Never — manual minor version bump via `version` config |
| Automatic storage growth | Never — bump `storage` plan parameter manually         |

## Core Concepts

- **Pinned version pairs**: `version` config pins a compatible PostgreSQL + Babelfish pair, e.g. `17.6-5.3.0` = PostgreSQL 17.6 + Babelfish 5.3.0.
- **Shared foundation with `postgres`**: storage (Ceph block), networking, and the backup mechanism are identical to the PostgreSQL provider — see [postgresql.md](./postgresql.md#core-concepts) for details.
- **TDS endpoint**: SQL Server clients/drivers connect directly on port `1433`.
- **T-SQL dialect**: exposed by the Babelfish extension on top of standard PostgreSQL.

## API / Syntax

### Config Schema

| Name      | Type   | Required | Description                                                                                                                                          |
| --------- | ------ | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| `version` | string | No       | Combined PostgreSQL+Babelfish version. Default `17.6-5.3.0`. Allowed: `17.7-5.4.0`, `16.11-4.8.0`, `17.6-5.3.0`, `16.10-4.7.0`. Minor upgrades only. |

### Secrets Schema

| Name                | Type   | Required | Description                                    |
| ------------------- | ------ | -------- | ---------------------------------------------- |
| `superuserPassword` | string | Yes      | Password for the administrative database user. |

### Details / Output Schema

| Name       | Type    | Availability       | Description                                                            |
| ---------- | ------- | ------------------ | ---------------------------------------------------------------------- |
| `hostname` | string  | after provisioning | Internal service hostname.                                             |
| `port`     | integer | after provisioning | TDS port (`1433`).                                                     |
| `dsn`      | string  | after provisioning | TDS connection string for the superuser against the `master` database. |
| `ready`    | boolean | after provisioning | Whether the instance accepts connections.                              |

### Plan: `Small` (`id: 0`)

| Name      | Type    | Default | Min   | Max | Static | Description              |
| --------- | ------- | ------- | ----- | --- | ------ | ------------------------ |
| `cpu`     | number  | `1`     | -     | -   | Yes    | Priced as `cpu-tenths`.  |
| `memory`  | integer | `128`   | -     | -   | Yes    | Priced as `ram-mib`.     |
| `storage` | integer | `1024`  | `512` | -   | No     | Priced as `storage-mib`. |

### Landscape Example

- **Example:**

```yaml
schemaVersion: v0.4
run:
  sqlserver-compatible-db:
    provider:
      name: babelfish
      schemaVersion: v1
    plan:
      id: 0
      parameters:
        storage: 2048
    config:
      version: "17.7-5.4.0"
    secrets:
      superuserPassword: "${{ vault.babelfishSuperuserPassword }}"
```

### Connecting — Terminal (`tsql`, FreeTDS)

- **Example:**

```bash
# Install FreeTDS
nix-env -iA nixpkgs.freetds
# Syntax
TDSENCRYPTION=required tsql -H <hostname> -p 1433 -U <username> -D <database>
# Example
TDSENCRYPTION=required tsql -H ms-babelfish-v1-123-my-server.ms-postgres -p 1433 -U postgres -D master
```

### Connecting — Node.js (`mssql`)

- **Example:**

```javascript
const sql = require("mssql");
(async () => {
  const config = {
    user: "postgres",
    password: "superSecret",
    server: "ms-babelfish-v1-123-my-server.ms-postgres",
    port: 1433,
    database: "master",
    options: { encrypt: true, trustServerCertificate: false },
  };
  await sql.connect(config);
  const result = await sql.query`SELECT 1 as val`;
  console.log(result.recordset[0].val);
})();
```

### Backups

- **Description:** Same mechanism as PostgreSQL — CloudNativePG Barman Cloud plugin stores native PostgreSQL physical backups + continuous WAL archiving to an S3-compatible store, enabling PITR. See [postgresql.md](./postgresql.md#patch-managed-servicesid--enable-backups) for the full config/recovery example payloads (identical shape, swap `provider.name` to `babelfish`).

## Common Pitfalls

- Assuming this is a native SQL Server engine — it's PostgreSQL with a T-SQL/TDS compatibility layer; not every SQL Server feature/behavior is guaranteed to match.
- Expecting HA — single instance only, same as `postgres`.
- Forgetting the preview flag — the provider must be explicitly enabled by the operator before it's usable on a given installation.
- Using the PostgreSQL wire protocol/port instead of the TDS endpoint (`1433`) when connecting from SQL Server tooling.

## Known Documentation Discrepancies

> This section may be outdated. Please verify against the official documentation:
> <https://docs.codesphere.com/managed-services/providers/babelfish>

- Allowed `version` pairs are a point-in-time snapshot; as PostgreSQL/Babelfish upstream releases new minors, the allowed-values list will change.
- Being a preview feature, capabilities (e.g. HA support) are explicitly called out as likely to change — re-check `capabilities` via `GET /api/managed-services/providers` before relying on this file.

## Further Reading

- Official docs: <https://docs.codesphere.com/managed-services/providers/babelfish>
- PostgreSQL provider reference (shared foundation): [postgresql.md](./postgresql.md)
- Babelfish for PostgreSQL project: <https://babelfishpg.org/>
- CloudNativePG: <https://cloudnative-pg.io/>
- `mssql` Node.js driver: <https://www.npmjs.com/package/mssql>
