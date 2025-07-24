## cs open workspace

Open workspace in the Codesphere IDE

### Synopsis

Open workspace in the Codesphere IDE in your web browser.

```
cs open workspace [flags]
```

### Examples

```
# open workspace 42 in web browser
$ cs open workspace -w 42

# open workspace set by environment variable CS_WORKSPACE_ID
$ cs open workspace 
```

### Options

```
  -h, --help   help for workspace
```

### Options inherited from parent commands

```
  -a, --api string      URL of Codesphere API (can also be CS_API)
  -t, --team int        Team ID (relevant for some commands, can also be CS_TEAM_ID) (default -1)
  -v, --verbose         Verbose output
  -w, --workspace int   Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs open](cs_open.md)	 - Open the Codesphere IDE

