## cs stop pipeline

Stop pipeline stages of a workspace

### Synopsis

Stop one or many pipeline stages of a workspace.

Stages can be 'prepare', 'test', or 'run'.
When multiple stages are specified, the command will stop them in the provided order.
The command sends a stop request for each stage and returns after all requests succeed.

```
cs stop pipeline [flags]
```

### Examples

```
# Stop the run stage
$ cs stop pipeline run

# Stop the prepare and test stages in order
$ cs stop pipeline prepare test

# Stop the prepare, test, and run stages in order
$ cs stop pipeline prepare test run
```

### Options

```
  -h, --help   help for pipeline
```

### Options inherited from parent commands

```
  -a, --api string      URL of Codesphere API (can also be CS_API)
  -t, --team int        Team ID (relevant for some commands, can also be CS_TEAM_ID) (default -1)
  -v, --verbose         Verbose output
  -w, --workspace int   Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs stop](cs_stop.md)	 - Stop workspace pipeline

