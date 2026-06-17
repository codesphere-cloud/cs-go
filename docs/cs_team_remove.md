## cs team remove

Remove team

### Synopsis

Remove a team from Codesphere or an Organization

```
cs team remove [flags]
```

### Examples

```
# Remove a team that does not belong to an Organization
$ cs team remove -t <teamId>

# Remove a team that does belong to an Organization
$ cs team remove -O <orgId> -t <teamId>
```

### Options

```
  -h, --help   help for remove
```

### Options inherited from parent commands

```
  -a, --api string      URL of Codesphere API (can also be CS_API)
  -O, --org string      Organization ID (relevant for some commands)
  -t, --team int        Team ID (relevant for some commands, can also be CS_TEAM_ID) (default -1)
  -v, --verbose         Verbose output
  -w, --workspace int   Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs team](cs_team.md)	 - Manage Team

