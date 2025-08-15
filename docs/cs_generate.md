## cs generate

Generate codesphere artifacts

### Synopsis

Collection of commands to generate codesphere related artifacts, such as dockerfiles based on a specific workspace.
If the input file is not found, cs will attempt to clone a branch (default is 'main') of the repository of the workspace
on your local machine to run the artifact generation.

### Options

```
      --branch string     Branch of the repository to clone if the input file is not found (default "main")
  -f, --force             Overwrite any files if existing
  -h, --help              help for generate
  -i, --input string      CI profile to use as input for generation, relative to repository root (default "ci.yml")
  -o, --output string     Output path of the folder including generated artifacts, relative to repository root (default "export")
      --reporoot string   root directory of the workspace repository to export. Will be used to clone the repository if it doesn't exist. (default "./workspace-repo")
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
* [cs generate docker](cs_generate_docker.md)	 - Generates docker artifacts based on a ci.yml of a workspace
* [cs generate images](cs_generate_images.md)	 - Builds and pushes container images from the output folder of the `generate docker` command.
* [cs generate kubernetes](cs_generate_kubernetes.md)	 - Generates kubernetes artifacts based on a ci.yml of a workspace

