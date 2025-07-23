## cs exec

Run a command in Codesphere workspace

### Synopsis

Run a command in a Codesphere workspace.
Output will be printed to STDOUT, errors to STDERR.

```
cs exec [flags]
```

### Examples

```
# Print `hello world`
$ cs exec -- echo hello world

# List all files in workspace
$ cs exec -- find .

# List all files in the user directory
$ cs exec -d user -- find .

# Set custom environment variables for this command
$ cs exec -e FOO=bar -- 'echo $FOO'
```

### Options

```
  -e, --env stringArray   Additional environment variables to pass to the command in the form key=val
  -h, --help              help for exec
  -d, --workdir string    Working directory for the command
```

### Options inherited from parent commands

```
  -a, --api string      URL of Codesphere API (can also be CS_API)
  -t, --team int        Team ID (relevant for some commands, can also be CS_TEAM_ID) (default -1)
  -w, --workspace int   Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs](cs.md)	 - The Codesphere CLI

