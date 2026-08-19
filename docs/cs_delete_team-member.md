## cs delete team-member

Delete team member

### Synopsis

Delete a member from a team.

To delete a member from a team within an organization, the CS_ORG_ID environment variable or the -O/--org flag must be set.

```
cs delete team-member [flags]
```

### Examples

```
# Delete a user from a team
$ cs delete team-member -t <teamId> -u <userId>

# Delete a user from a team within an organization
$ cs delete team-member -O <org-id> -t <teamId> -u <userId>
```

### Options

```
  -h, --help       help for team-member
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

* [cs delete](cs_delete.md)	 - Delete Codesphere resources

