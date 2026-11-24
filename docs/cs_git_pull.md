## cs git pull

Pull latest changes from git repository

### Synopsis

Pull latest changes from the remote git repository.

if specified, pulls a specific branch.

The submodule and Git LFS options need Codesphere 1.117.0 or later.
Older versions ignore them.

```
cs git pull [flags]
```

### Examples

```
# Pull latest HEAD from current branch
$ cs git pull 

# Pull branch staging from remote origin
$ cs git pull --remote origin --branch staging

# Pull and update all submodules
$ cs git pull --recurse-submodules

# Pull without downloading Git LFS objects
$ cs git pull --skip-lfs
```

### Options

```
      --branch string        Branch to pull
  -h, --help                 help for pull
      --recurse-submodules   Also initialize and update all submodules after the pull
      --remote string        Remote to pull from
      --skip-lfs             Do not download Git LFS objects, keep their pointer files
      --timeout duration     Timeout for waking up the workspace (default 2m0s)
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

* [cs git](cs_git.md)	 - Interacting with the git repository of the workspace

