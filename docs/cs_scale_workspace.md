## cs scale workspace

Scale landscape services of a workspace

### Synopsis

Scale landscape services of a workspace by specifying service name and replica count.

```
cs scale workspace [flags]
```

### Examples

```
# scale frontend to 2 and backend to 3 replicas
$ cs scale workspace --service frontend=2 --service backend=3

# scale web service to 1 replica on workspace 1234
$ cs scale workspace -w 1234 --service web=1

# scale api service to 0 replicas
$ cs scale workspace --service api=0
```

### Options

```
  -h, --help                  help for workspace
      --service stringArray   Service to scale (format: 'service=replicas'), can be specified multiple times
```

### Options inherited from parent commands

```
  -a, --api string      URL of Codesphere API (can also be CS_API)
  -t, --team int        Team ID (relevant for some commands, can also be CS_TEAM_ID) (default -1)
  -v, --verbose         Verbose output
  -w, --workspace int   Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs scale](cs_scale.md)	 - Scale Codesphere resources

