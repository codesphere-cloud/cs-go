## cs update organization-member

Change organization member role

### Synopsis

Change organization member role. Select the organization using --org or CS_ORG_ID.

```
cs update organization-member [flags]
```

### Examples

```
# Change organization member role
$ cs update organization-member --org <orgId> -u <userId> -r admin
```

### Options

```
  -h, --help          help for organization-member
  -r, --role string   Organization role (admin, member)
  -u, --user int      Organization member user ID (default -1)
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

* [cs update](cs_update.md)	 - Update Codesphere CLI

