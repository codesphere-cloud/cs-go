## cs create organization

Create organization

### Synopsis

Create an organization in Codesphere

```
cs create organization [flags]
```

### Examples

```
# Create an organization with a specific name and admin email
$ cs create organization -n <name> -e <adminEmail>
```

### Options

```
  -e, --admin-email string   Organization admin email
  -h, --help                 help for organization
  -n, --name string          Organization name
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

* [cs create](cs_create.md)	 - Create codesphere resource

