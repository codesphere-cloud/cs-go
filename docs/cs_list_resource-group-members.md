## cs list resource-group-members

List resource group members

### Synopsis

List all members of a resource group

```
cs list resource-group-members [flags]
```

### Examples

```
# List all members of a resource group
$ cs list resource-group-members -t <resourceGroupId>

# List all members of a resource group in JSON format
$ cs list resource-group-members -t <resourceGroupId> -o json
```

### Options

```
  -h, --help   help for resource-group-members
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

