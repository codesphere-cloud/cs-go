## cs add ssh-key

Add a public SSH key to your account

### Synopsis

Upload an OpenSSH public key for workspace SSH access using the account associated with your API token.

```
cs add ssh-key <public-key-file> [flags]
```

### Examples

```
# Upload your Ed25519 public key
$ cs add ssh-key ~/.ssh/id_ed25519.pub
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

* [cs add](cs_add.md)	 - Add Codesphere resources

