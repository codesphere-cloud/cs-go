## cs team member list

List team members

### Synopsis

List all members of a team

```
cs team member list [flags]
```

### Examples

```
# List all members of a team
$ cs team member list -t <teamId>

# List all members of a team in JSON format
$ cs team member list -t <teamId> -o json
```

### Options

```
  -h, --help            help for list
  -o, --output string   Output format (table, json, yaml) (default "table")
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

