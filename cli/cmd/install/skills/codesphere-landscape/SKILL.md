---
name: codesphere-landscape
description: Use this skill when authoring or editing a Codesphere Landscape's ci.yml by hand — adding, changing, or debugging a Reactive, Managed Container, or embedded Managed Service entry under run:. Provides the full annotated v0.4 schema fixture as the canonical safe starting template (not the shorter per-feature doc snippets, which omit fields that are required in practice), a specific trap around adding just one Managed Container, a note on why runAsUser/runAsGroup alone don't make an image's own WORKDIR writable, and what to do when `cs sync landscape` / POST /workspaces/{id}/landscape/deploy returns a bare 500 instead of a real validation error. Trigger on "ci.yml", "landscape deploy", "landscape/deploy 500", "Config Schema Violation", "cs sync landscape", adding a Managed Container/Managed Service to an existing ci.yml, or a schema-shaped 500 with no useful error body.
license: none
allowed-tools: Read Write Glob Grep
metadata:
  version: "1.0.0"
  updated: "2026-09-09"
  cost-tier: low
---

> **Process:** Read this file in full before writing or editing a `ci.yml`. The canonical template below is the safe starting point — start from it and add/remove services, don't start from a blank file or from a documentation snippet.

## When to use this

Trigger when the task is to hand-author or hand-edit a Landscape's `ci.yml` — adding a new service (Reactive, Managed Container, or embedded Managed Service) to an existing Landscape, writing one from scratch, or debugging why `cs sync landscape` (or a direct `POST /workspaces/{id}/landscape/deploy` call) is failing. This skill is specifically about the **hand-authoring trap**: the public docs' shorter, per-feature snippets look complete and parse as valid YAML, but omit fields the real backend validator requires — and the API's error response when that happens gives you nothing to go on.

This complements, rather than replaces, the broader `codesphere` skill family: `codesphere` holds the full field reference and per-provider details, `codesphere-create-reactive-deployment` / `codesphere-create-container-deployment` / `codesphere-add-managed-service` generate a `ci.yml` from source/Dockerfiles/a service selection. Reach for those when generating a new deployment end-to-end; reach for this skill specifically when hand-editing YAML and you want the safe template plus the two confirmed failure modes below.

## The core problem

`POST /workspaces/{id}/landscape/deploy` (what `cs sync landscape` calls) returns, for **any** `ci.yml` schema violation:

```json
{"status":500,"title":"An internal error occurred"}
```

This is indistinguishable from a real backend bug — no field name, no expected-vs-got diff, nothing. The only place the actual validation error surfaces — a proper "Config Schema Violation" naming the offending field and its expected vs. actual type — is the web UI's **Landscape Config Editor**. See Troubleshooting below.

Compounding this: several of the official docs' shorter per-feature snippets (e.g. the "Managed Containers" section's `alpine-image`/`nginx-server` examples) omit fields — `prepare`, `test`, `env`, `runAsUser`, `runAsGroup`, `volumeMounts` — that turn out to be required in practice, while the docs' own comprehensive schema-version fixtures (the "Versioned ci.yml Schemas" section, v0.1–v0.4) already include all of them. Copying a short feature snippet instead of starting from the full fixture is the single most common way to end up staring at the bare 500 above.

## The canonical safe template (v0.4, annotated, verbatim from the docs)

This is the official docs' full v0.4 fixture, reproduced **verbatim, comments included** — not a paraphrase. Start every new `ci.yml`, and every addition to an existing one, from this shape. It demonstrates `prepare`/`test`, a Reactive (`reactive-app`), a Managed Container (`container-app`), and an embedded Managed Service (`postgres`) with every field the real validator expects present.

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

## Hard rule: adding just one Managed Container

When the task is "add one Managed Container to an existing `ci.yml`" (the single most common hand-edit — e.g. dropping in an off-the-shelf image alongside an existing Reactive), **do not trim the new service down to only the fields that feel logically necessary for that image.** Keep the full per-service field set present — `env`, `runAsUser`, `runAsGroup`, `volumeMounts` — even when they look trivial for this particular image (an empty-ish `env: {}`, a `volumeMounts` entry that doesn't seem to matter for a stateless container). Also keep `prepare: { steps: [] }` and `test: { steps: [] }` present at the top level even if this Landscape has no setup or test commands at all — an empty list is not the same as an absent key to the validator.

This is the exact gap between the docs' short "Managed Containers" feature snippets and their full "Versioned ci.yml Schemas" fixtures: the short snippets (`alpine-image`, `nginx-server`) read as self-contained, minimal, correct examples, and are missing exactly the fields above. The full v0.4 fixture above already includes all of them on both `container-app` and `reactive-app`. When in doubt, copy the shape from the fixture above, not from a shorter snippet elsewhere in the docs — including this skill's own future edits, if a shorter example ever gets added here.

## `runAsUser`/`runAsGroup` only cover the volume mount, not the image's own `WORKDIR`

`runAsUser: 1501` / `runAsGroup: 1010` is the combination documented as getting the same read/write access to workspace files as a Reactive. But that access is scoped to the **`volumeMounts` target** (e.g. `/home/user/app`) — it does **not** retroactively make the image's own build-time `WORKDIR` writable. A Docker image's `WORKDIR` (and anything under it not covered by a volume mount) is still owned by whatever user built the image — typically `root` — and stays read-only (or at least not writable by 1501:1010) at runtime.

This matters concretely for any image that writes relative-path caches or config with no environment-variable override. A real example: `hiyouga/llamafactory`'s LlamaBoard hardcodes its cache directory (`llamaboard_cache`) relative to the process's current working directory, with no env var to redirect it. If the container's launch command starts the process from the image's `WORKDIR` (root-owned at build time) instead of the writable mount, it fails at runtime trying to create that cache directory — even with `runAsUser`/`runAsGroup` set correctly and a `volumeMounts` entry present, because the process never actually `cd`'d into the mount before writing.

The fix is to make the launch command `cd` into the writable mount **before** starting the process:

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

This is why the field is `steps:` (a shell command string — supports `&&`, `cd`, multiple statements) and not `command:` (a bare argv array, no shell) whenever the launch needs to change directory or otherwise chain shell logic first. `command:` overrides the image's CMD directly via exec, with no shell in between — there is no way to `cd` inside a bare argv array. If the working directory needs adjusting before the actual process starts, the service needs `steps:`, which for a Managed Container means providing a shell-based startup step rather than relying on `command:` alone. (`steps:` and `image:` together — a Managed Container running its own startup shell command against a custom base image — is exactly the Reactive-with-custom-`image:` shape documented in the Reactive field reference; see `codesphere`'s `references/ci-pipeline.md` for the full field table.)

## Troubleshooting: `landscape/deploy` returns a bare 500

If `cs sync landscape` (or a direct call to `POST /workspaces/{id}/landscape/deploy`) fails with `{"status":500,"title":"An internal error occurred"}`:

1. **Don't iterate blindly against the API.** Changing a field, re-running `cs sync landscape`, getting the same opaque 500, changing another field, and repeating is slow and often converges on the wrong fix, because the 500 body gives no signal about which field (if any) is actually wrong versus a real backend issue.
2. **Open the Landscape Config Editor in the Codesphere web UI** and paste (or load) the same `ci.yml`. The UI surfaces the real validator error there — a "Config Schema Violation" naming the specific field and an expected-vs-got type mismatch — where the API gives you nothing.
3. **If you don't have a working baseline yet, scaffold the service in the UI first**, then treat the UI-generated `ci.yml` as ground truth. Diff your hand-authored version against it field-by-field rather than guessing further from the docs alone.
4. Once the UI has told you the actual offending field, cross-check it against the full field reference in `codesphere`'s `references/ci-pipeline.md` and the canonical template above before re-attempting `cs sync landscape`.

This gap — the API returning a generic 500 instead of the UI's detailed schema-violation error — is a backend gap, not something the CLI can work around client-side. If you hit this repeatedly, it's worth flagging to the Codesphere API team directly rather than treating it as a CLI limitation.

## Related

- `codesphere` — full field reference (`references/ci-pipeline.md`), runtimes (`references/runtimes.md`), and provider catalog this skill's template is drawn from
- `codesphere-create-reactive-deployment` / `codesphere-create-container-deployment` / `codesphere-add-managed-service` — generate a new `ci.yml` (or a new service within one) from source/Dockerfiles/a service selection, rather than hand-editing an existing one
