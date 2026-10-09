## cs add organization-member

Add organization member

### Synopsis

Add organization member. Select the organization using --org or CS_ORG_ID.

```
cs add organization-member [flags]
```

### Examples

```
# Add organization member
$ cs add organization-member --org <orgId> -e user@example.com -r member
```

### Options

```
  -e, --email string   Organization member email
  -h, --help           help for organization-member
  -r, --role string    Organization role (admin, member) (default "member")
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

* [cs add](cs_add.md)	 - Add Codesphere resources

