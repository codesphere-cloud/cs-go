## cs delete organization-member

Remove organization member

### Synopsis

Remove organization member. Select the organization using --org or CS_ORG_ID.

```
cs delete organization-member [flags]
```

### Examples

```
# Remove organization member
$ cs delete organization-member --org <orgId> -u <userId>
```

### Options

```
  -h, --help       help for organization-member
  -u, --user int   Organization member user ID (default -1)
```

### Options inherited from parent commands

```
  -a, --api string      URL of Codesphere API (can also be CS_API)
  -g, --org string      Organization ID (relevant for some commands)
  -t, --team int        Team ID (relevant for some commands, can also be CS_TEAM_ID) (default -1)
  -v, --verbose         Verbose output
  -w, --workspace int   Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs delete](cs_delete.md)	 - Delete Codesphere resources

