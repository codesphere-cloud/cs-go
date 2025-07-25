## cs set-env

Set environment varariables

### Synopsis

Set environment variables in a workspace

```
cs set-env [flags]
```

### Examples

```
# Set single environment variable
$ cs set-env --workspace <workspace-id> --env-var foo=bar

# Set multiple environment variables
$ cs set-env --workspace <workspace-id> --env-var foo=bar --env-var hello=world
```

### Options

```
  -e, --env-var stringArray   env vars to set in form key=val
  -h, --help                  help for set-env
```

### Options inherited from parent commands

```
  -a, --api string      URL of Codesphere API (can also be CS_API)
  -t, --team int        Team ID (relevant for some commands, can also be CS_TEAM_ID) (default -1)
  -v, --verbose         Verbose output
  -w, --workspace int   Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs](cs.md)	 - The Codesphere CLI

