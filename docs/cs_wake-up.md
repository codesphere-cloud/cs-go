## cs wake-up

Wake up an on-demand workspace

### Synopsis

Wake up an on-demand workspace by scaling it to 1 replica via the API. Optionally syncs the landscape to start services.

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

# wake up workspace and deploy landscape from CI profile
$ cs wake-up -w 1234 --sync-landscape

# wake up workspace and deploy landscape with prod profile
$ cs wake-up -w 1234 --sync-landscape --profile prod

# wake up workspace and scale specific services
$ cs wake-up -w 1234 --scale-services web=1,api=2
```

### Options

```
  -h, --help                    help for wake-up
  -p, --profile string          CI profile to use for landscape deploy (e.g. 'prod' for ci.prod.yml)
      --scale-services string   Scale specific landscape services (format: 'service=replicas,service2=replicas')
      --sync-landscape          Deploy landscape from CI profile after waking up
      --timeout duration        Timeout for waking up the workspace (default 2m0s)
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

