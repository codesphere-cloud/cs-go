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

See our [Usage Documentation](docs) for usage information about the specific subcommands.

## Go SDK

## Community & Contributions

Please review our [Code of Conduct](CODE_OF_CONDUCT.md) to understand our community expectations.
We welcome contributions! All contributions to this project must be made in accordance with the Developer Certificate of Origin (DCO). See our full [Contributing Guidelines](CONTRIBUTING.md) for details.
