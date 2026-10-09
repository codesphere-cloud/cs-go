## cs delete resource-group

Delete resource group

### Synopsis

Delete a resource group from Codesphere or an Organization

```
cs delete resource-group [flags]
```

### Examples

```
# Delete a resource group
$ cs delete resource-group -t <resourceGroupId>
```

### Options

```
  -h, --help   help for resource-group
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

* [cs delete](cs_delete.md)	 - Delete Codesphere resources

