## cs add team-member

Add team member

### Synopsis

Add team member to a team.

To add a member to a team within an organization or a standalone team

```
cs add team-member [flags]
```

### Examples

```
# Add a user to a team as a member
$ cs add team-member -t <teamId> -e user@example.com -r 1

# Add a user to a team as an admin
$ cs add team-member -t <teamId> -e admin@example.com -r -1
```

### Options

```
  -e, --email string   Team member email
  -h, --help           help for team-member
  -r, --role int       Team member role 1=member, -1=admin (default 1)
```

### Options inherited from parent commands

```
  -a, --api string      URL of Codesphere API (can also be CS_API)
  -O, --org string      Organization ID (relevant for some commands)
  -t, --team int        Team ID (relevant for some commands, can also be CS_TEAM_ID) (default -1)
  -v, --verbose         Verbose output
  -w, --workspace int   Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs add](cs_add.md)	 - Add Codesphere resources

