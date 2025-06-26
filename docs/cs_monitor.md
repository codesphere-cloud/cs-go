## cs monitor

Monitor a command and report health information

### Synopsis

Loops over running a command and report information in a health endpoint.

Codesphere watches for health information of an application on port 3000, which is the default endpoint for this command.
You can specify a different port if your application is running on port 3000.

The monitor command keeps restarting an application and reports the metrics about the restarts in prometheus metrics format.
Metrics reported are
* cs_monitor_total_restarts_total - Total number of command executions completed

```
cs monitor [flags]
```

### Examples

```
# monitor application that ist started by npm
$ cs monitor -- npm start

# monitor application from local binary on port 3000, expose metrics on port 8080
$ cs monitor --address 8080 -- ./my-app -p 3000 
```

### Options

```
      --address string     Custom listen address for the metrics endpoint (default ":3000")
  -h, --help               help for monitor
      --max-restarts int   Maximum number of restarts before exiting (default -1)
```

### Options inherited from parent commands

```
  -a, --api string      URL of Codesphere API (can also be CS_API)
  -t, --team int        Team ID (relevant for some commands, can also be CS_TEAM_ID) (default -1)
  -w, --workspace int   Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs](cs.md)	 - The Codesphere CLI

