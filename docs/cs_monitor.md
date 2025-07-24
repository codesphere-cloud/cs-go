## cs monitor

Monitor a command and report health information

### Synopsis

Loops over running a command and report information in a health endpoint.

Codesphere watches for health information of an application on port 3000, which is the default endpoint for this command.
You can specify a different port if your application is running on port 3000.

The monitor command keeps restarting an application and reports the metrics about the restarts in prometheus metrics format.
Metrics reported are
* cs_monitor_total_restarts_total - Total number of command executions completed

With the Forwarding option, instead of providing a healthcheck endpoint, requests are forwarded to the specified application endpoint. 
This is useful if the application does provide own healthcheck output but not on the default localhost:3000/

```
cs monitor [flags]
```

### Examples

```
# monitor application that ist started by npm
$ cs monitor -- npm start

# monitor application from local binary on port 3000, expose metrics on port 8080
$ cs monitor --address :8080 -- ./my-app -p 3000 

# forward health-check to application health endpoint
$ cs monitor --forward http://localhost:8080/my-healthcheck -- ./my-app --healthcheck :8080

# forward health-check to application health endpoint, ignore invalid TLS certs
$ cs monitor --forward --insecure-skip-verify -- ./my-app --healthcheck https://localhost:8443

# forward health-check to application health endpoint, using custom CA cert, e.g. for self-signed certs
$ cs monitor --forward --ca-cert-file ca.crt -- ./my-app --healthcheck https://localhost:8443
```

### Options

```
      --address string         Custom listen address for the endpoint (metrics endpoint or forwarding port when --forward option is used) (default ":3000")
      --ca-cert-file string    TLS CA certificate (only relevant for --forward option when healthcheck is exposed as HTTPS enpoint with custom certificate)
      --forward string         Forward healthcheck requests to application health endpoint
  -h, --help                   help for monitor
      --insecure-skip-verify   Skip TLS validation (only relevant for --forward option when healthcheck is exposed as HTTPS endpoint with custom certificate)
      --max-restarts int       Maximum number of restarts before exiting (default -1)
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

