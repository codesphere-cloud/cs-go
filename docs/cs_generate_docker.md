## cs generate docker

Generates docker artifacts based on a ci.yml of a workspace

### Synopsis

The generated artifacts will be saved in the output folder (default is ./export).
It then generates following artifacts inside the output folder:

./<service-n> Each service is exported to a separate folder.
./<service-n>/dockerfile docker to build the container of the service.
./<service-n>/entrypoint.sh Entrypoint of the container (run stage of Codesphere workspace).
./docker-compose.yml Environment to allow running the services with docker-compose.
./nginx.conf Configuration for NGINX, which is used by as router between services.

Codesphere recommends adding the generated artifacts to the source code repository.

```
cs generate docker [flags]
```

### Examples

```
# Generate docker for workspace 1234
$ cs generate docker -w 1234

# Generate docker for workspace 1234 based on ci profile ci.prod.yml
$ cs generate docker -w 1234 -i ci.prod.yml
```

### Options

```
  -b, --baseimage string   Base image for the docker
  -e, --env stringArray    Env vars to put into generated artifacts
  -h, --help               help for docker
```

### Options inherited from parent commands

```
  -a, --api string        URL of Codesphere API (can also be CS_API)
      --branch string     Branch of the repository to clone if the input file is not found (default "main")
  -f, --force             Overwrite any files if existing
  -i, --input string      CI profile to use as input for generation, relative to repository root (default "ci.yml")
  -o, --output string     Output path of the folder including generated artifacts, relative to repository root (default "export")
      --reporoot string   root directory of the workspace repository to export. Will be used to clone the repository if it doesn't exist. (default "./workspace-repo")
  -t, --team int          Team ID (relevant for some commands, can also be CS_TEAM_ID) (default -1)
  -v, --verbose           Verbose output
  -w, --workspace int     Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs generate](cs_generate.md)	 - Generate codesphere artifacts

