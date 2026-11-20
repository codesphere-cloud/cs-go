## cs delete ssh-key

Delete a public SSH key from your account

### Synopsis

Delete a public SSH key by fingerprint, or select a key interactively to delete it.

```
cs delete ssh-key [fingerprint] [flags]
```

### Examples

```
# Select a public SSH key to delete
$ cs delete ssh-key 

# Delete a public SSH key by fingerprint
$ cs delete ssh-key <fingerprint>
```

### Options

```
  -h, --help   help for ssh-key
```

### Options inherited from parent commands

```
  -a, --api string           URL of Codesphere API (can also be CS_API)
  -g, --org string           Organization ID (relevant for some commands)
  -G, --resource-group int   Resource group ID (relevant for some commands, can also be CS_RESOURCE_GROUP_ID; legacy --team, -t and CS_TEAM_ID still work) (default -1)
  -v, --verbose              Verbose output
  -w, --workspace int        Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs delete](cs_delete.md)	 - Delete Codesphere resources

