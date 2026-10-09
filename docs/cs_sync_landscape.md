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
  -a, --api string           URL of Codesphere API (can also be CS_API)
  -g, --org string           Organization ID (relevant for some commands)
  -t, --resource-group int   Resource group ID (relevant for some commands, can also be CS_RESOURCE_GROUP_ID; --team and CS_TEAM_ID still work) (default -1)
  -v, --verbose              Verbose output
  -w, --workspace int        Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs sync](cs_sync.md)	 - Sync Codesphere resources

