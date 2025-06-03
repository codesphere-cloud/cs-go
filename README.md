[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![GitHub Workflow Status](https://github.com/codesphere-cloud/cs-go/actions/workflows/build.yml/badge.svg)

# Codesphere Go SDK & CLI

Seamlessly integrate CS into your local development flow.

## License

This project is licensed under the Apache License, Version 2.0 - see the [LICENSE](LICENSE) file for details.

## CLI

### Installation

You can install the Codesphere CLI in a few ways:

#### Using GitHub CLI (`gh`)

If you have the [GitHub CLI](https://cli.github.com/) installed, you can install the `cs` CLI with a command like the following.
Not that some commands may require you to elevate to the root user with `sudo`.

##### ARM Mac

```
gh release download -R codesphere-cloud/cs-go -o cs -D /usr/local/bin/cs -p *darwin_arm64
chmod +x /usr/local/bin/cs
```

##### Linux Amd64

```
gh release download -R codesphere-cloud/cs-go -o cs -D /usr/local/bin/cs -p *linux_amd4
chmod +x /usr/local/bin/cs
```

#### Using `wget`

This option requires to have the `wget` and `jq` utils installed. Download the `cs` CLI and add permissions to run it with the following commands:
Not that some commands may require you to elevate to the root user with `sudo`.

##### ARM Mac

```
wget -qO- 'https://api.github.com/repos/codesphere-cloud/cs-go/releases/latest' | jq -r '.assets[] | select(.name | match("darwin_arm64")) | .browser_download_url' | xargs wget -O cs
mv cs /usr/local/bin/cs
chmod +x /usr/local/bin/cs
```

##### Linux Amd64

```
wget -qO- 'https://api.github.com/repos/codesphere-cloud/cs-go/releases/latest' | jq -r '.assets[] | select(.name | match("linux_amd64")) | .browser_download_url' | xargs wget -O cs
mv cs /usr/local/bin/cs
chmod +x /usr/local/bin/cs
```

#### Manual Download

You can also download the pre-compiled binaries from the [Codesphere CLI Releases page](https://github.com/codesphere-cloud/cs-go/releases).
Not that some commands may require you to elevate to the root user with `sudo`.

1. Go to the [latest release](https://github.com/codesphere-cloud/cs-go/releases/latest).

2. Download the appropriate release for your operating system and architecture (e.g., `cs-go_darwin_amd64` for macOS, `cs-go_linux_amd64` for Linux, or `cs-go_windows_amd64` for Windows).

3. Move the `cs` binary to a directory in your system's `PATH` (e.g., `/usr/local/bin` on Linux/Mac, or a directory added to `Path` environment variable on Windows).

4. Make the binary executable (e.g. by running `chmod +x /usr/local/bin/cs` on Mac or Linux)

### Usage Guide

The Codesphere CLI (`cs`) allows you to manage and debug resources deployed in Codesphere directly from your command line.

#### Global Options & Environment Variables

The `cs` CLI supports several global options that you can set via command-line flags or environment variables. Using environment variables is handy for setting persistent configurations.

| Required      | Command Line Flag                   | Environment Variable | Description |
| ----- | ----- | ----- | ----- |
| Yes           |                        | `CS_TOKEN`           | Codesphere API token. Generate one in your user settings at https://codesphere.com/. |
|               | `--api`<br/>`-a`       | `CS_API`             | URL of the Codesphere API. Default: `https://codesphere.com/api` |
| Some commands | `--team`<br/>`-t`      | `CS_TEAM_ID`         | Your Codesphere Team ID. This is relevant for commands operating on a specific team. |
| Some commands | `--workspace`<br/>`-w` | `CS_WORKSPACE_ID`    | Your Codesphere Workspace ID. Relevant for commands targeting a specific workspace. |

**Note on Team ID and Workspace ID:** If you don't provide these via a flag, the CLI will try to get them from the corresponding environment variables (`CS_TEAM_ID`, `CS_WORKSPACE_ID`). If they're still not found and a command requires them, the CLI will return an error.

#### Available Commands

The `cs` CLI organizes its functionality into several top-level commands, each with specific subcommands and flags.

##### `cs list`

Use this command to list various resources available in Codesphere.

**Usage:**

```
cs list [command]
```

###### `cs list teams`

Lists all teams you have access to in Codesphere.

**Usage:**

```
cs list teams
```

**Example:**

```
$ cs list teams
```

###### `cs list workspaces`

Lists all workspaces available in Codesphere.

**Usage:**

```
cs list workspaces [--team-id <team-id>]
```

**Example:**

```
$ cs list workspaces --team-id <team-id>
```

If you don't specify `--team-id`, the command will try to list workspaces for all teams you can access (or for the team specified by `CS_TEAM_ID`).

##### `cs log`

Retrieves run logs from services within your workspaces.

**Usage:**

```
cs log --workspace-id <workspace-id> [options]
```

**Description:**

You can retrieve logs based on the given scope. If you provide both the step number and server, it returns all logs from all replicas of that server. If you provide a specific replica ID, it will return logs for that replica only.

**Examples:**

```
# Get logs from a specific server within a workspace
$ cs log -w 637128 -s app

# Get all logs from all servers in a workspace
$ cs log -w 637128

# Get logs from a specific replica
$ cs log -w 637128 -r workspace-213d7a8c-48b4-42e2-8f70-c905ab04abb5-58d657cdc5-m8rrp

# Get logs from a self-hosted Codesphere installation (using a custom API URL)
$ cs log --api https://codesphere.acme.com/api -w 637128 -s app
```

**Flags:**

* `--server`, `-s` (string): Name of the landscape server.

* `--workspace-id`, `-w` (int): ID of your Codesphere workspace. You can also set this via the `CS_WORKSPACE_ID` environment variable. **This flag or environment variable is required if not set globally.**

* `--step`, `-n` (int): Index of the execution step (default 0).

* `--replica`, `-r` (string): ID of the server replica. If you provide this, the `--server` flag will be ignored.

##### `cs set-env`

Sets environment variables for your workspace.

**Usage:**

```
cs set-env --workspace-id <workspace-id> --env <key>=<value> [--env <key2>=<key2> ...]
```

**Example:**

```
# Set environment variables for a specific workspace
$ cs set-env --workspace-id <workspace-id> --env foo=bar --env hello=world
```

**Flags:**

* `--env-var`, `-e` (stringArray): Environment variables to set, in the format `key=val`. You can use this flag multiple times to set several variables.

## Go SDK

## Community & Contributions

Please review our [Code of Conduct](CODE_OF_CONDUCT.md) to understand our community expectations.
We welcome contributions! All contributions to this project must be made in accordance with the Developer Certificate of Origin (DCO). See our full [Contributing Guidelines](CONTRIBUTING.md) for details.
