## cs up

Deploy your local code to Codesphere

### Synopsis

Deploys your local code to a new or existing Codepshere workspace.

Prerequisite: Your code needs to be located in a git repository where you can create and push WIP branches.
When running cs up, the cs CLI will do the following:

* Push local changes to a branch (customizable with -b), if none is specified, cs go creates a WIP branch (will be stored and reused in .cs-up.yaml)
* Create workspace if it doesn't exist yet (state in .cs-up.yaml)
* Start the deployment in Codesphere, customize which profile to use with -p flag (defaults to 'ci.yml')
* Print the dev domain of the workspace to the console once the deployment is successful

```
cs up [flags]
```

### Options

```
      --base-image string          Base image to use for the workspace, if not set, the default base image will be used
  -b, --branch string              Branch to push to, if not set, a WIP branch will be created and reused for subsequent runs
  -e, --env stringArray            Environment variables to set in the format KEY=VALUE, can be specified multiple times for multiple variables
  -h, --help                       help for up
      --plan int                   Plan ID to use for the workspace, if not set, the first available plan will be used (default -1)
      --private-repo string        Whether the git repository is public or private (requires authentication), defaults to 'public'
  -p, --profile string             CI profile to use (e.g. 'ci.dev.yml' for a dev profile, you may have defined in 'ci.dev.yml'), defaults to the ci.yml profile
      --public-dev-domain string   Whether to create a public or private dev domain for the workspace (only applies to new workspaces), defaults to 'public'
      --remote string              Git remote to use for pushing the code, defaults to 'origin' (default "origin")
      --timeout duration           Timeout for the deployment process, e.g. 10m, 1h, defaults to 1m
  -v, --verbose                    Enable verbose output
      --workspace-name string      Name of the workspace to create, if not set, a random name will be generated
  -y, --yes                        Skip confirmation prompt for pushing changes to the git repository
```

### Options inherited from parent commands

```
  -a, --api string          URL of Codesphere API (can also be CS_API)
      --state-file string   Path to the state file, defaults to .cs-up.yaml (default ".cs-up.yaml")
  -t, --team int            Team ID (relevant for some commands, can also be CS_TEAM_ID) (default -1)
  -w, --workspace int       Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs](cs.md)	 - The Codesphere CLI

