## cs start pipeline

Start pipeline stages of a workspace

### Synopsis

Start one or many pipeline stages of a workspace.

Stages can be 'prepare', 'test', or 'run'.
When multiple stages are specified, the command will start the next stage when the previous stage is finished successfully.
If a stage fails, the command won't attempt running the next stage.
The command will not wait for the run stage to finish, but exit when the stage is running.

When only a single stage is specified, the command will wait until the stage is finished, except for the run stage.
Use 'cs log' to stream logs.

```
cs start pipeline [flags]
```

### Examples

```
# Start the prepare stage and wait for it to finish
$ cs start pipeline prepare

# Start the prepare and test stages sequencially and wait for them to finish
$ cs start pipeline prepare test

# Start the prepare, test, and run stages sequencially. Exits after the run stage is triggered
$ cs start pipeline prepare test run

# Start the run stage and exit when running
$ cs start pipeline run

# Start the run stage of the prod profile
$ cs start pipeline -p prod run

# start the prepare stage, timeout after 5 minutes.
$ cs start pipeline -t 5m prepare
```

### Options

```
  -h, --help               help for pipeline
  -p, --profile string     CI profile to use (e.g. 'prod' for the profile defined in 'ci.prod.yml'), defaults to the ci.yml profile
      --timeout duration   Time to wait per stage before stopping the command execution (e.g. 10m) (default 30m0s)
```

### Options inherited from parent commands

```
  -a, --api string      URL of Codesphere API (can also be CS_API)
  -t, --team int        Team ID (relevant for some commands, can also be CS_TEAM_ID) (default -1)
  -v, --verbose         Verbose output
  -w, --workspace int   Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs start](cs_start.md)	 - Start workspace pipeline

