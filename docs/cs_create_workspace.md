## cs create workspace

Create a workspace

### Synopsis

Create a workspace in Codesphere.

Specify a (private) git repository or start an empty workspace.
Environment variables can be set to initialize the workspace with a specific environment.
The command will wait for the workspace to become running or a timeout is reached.

To decide which plan suits your needs, run 'cs list plans'


```
cs create workspace [flags]
```

### Examples

```
# Create an empty workspace, using plan 20
$ cs create workspace my-workspace -p 20

# Create a workspace from a git repository
$ cs create workspace my-workspace -r https://github.com/codesphere-cloud/landingpage-temp.git

# Create a workspace and set environment variables
$ cs create workspace my-workspace -r https://github.com/codesphere-cloud/landingpage-temp.git -e DEPLOYMENT=prod -e A=B

# Create a workspace and connect to VPN myVpn
$ cs create workspace my-workspace -r https://github.com/codesphere-cloud/landingpage-temp.git --vpn myVpn

# Create a workspace and wait 30 seconds for it to become running
$ cs create workspace my-workspace -r https://github.com/codesphere-cloud/landingpage-temp.git --timeout 30s

# Create a workspace from branch 'staging'
$ cs create workspace my-workspace -r https://github.com/codesphere-cloud/landingpage-temp.git -b staging

# Create a workspace from a private git repository
$ cs create workspace my-workspace -r https://github.com/my-org/my-private-project.git -P
```

### Options

```
  -b, --branch string       branch to check out
  -e, --env stringArray     Environment variables to set in the workspace in key=value form (e.g. --env DEPLOYMENT=prod)
  -h, --help                help for workspace
  -p, --plan int            Plan ID for the workspace (default 8)
  -P, --private             Use private repository
  -r, --repository string   Git repository to create the workspace from
      --timeout duration    Time to wait for the workspace to start (e.g. 5m for 5 minutes) (default 10m0s)
      --vpn string          Vpn config to use
```

### Options inherited from parent commands

```
  -a, --api string      URL of Codesphere API (can also be CS_API)
  -t, --team int        Team ID (relevant for some commands, can also be CS_TEAM_ID) (default -1)
  -v, --verbose         Verbose output
  -w, --workspace int   Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs create](cs_create.md)	 - Create codesphere resource

