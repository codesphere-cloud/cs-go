## cs delete team

Delete resource group (team)

### Synopsis

Delete a resource group (team) from Codesphere or an Organization

```
cs delete team [flags]
```

### Examples

```
# Delete a team
$ cs delete team -t <teamId>
```

### Options

```
  -h, --help   help for team
```

### Options inherited from parent commands

```
  -a, --api string      URL of Codesphere API (can also be CS_API)
  -g, --org string      Organization ID (relevant for some commands)
  -t, --team int        Resource group (team) ID (relevant for some commands, alias --resource-group, can also be CS_RESOURCE_GROUP_ID or CS_TEAM_ID) (default -1)
  -v, --verbose         Verbose output
  -w, --workspace int   Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs delete](cs_delete.md)	 - Delete Codesphere resources

