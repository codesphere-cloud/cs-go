## cs log

Retrieve run logs from services

### Synopsis

You can retrieve logs based on the given scope.

If you provide the step number and server, it returns all logs from
all replicas of that server.

If you provide a specific replica id, it will return the logs of
that replica only.

```
cs log [flags]
```

### Examples

```
# Get logs from a server
$ cs log -w 637128 -s app

# Get all logs of all servers
$ cs log -w 637128

# Get logs from a replica
$ cs log -w 637128 -r workspace-213d7a8c-48b4-42e2-8f70-c905ab04abb5-58d657cdc5-m8rrp
```

### Options

```
  -h, --help             help for log
  -r, --replica string   ID of server replica
  -s, --server string    Name of the landscape server
      --stage string     Stage to stream logs from (default "run")
  -n, --step int         Index of execution step (default 0)
```

### Options inherited from parent commands

```
  -a, --api string      URL of Codesphere API (can also be CS_API)
  -t, --team int        Team ID (relevant for some commands, can also be CS_TEAM_ID) (default -1)
  -v, --verbose         Verbose output
  -w, --workspace int   Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs](cs.md)	 - The Codesphere CLI

