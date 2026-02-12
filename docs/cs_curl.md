## cs curl

Send authenticated HTTP requests to workspace dev domain

### Synopsis

Send authenticated HTTP requests to a workspace's development domain using curl-like syntax.

```
cs curl [path] [-- curl-args...] [flags]
```

### Examples

```
# GET request to workspace root
$ cs curl / -w 1234

# GET request to health endpoint
$ cs curl /api/health -w 1234

# POST request with data
$ cs curl /api/data -w 1234 -- -XPOST -d '{"key":"value"}'

# verbose output
$ cs curl /api/endpoint -w 1234 -- -v

# HEAD request using workspace from env var
$ cs curl / -- -I
```

### Options

```
  -h, --help               help for curl
      --insecure           skip TLS certificate verification (for testing only)
      --timeout duration   Timeout for the request (default 30s)
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

