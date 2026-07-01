## cs team member remove

Remove team member

### Synopsis

Remove member from a team.

To remove a member from a team within an organization, the CS_ORG_ID environment variable or the -O/--org flag must be set.

```
cs team member remove [flags]
```

### Examples

```
# Remove a user from a team
$ cs team member remove -t <teamId> -u <userId>

# Remove a user from a team within an organization
$ cs team member remove -O <org-id> -t <teamId> -u <userId>
```

### Options

```
  -h, --help       help for remove
  -u, --user int   Team member user ID
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

* [cs team member](cs_team_member.md)	 - Manage team members

