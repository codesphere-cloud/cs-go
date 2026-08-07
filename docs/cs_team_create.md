## cs team create

Create team

### Synopsis

Create a team in Codesphere or an Organization

```
cs team create [flags]
```

### Examples

```
# Create a team in a specific datacenter
$ cs team create -d <datacenterId> -n <teamName>

# Create a team in a specific datacenter within an organization
$ cs team create -d <datacenterId> -n <teamName> -O <orgId>
```

### Options

```
  -d, --dc-id int     Data center ID
  -h, --help          help for create
  -n, --name string   Team name
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

* [cs team](cs_team.md)	 - Manage Team

