## cs list organization-members

List organization members

### Synopsis

List all members of an organization

```
cs list organization-members [flags]
```

### Examples

```
# List all members of an organization
$ cs list organization-members --org <orgId>

# List all members of an organization in JSON format
$ cs list organization-members --org <orgId> --output json
```

### Options

```
  -h, --help   help for organization-members
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

