## cs wake-up

Wake up an on-demand workspace

### Synopsis

Wake up an on-demand workspace by scaling it to 1 replica via the API.

```
cs wake-up [flags]
```

### Examples

```
# wake up workspace 1234
$ cs wake-up -w 1234

# wake up workspace set by environment variable CS_WORKSPACE_ID
$ cs wake-up 

# wake up workspace with 60 second timeout
$ cs wake-up -w 1234 --timeout 60s
```

### Options

```
  -h, --help               help for wake-up
      --timeout duration   Timeout for waking up the workspace (default 2m0s)
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

