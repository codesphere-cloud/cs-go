## cs add resource-group-member

Add resource group member

### Synopsis

Add a member to a resource group.

To add a member to a resource group within an organization or a standalone resource group

```
cs add resource-group-member [flags]
```

### Examples

```
# Add a user to a resource group as a member
$ cs add resource-group-member -t <resourceGroupId> -e user@example.com -r member

# Add a user to a resource group as an admin
$ cs add resource-group-member -t <resourceGroupId> -e admin@example.com -r admin
```

### Options

```
  -e, --email string   Resource group member email
  -h, --help           help for resource-group-member
  -r, --role string    Resource group member role (member, admin) (default "member")
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

* [cs add](cs_add.md)	 - Add Codesphere resources

