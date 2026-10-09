## cs create env

Set environment variables

### Synopsis

Set environment variables in a workspace

```
cs create env [flags]
```

### Examples

```
# Set single environment variable
$ cs create env --workspace <workspace-id> --env-var foo=bar

# Set multiple environment variables
$ cs create env --workspace <workspace-id> --env-var foo=bar --env-var hello=world
```

### Options

```
  -e, --env-var stringArray   env vars to set in form key=val
  -h, --help                  help for env
```

### Options inherited from parent commands

```
  -a, --api string           URL of Codesphere API (can also be CS_API)
  -g, --org string           Organization ID (relevant for some commands)
  -t, --resource-group int   Resource group ID (relevant for some commands, can also be CS_RESOURCE_GROUP_ID; --team and CS_TEAM_ID still work) (default -1)
  -v, --verbose              Verbose output
  -w, --workspace int        Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs create](cs_create.md)	 - Create codesphere resource

