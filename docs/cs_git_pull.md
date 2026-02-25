## cs git pull

Pull latest changes from git repository

### Synopsis

Pull latest changes from the remote git repository.

if specified, pulls a specific branch.

```
cs git pull [flags]
```

### Examples

```
# Pull latest HEAD from current branch
$ cs git pull 

# Pull branch staging from remote origin
$ cs git pull --remote origin --branch staging
```

### Options

```
      --branch string   Branch to pull
  -h, --help            help for pull
      --remote string   Remote to pull from
```

### Options inherited from parent commands

```
  -a, --api string      URL of Codesphere API (can also be CS_API)
  -t, --team int        Team ID (relevant for some commands, can also be CS_TEAM_ID) (default -1)
  -v, --verbose         Verbose output
  -w, --workspace int   Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs git](cs_git.md)	 - Interacting with the git repository of the workspace

