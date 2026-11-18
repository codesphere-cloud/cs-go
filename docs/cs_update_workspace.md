## cs update workspace

Change workspace access, name or plan

### Synopsis

Change settings of a workspace: who can reach it, its name or its plan.

Access: a restricted workspace (the default) sends visitors of its dev domain
and of every landscape path on it to the Codesphere sign-in. --public (or
--restricted=false) makes it reachable by anyone without signing in; --private
(or --restricted=true) restricts it again.

In AI agent sessions access changes need the user's approval: agents use the
MCP tool update_workspace, which asks the user first.

The command prints the workspace as it is after the change.

```
cs update workspace [flags]
```

### Examples

```
# Make the current workspace (CS_WORKSPACE_ID) reachable without sign-in
$ cs update workspace --public

# Same as --public, for workspace 1234
$ cs update workspace -w 1234 --restricted=false

# Require sign-in again
$ cs update workspace --private

# Rename the workspace and change its plan
$ cs update workspace --name docs --plan 21
```

### Options

```
  -h, --help          help for workspace
      --name string   New workspace name
      --plan int      New workspace plan ID (see 'cs list plans')
      --private       Require sign-in to reach the workspace (same as --restricted=true)
      --public        Make the workspace reachable without sign-in (same as --restricted=false)
      --restricted    Whether visitors must sign in to Codesphere (--restricted=false makes the workspace public) (default true)
```

### Options inherited from parent commands

```
  -a, --api string           URL of Codesphere API (can also be CS_API)
  -g, --org string           Organization ID (relevant for some commands)
  -G, --resource-group int   Resource group ID (relevant for some commands, can also be CS_RESOURCE_GROUP_ID; legacy --team, -t and CS_TEAM_ID still work) (default -1)
  -v, --verbose              Verbose output
  -w, --workspace int        Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs update](cs_update.md)	 - Update Codesphere CLI

