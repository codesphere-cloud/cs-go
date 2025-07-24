## cs generate dockerfile

Generates a dockerfile based on a ci.yml of a workspace

### Synopsis

If the input file is not found, cs will attempt to clone the repository of the workspace
on your local machine to run the artifact generation.
For that a folder will be generated containing the repository and the generated artifacts.

The export then generates a subdirectory containing the following artifacts:

./<service-n> Each service is exported to a separate folder.
./<service-n>/Dockerfile Dockerfile to build the container of the service.
./<service-n>/entrypoint.sh Entrypoint of the container (run stage of Codesphere workspace).
./docker-compose.yml Environment to allow running the services with docker-compose.
./export/nginx.conf Configuration for NGINX, which is used by as router between services.

Codesphere recommends adding the generated artifacts to the source code repository.

```
cs generate dockerfile [flags]
```

### Examples

```
# Generate dockerfile for workspace 1234
$ cs generate dockerfile -w 1234

# Generate dockerfile for workspace 1234 based on ci profile ci.prod.yml
$ cs generate dockerfile -w 1234 -i ci.prod.yml
```

### Options

```
  -b, --baseimage string   Base image for the dockerfile
      --branch string      Branch of the repository to clone if the input file is not found (default "main")
  -e, --env stringArray    Env vars to put into generated artifacts
  -h, --help               help for dockerfile
  -i, --input string       CI profile to use as input for generation (default "ci.yml")
  -o, --output string      Output path of the folder including generated artifacts (default "./export")
```

### Options inherited from parent commands

```
  -a, --api string      URL of Codesphere API (can also be CS_API)
  -t, --team int        Team ID (relevant for some commands, can also be CS_TEAM_ID) (default -1)
  -v, --verbose         Verbose output
  -w, --workspace int   Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs generate](cs_generate.md)	 - Generate codesphere artifacts

