## cs delete resource-group-member

Delete resource group member

### Synopsis

Delete a member from a resource group.

To delete a member from a resource group within an organization, the CS_ORG_ID environment variable or the -g/--org flag must be set.

```
cs delete resource-group-member [flags]
```

### Examples

```
# Delete a user from a resource group
$ cs delete resource-group-member -t <resourceGroupId> -u <userId>

# Delete a user from a resource group within an organization
$ cs delete resource-group-member -g <org-id> -t <resourceGroupId> -u <userId>
```

### Options

```
  -h, --help       help for resource-group-member
  -u, --user int   Resource group member user ID (default -1)
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

