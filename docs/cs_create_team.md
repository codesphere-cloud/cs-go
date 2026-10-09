## cs create team

Create resource group (team)

### Synopsis

Create a resource group (team) in Codesphere or an Organization

```
cs create team [flags]
```

### Examples

```
# Create a team in a specific datacenter
$ cs create team -d <datacenterId> -n <teamName>

# Create a team in a specific datacenter within an organization
$ cs create team -d <datacenterId> -n <teamName> -g <orgId>
```

### Options

```
  -d, --dc-id int     Data center ID
  -h, --help          help for team
  -n, --name string   Team name
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

* [cs create](cs_create.md)	 - Create codesphere resource

