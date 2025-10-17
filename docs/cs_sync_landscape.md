## cs sync landscape

Sync landscape

### Synopsis

Sync landscape according to CI profile, i.e. allocate resources for defined services.

```
cs sync landscape [flags]
```

### Options

```
  -h, --help             help for landscape
  -p, --profile string   CI profile to use (e.g. 'prod' for the profile defined in 'ci.prod.yml'), defaults to the ci.yml profile
```

### Options inherited from parent commands

```
  -a, --api string      URL of Codesphere API (can also be CS_API)
  -t, --team int        Team ID (relevant for some commands, can also be CS_TEAM_ID) (default -1)
  -v, --verbose         Verbose output
  -w, --workspace int   Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs sync](cs_sync.md)	 - Sync Codesphere resources

