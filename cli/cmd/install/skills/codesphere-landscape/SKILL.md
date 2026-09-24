---
name: codesphere-landscape
description: 'Use this skill when authoring or editing a Codesphere Landscape''s ci.yml by hand — adding, changing, or debugging a Reactive, Managed Container, or embedded Managed Service entry under run:. Provides the full annotated v0.4 schema fixture as the canonical safe starting template (the shorter per-feature doc snippets omit fields that are required in practice), two confirmed traps around Managed Container fields and runAsUser/runAsGroup not covering an image''s own WORKDIR, and what to do when `cs sync landscape` returns a bare 500 instead of a real validation error. Trigger on "ci.yml", "landscape deploy", "landscape/deploy 500", "Config Schema Violation", "cs sync landscape", adding a service to an existing ci.yml, or a schema-shaped 500 with no useful error body.'
license: none
allowed-tools: Read Write Glob Grep
metadata:
  version: "1.1.0"
  updated: "2026-09-23"
  cost-tier: low
---

> **Process:** Read this file in full before writing or editing a `ci.yml`. Start from the canonical template below and add/remove services — never from a blank file or a documentation snippet.

## When to use this

Trigger when hand-authoring or hand-editing a Landscape's `ci.yml` — adding a service to an existing Landscape, writing one from scratch, or debugging a failing `cs sync landscape` / `POST /workspaces/{id}/landscape/deploy` call. This is specifically the **hand-authoring trap**: the public docs' shorter per-feature snippets look complete and parse as valid YAML, but omit fields the real backend validator requires, and the API's error response gives nothing to go on when that happens. It complements the rest of the `codesphere` family — `codesphere` holds the full field reference, and `codesphere-create-reactive-deployment` / `-container-deployment` / `codesphere-add-managed-service` generate a `ci.yml` end-to-end. Reach for this one specifically when hand-editing YAML.

## The core problem

`POST /workspaces/{id}/landscape/deploy` (what `cs sync landscape` calls) returns, for **any** schema violation:

```json
{"status":500,"title":"An internal error occurred"}
```

No field name, no expected-vs-got diff — indistinguishable from a real backend bug. The only place the actual "Config Schema Violation" error surfaces (naming the field and its expected vs. actual type) is the web UI's **Landscape Config Editor**. See Troubleshooting below.

Several official per-feature snippets (e.g. the "Managed Containers" section's `alpine-image`/`nginx-server` examples) omit fields — `prepare`, `test`, `env`, `runAsUser`, `runAsGroup`, `volumeMounts` — that turn out to be required in practice, while the docs' own comprehensive schema-version fixtures already include all of them. Copying a short feature snippet instead of the full fixture is the single most common way to end up staring at the bare 500 above.

## The canonical safe template (v0.4, annotated, verbatim from the docs)

Reproduced **verbatim, comments included** — not a paraphrase. Start every new `ci.yml`, and every addition to an existing one, from this shape: `prepare`/`test`, a Reactive (`reactive-app`), a Managed Container (`container-app`), and an embedded Managed Service (`postgres`), every field the real validator expects present.

```yaml
# Example fixture for the current schema after the managed-service provider schemaVersion rename.
# `schemaVersion` pins the fixture to the current pipeline schema.
schemaVersion: v0.4

# `prepare` runs setup before tests and deploys.
prepare:
  # `steps` lists sequential setup commands.
  steps:
    # `name` is a descriptive label only.
    - name: Install dependencies
      # `command` is the shell command to execute.
      command: yarn install

# `test` runs verification commands after setup.
test:
  # `steps` keeps the same stage shape as `prepare`.
  steps:
    # `name` documents the test step.
    - name: Run unit tests
      # `command` runs the example test suite.
      command: yarn test --run

# `run` contains the named deploy targets for the workspace.
run:
  # Reactive runtime with the full set of supported server fields.
  reactive-app:
    # `steps` defines the lifecycle commands for a reactive runtime.
    steps:
      # Build the app before startup.
      - name: Build app
        # Produce deployable build output.
        command: yarn build
      # Start the application process.
      - name: Start app
        # Launch the runtime server.
        command: yarn start
    # `image` optionally customizes the reactive base image.
    image: ghcr.io/codesphere-cloud/example-reactive:latest
    # `healthEndpoint` is probed to determine readiness.
    healthEndpoint: http://127.0.0.1:3000/healthz/app
    # `plan` selects the workspace plan.
    plan: 20
    # `replicas` requests horizontal scaling.
    replicas: 2
    # `isPublic` is ignored when advanced `network` settings are present.
    isPublic: true
    # `network` defines exposed ports and routed paths.
    network:
      # `ports` exposes raw runtime ports.
      ports:
        # `port` is the runtime port number to expose.
        - port: 3000
          # `isPublic` makes the port reachable from outside the workspace.
          isPublic: true
        - port: 9229
          isPublic: false
      # `paths` maps HTTP prefixes onto runtime ports.
      paths:
        # Route the root path to port `3000`.
        - port: 3000
          path: /
          stripPath: false
        # Route `/api` to port `3000` and strip the prefix.
        - port: 3000
          path: /api
          stripPath: true
    # `env` injects runtime environment variables.
    env:
      # Example string env var.
      NODE_ENV: production
      # Example numeric env var.
      PORT: 3000
    # `runAsUser` sets the runtime UID.
    runAsUser: 1000
    # `runAsGroup` sets the runtime GID.
    runAsGroup: 1000
    # `volumeMounts` attaches workspace storage into the runtime.
    volumeMounts:
      # `name` selects the workspace volume.
      - name: _workspace
        # `mountPath` is the destination inside the runtime.
        mountPath: /home/user/app
        # `workspacePath` selects which subdirectory to mount.
        workspacePath: ""

  # Container runtime with the full set of supported server fields.
  container-app:
    # `image` is required for container runtimes.
    image: ghcr.io/codesphere-cloud/example-container:latest
    # `command` overrides the image entrypoint/command.
    command:
      # First argv entry.
      - node
      # Second argv entry.
      - server.js
    # The rest of the fields are shared with reactive runtimes.
    healthEndpoint: http://127.0.0.1:8080/healthz/app
    # `plan` selects the workspace plan.
    plan: 20
    # `replicas` requests horizontal scaling for this runtime.
    replicas: 1
    # `isPublic` allows the routed service to be reachable externally.
    isPublic: true
    # This simple network shape publishes one routed path.
    network:
      # `path` is the public prefix for this service.
      path: /container
      # `stripPath` controls whether the prefix is preserved upstream.
      stripPath: false
    env:
      # Example string env var.
      NODE_ENV: production
      # Example numeric env var.
      PORT: 8080
    # `runAsUser` sets the runtime UID.
    runAsUser: 1000
    # `runAsGroup` sets the runtime GID.
    runAsGroup: 1000
    # `volumeMounts` attaches workspace storage into the runtime.
    volumeMounts:
      # `name` selects the workspace volume.
      - name: _workspace
        # `mountPath` is the destination inside the runtime.
        mountPath: /home/user/app
        # `workspacePath` selects which subdirectory to mount.
        workspacePath: ""

  # Managed service runtime in v0.4:
  # - `provider.version` was renamed to `provider.schemaVersion`
  # This key names the managed service instance.
  postgres:
    # `provider` identifies the catalog entry to provision.
    provider:
      # `name` selects the managed service type.
      name: postgres
      # `schemaVersion` is the renamed provider version field in v0.4.
      schemaVersion: "14"
    # `plan` selects the service plan and sizing.
    plan:
      # `id` identifies the chosen catalog plan.
      id: 1
      # `parameters` contains provider-specific numeric options.
      parameters:
        # `storage` requests disk capacity.
        storage: 1024
        # `cpu` requests compute capacity.
        cpu: 1
        # `memory` requests RAM capacity.
        memory: 512
    # `config` contains non-secret provider configuration.
    config:
      # `database` is an example provider config key.
      database: app
      # `extensions` shows that lists are allowed in config payloads.
      extensions:
        - pgcrypto
    # `secrets` contains secret provider configuration.
    secrets:
      # `password` is an example secret key.
      password: secret
```

## Keep in mind

- **Adding just one Managed Container** (the most common hand-edit) — don't trim it down to only the fields that feel logically necessary. Keep `env`, `runAsUser`, `runAsGroup`, `volumeMounts` present even when they look trivial for that image, and keep `prepare: { steps: [] }` / `test: { steps: [] }` present at the top level even with nothing to put in them — an empty list is not the same as an absent key to the validator. This is exactly the gap between the docs' short "Managed Containers" snippets and their full versioned fixtures; when in doubt, copy the shape from the fixture above, not a shorter snippet.
- **`runAsUser`/`runAsGroup` only cover the `volumeMounts` target, not the image's own `WORKDIR`.** `runAsUser: 1501` / `runAsGroup: 1010` gets Reactive-equivalent read/write access to the mounted path (e.g. `/home/user/app`) — it does not retroactively make a Docker image's build-time `WORKDIR` writable, which is still owned by whatever user built the image (typically `root`). Concretely: `hiyouga/llamafactory`'s LlamaBoard hardcodes a cache directory relative to the process's cwd, with no env override — if the launch command starts from the image's `WORKDIR` instead of the writable mount, it fails to write that cache even with `runAsUser`/`runAsGroup`/`volumeMounts` all set correctly. Fix by `cd`-ing into the mount before starting the process:
  ```yaml
  run:
    llamaboard:
      image: hiyouga/llamafactory:latest
      steps:
        - name: Start LlamaBoard
          command: cd /home/user/app && llamafactory-cli webui
      runAsUser: 1501
      runAsGroup: 1010
      volumeMounts:
        - name: _workspace
          mountPath: /home/user/app
          workspacePath: ""
      # ...plus env, network, plan, etc. per the field set above.
  ```
  This is why `steps:` (a shell string, supports `&&`/`cd`) is the right field when the working directory needs adjusting first — `command:` execs the image's CMD directly with no shell, so there's no way to `cd` inside it.
- **`landscape/deploy` returns a bare 500:** don't iterate blindly against the API — the 500 body gives no signal about which field, if any, is wrong. Open the Landscape Config Editor in the Codesphere web UI and paste the same `ci.yml`; it surfaces the real "Config Schema Violation" (field + expected-vs-got type) that the API doesn't. If there's no working baseline yet, scaffold the service in the UI first and diff your hand-authored version against the UI-generated one field-by-field. Once the UI names the offending field, cross-check it against `codesphere`'s `references/ci-pipeline.md` and the template above before retrying. This is a backend gap (generic 500 vs. the UI's detailed error), not something the CLI can work around client-side.

## Related

- `codesphere` — full field reference (`references/ci-pipeline.md`), runtimes (`references/runtimes.md`), and provider catalog this skill's template is drawn from
- `codesphere-create-reactive-deployment` / `codesphere-create-container-deployment` / `codesphere-add-managed-service` — generate a new `ci.yml` (or a new service within one) from source/Dockerfiles/a service selection, rather than hand-editing an existing one
