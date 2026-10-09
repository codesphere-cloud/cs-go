## cs list resource-groups

List resource groups

### Synopsis

List resource groups available in Codesphere

```
cs list resource-groups [flags]
```

### Examples

```
# List all resource groups
$ cs list resource-groups 

# List resource groups in an organization
$ cs list resource-groups --org <orgId>
```

### Options

```
  -h, --help   help for resource-groups
```

### Options inherited from parent commands

```
  -a, --api string           URL of Codesphere API (can also be CS_API)
  -g, --org string           Organization ID (relevant for some commands)
  -o, --output string        Output format (table, json, yaml) (default "table")
  -t, --resource-group int   Resource group ID (relevant for some commands, can also be CS_RESOURCE_GROUP_ID; --team and CS_TEAM_ID still work) (default -1)
  -v, --verbose              Verbose output
  -w, --workspace int        Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs list](cs_list.md)	 - List resources

