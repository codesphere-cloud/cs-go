# Codesphere Runtimes & Language Toolchains Reference

> **Last updated:** 2026-07-28 · **Source:** https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/installing-deps-with-nix

> The documentation changes regularly. If you are facing issues, please refer to the official documentation:
> <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/installing-deps-with-nix>

## Overview

Reactives run on Codesphere's pooled, pre-warmed Ubuntu base image. Some language runtimes come preinstalled at a default version; anything else — a different version, or a language not preinstalled at all (Java, Rust, Ruby, PHP, ...) — is installed **via Nix**, since workspaces have no root/sudo access for `apt`. Use this reference whenever generating `prepare`/`run` steps that need to enable or version-pin a language toolchain in `ci.yml`.

## Core Concepts

- **No root access**: `apt`/`sudo` package managers are not available inside a Reactive. Nix (`nix-env -iA nixpkgs.<package>`) is the documented mechanism for installing OS-level packages/toolchains.
- **Nix packages are workspace-shared, not stage-persistent**: Nix packages persist in `/nix/store`, shared across every service in a Landscape — but a version switched with a non-Nix tool (like Node's `n`) does **not** persist between the `prepare` stage and the `run` stage, or between replicas. `sudo n <version>` (or any per-stage version switch) must be repeated in **both** `prepare` and `run`.
- **`prepare` vs `run`**: `prepare` runs once, on the main replica only (build/install steps). `run` executes on every replica, restarts automatically on crash, and is where the long-running start command belongs.
- **Reactive vs. Managed Container**: everything below assumes a Reactive (`steps:`, no `image:`). If a language/runtime is easier to ship as a prebuilt Docker image instead, use a Managed Container (`image:`) — see `ci-pipeline.md`.

## API / Syntax

### Installing any package via Nix

- **Description:** General-purpose pattern for enabling a language runtime, tool, or system package that isn't preinstalled. Runs as a `prepare` step (or a `run` step, if the tool must also be present when the app restarts).
- **Parameters:**

| Name        | Type   | Required | Description                                                                                                                                |
| ----------- | ------ | -------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| `<package>` | string | Yes      | A [nixpkgs](https://search.nixos.org/packages) attribute name, e.g. `nodejs_24`, `python311`, `jdk21`, `rustc`, `php83`, `go`, `ruby_3_3`. |

- **Example:**

```yaml
prepare:
  steps:
    - name: Install a toolchain
      command: nix-env -iA nixpkgs.<package>
```

### Node.js — preinstalled default, or pin a version

- **Description:** A default Node.js version ships on the base image. To pin a specific version, either switch it with `n` (must repeat in **both** `prepare` and `run` — it does not persist across stages/replicas) or install a specific version via Nix (persists automatically, no repetition needed).
- **Example — `n` (per-stage, repeat in both prepare and run):**

```yaml
schemaVersion: v0.4
prepare:
  steps:
    - name: Set Node.js version
      command: sudo n 20
    - name: Install dependencies
      command: npm install
    - name: Build
      command: npm run build
test:
  steps: []
run:
  app:
    steps:
      - name: Set Node.js version
        command: sudo n 20
      - name: Start server
        command: npm start
    plan: 8
    replicas: 1
    network:
      ports:
        - port: 3000
          isPublic: true
      paths:
        - port: 3000
          path: /
          stripPath: false
```

- **Example — Nix (persists across stages automatically):**

```yaml
schemaVersion: v0.4
prepare:
  steps:
    - name: Install Node.js 24
      command: nix-env -iA nixpkgs.nodejs_24
    - name: Install dependencies
      command: npm install
test:
  steps: []
run:
  app:
    steps:
      - name: Start server
        command: npm start
    plan: 8
    replicas: 1
    network:
      ports:
        - port: 3000
          isPublic: true
      paths:
        - port: 3000
          path: /
          stripPath: false
```

### Python (Pipenv)

- **Example:**

```yaml
schemaVersion: v0.4
prepare:
  steps:
    - name: Install dependencies
      command: pipenv install -r requirements.txt
test:
  steps: []
run:
  app:
    steps:
      - name: Start server
        command: >
          source "$(pipenv --venv)/bin/activate" &&
          uvicorn main:app --reload --host 0.0.0.0 --port 3000
    plan: 8
    replicas: 1
    network:
      ports:
        - port: 3000
          isPublic: true
      paths:
        - port: 3000
          path: /
          stripPath: false
```

### Python (pip, target dir) / pinning a Python version via Nix

- **Description:** Default `pip install` needs a target dir for the interpreter to find packages without a venv. Pin a specific Python via Nix (e.g. `nixpkgs.python311`) the same way as Node. **The bare `nixpkgs.python311` package does not bundle `pip`** — confirmed live: installing only `nixpkgs.python311` leaves both `pip`/`pip3` and `python3.11 -m ensurepip` non-functional ("command not found" / "No module named pip" respectively, even from ensurepip's own bootstrap wheels on this build). Install `nixpkgs.<version>Packages.pip` explicitly alongside the interpreter in the same `nix-env` call, and use that same pinned interpreter (not the bare `python3`, which resolves to the base image's own system Python and was never the target of the `--target=` install) to both install and start the app.
- **Example:**

```yaml
schemaVersion: v0.4
prepare:
  steps:
    - name: Install Python 3.11 and pip
      command: nix-env -iA nixpkgs.python311 nixpkgs.python311Packages.pip
    - name: Install dependencies
      command: pip install -r requirements.txt --target=/home/user/app/pipLib
test:
  steps: []
run:
  app:
    steps:
      - name: Start server
        command: PYTHONPATH=/home/user/app/pipLib python3.11 server.py
    plan: 8
    replicas: 1
    network:
      ports:
        - port: 3000
          isPublic: true
      paths:
        - port: 3000
          path: /
          stripPath: false
```

### Go

- **Example:**

```yaml
schemaVersion: v0.4
prepare:
  steps:
    - name: Build
      command: go build -o app
test:
  steps:
    - name: Run tests
      command: go test ./...
run:
  app:
    steps:
      - name: Start server
        command: ./app
    plan: 8
    replicas: 1
    network:
      ports:
        - port: 3000
          isPublic: true
      paths:
        - port: 3000
          path: /
          stripPath: false
```

If Go isn't preinstalled at the version you need, add `nix-env -iA nixpkgs.go` (or a versioned attribute, e.g. `nixpkgs.go_1_22`) as the first `prepare` step.

### Ruby on Rails

- **Example:**

```yaml
schemaVersion: v0.4
prepare:
  steps:
    - name: Install dependencies
      command: bundle install
    - name: Setup database
      command: bundle exec rails db:setup
test:
  steps:
    - name: Run tests
      command: bundle exec rails test
run:
  app:
    steps:
      - name: Start Rails
        command: bundle exec rails server -b 0.0.0.0 -p 3000
    plan: 8
    replicas: 1
    network:
      ports:
        - port: 3000
          isPublic: true
      paths:
        - port: 3000
          path: /
          stripPath: false
```

If Ruby needs to be installed/pinned first, add `nix-env -iA nixpkgs.ruby_3_3` (or the matching nixpkgs attribute) as the first `prepare` step.

### PHP (Laravel)

- **Example:**

```yaml
schemaVersion: v0.4
prepare:
  steps:
    - name: Install dependencies
      command: composer install
test:
  steps: []
run:
  app:
    steps:
      - name: Start Laravel
        command: php artisan serve --host=0.0.0.0 --port=3000
    plan: 8
    replicas: 1
    network:
      ports:
        - port: 3000
          isPublic: true
      paths:
        - port: 3000
          path: /
          stripPath: false
```

If PHP itself (or Composer) needs to be installed/pinned first, add `nix-env -iA nixpkgs.php83` (or the matching nixpkgs attribute, e.g. `php82`) as the first `prepare` step.

### Vue.js

- **Example:**

```yaml
schemaVersion: v0.4
prepare:
  steps:
    - name: Install dependencies
      command: yarn install
    - name: Build
      command: yarn build
test:
  steps: []
run:
  app:
    steps:
      - name: Start preview server
        command: yarn preview --host --port 3000
    plan: 8
    replicas: 1
    network:
      ports:
        - port: 3000
          isPublic: true
      paths:
        - port: 3000
          path: /
          stripPath: false
```

### Java and Rust (via Nix — general pattern, no dedicated framework example in current docs)

- **Description:** The scraped documentation set does not include a dedicated worked example for Java or Rust the way it does for Node/Python/Go/Ruby/PHP/Vue. Both follow the same documented pattern used everywhere else in this file: install the toolchain via Nix in `prepare` (build tools too, if separate from the language runtime), build in `prepare`, run the resulting artifact in `run`. **Treat the exact `nixpkgs` attribute names and build commands below as illustrative, not verified against the official docs — confirm the current attribute name at <https://search.nixos.org/packages> before using in production.**
- **Example — Java (Maven):**

```yaml
schemaVersion: v0.4
prepare:
  steps:
    - name: Install JDK 21
      command: nix-env -iA nixpkgs.jdk21
    - name: Install Maven
      command: nix-env -iA nixpkgs.maven
    - name: Build
      command: mvn -B package
test:
  steps:
    - name: Run tests
      command: mvn -B test
run:
  app:
    steps:
      - name: Install JDK 21
        command: nix-env -iA nixpkgs.jdk21
      - name: Start server
        command: java -jar target/app.jar --server.port=3000
    plan: 8
    replicas: 1
    network:
      ports:
        - port: 3000
          isPublic: true
      paths:
        - port: 3000
          path: /
          stripPath: false
```

- **Example — Rust (Cargo):**

```yaml
schemaVersion: v0.4
prepare:
  steps:
    - name: Install Rust toolchain
      command: nix-env -iA nixpkgs.rustc nixpkgs.cargo
    - name: Build (release)
      command: cargo build --release
test:
  steps:
    - name: Run tests
      command: cargo test --release
run:
  app:
    steps:
      - name: Start server
        command: ./target/release/app
    plan: 8
    replicas: 1
    network:
      ports:
        - port: 3000
          isPublic: true
      paths:
        - port: 3000
          path: /
          stripPath: false
```

Since Rust/Java compile to a standalone binary/artifact and (unlike Node's `n`) the Nix-installed toolchain persists in the shared `/nix/store`, re-running the JDK install in `run` is only strictly required if the `run` step invokes `java`/`mvn` directly (as in the Java example above); a Rust binary launched directly (`./target/release/app`) needs no runtime toolchain present at all in `run`.

### Other / unlisted languages

- **Description:** Same pattern for anything else (Elixir, .NET, Deno, Bun, Zig, ...): find the nixpkgs attribute, install it in `prepare`, build if needed, start the process in `run` bound to `0.0.0.0` on the port referenced by `network`.
- **Example (generic template):**

```yaml
schemaVersion: v0.4
prepare:
  steps:
    - name: Install toolchain
      command: nix-env -iA nixpkgs.<package>
    - name: Install dependencies / build
      command: <build command>
test:
  steps: []
run:
  app:
    steps:
      - name: Start server
        command: <start command, bound to 0.0.0.0:3000>
    plan: 8
    replicas: 1
    network:
      ports:
        - port: 3000
          isPublic: true
      paths:
        - port: 3000
          path: /
          stripPath: false
```

## Common Pitfalls

- Using `sudo n <version>` (or any non-Nix version switch) in only `prepare` or only `run` — it must appear in **both**, since it doesn't persist across stages/replicas. Nix-installed versions don't have this problem.
- Trying `apt`/`sudo apt-get` — no root access; use Nix.
- Forgetting to bind the server to `0.0.0.0` — binding to `127.0.0.1`/`localhost` makes the service unreachable through the Workspace Router even with a correct `network.paths` entry.
- Omitting a `network.paths` entry (or `isPublic: true`) entirely — the Workspace Router never marks a route-less service healthy/reachable.
- Assuming a package installed in `prepare` is visible mid-`prepare` before its own install step finishes, or assuming Nix installs are instant — large toolchains (JDKs, Rust) add real time to `prepare`.
- **Confirmed live:** installing only `nixpkgs.python311` (or any bare `nixpkgs.python3X` attribute) via `nix-env` does **not** provide `pip` — the Python(pip) recipe above installs `nixpkgs.<version>Packages.pip` alongside it for exactly this reason. `python3.11 -m ensurepip` is not a working fallback on this build either (fails with "No module named pip" from ensurepip's own bootstrap wheels). If a `run`/`prepare` step calls bare `python3` after a Nix-pinned install, double check it isn't silently falling back to the base image's own system Python (which has no `pip` either) instead of the Nix-installed, version-pinned one — use the versioned binary (`python3.11`, not `python3`) explicitly throughout.
- **Not yet independently verified**, but very plausibly the same root cause: the Python (Pipenv) recipe above assumes `pipenv` itself is already on `PATH` — if it isn't preinstalled on a given base image, the same "Nix-installed interpreter doesn't automatically bring its usual companion tooling" issue confirmed for `pip` above likely applies; installing `nixpkgs.pipenv` explicitly first would be the analogous fix if this recipe fails the same way.

## Known Documentation Discrepancies

> This section may be outdated. Please verify against the official documentation:
> <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/installing-deps-with-nix>

- The Java and Rust examples in this file are **extrapolated** from the documented Nix pattern and the structurally-identical Go example — they are not copied from an official per-language guide, because the scraped source set doesn't contain one. Verify exact `nixpkgs` attribute names (`jdk21` vs `jdk17`, `rustc`/`cargo` vs a combined `rust` attribute, etc.) at <https://search.nixos.org/packages> before relying on them.
- The exact set of languages preinstalled by default on the base Reactive image (vs. requiring an explicit Nix install) is not enumerated in the scraped docs — check `GET /metadata/workspace-base-images` or the IDE's base-image picker for the current default.

## Further Reading

- Official docs: <https://docs.codesphere.com/workspace-toolkit/ci-and-deploy/installing-deps-with-nix>
- CI pipeline field reference (Reactive/Managed Container/Managed Service shapes): [ci-pipeline.md](./ci-pipeline.md)
- Nix package search: <https://search.nixos.org/packages>
- Metadata API (`/metadata/workspace-base-images`, `/metadata/workspace-plans`): [cli-and-api.md](./cli-and-api.md)
