## cs create resource-group

Create resource group

### Synopsis

Create a resource group in Codesphere or an Organization

```
cs create resource-group [flags]
```

### Examples

```
# Create a resource group in a specific datacenter
$ cs create resource-group -d <datacenterId> -n <resourceGroupName>

# Create a resource group in a specific datacenter within an organization
$ cs create resource-group -d <datacenterId> -n <resourceGroupName> -g <orgId>
```

### Options

```
  -d, --dc-id int     Data center ID
  -h, --help          help for resource-group
  -n, --name string   Resource group name
```

### Options inherited from parent commands

```
  -a, --api string           URL of Codesphere API (can also be CS_API)
  -g, --org string           Organization ID (relevant for some commands)
  -t, --resource-group int   Resource group ID (relevant for some commands, can also be CS_RESOURCE_GROUP_ID; --team and CS_TEAM_ID still work) (default -1)
  -v, --verbose              Verbose output
  -w, --workspace int        Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs create](cs_create.md)	 - Create codesphere resource

