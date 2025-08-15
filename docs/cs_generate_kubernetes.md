## cs generate kubernetes

Generates kubernetes artifacts based on a ci.yml of a workspace

### Synopsis

The generated artifacts will be saved in the output folder (default is ./export).
In the deployment files the image name is set to '<registry>/<imagePrefix>-<service-name>:latest'.
The nginx router is set to '<registry>/<imagePrefix>-cs-router:latest' as image name.
If the imagePrefix is not set, it uses '<registry>/<service-name>:latest'.
The imagePrefix is used as the namespace for the kubernetes resources, if the prefix is not set, it defaults to 'default'.
It then generates following artifacts inside the output folder:

./<service-n> Each service deployment file is exported to a separate folder.
./<service-n>/<service-n>.yml Kubernetes deployment and service resource to run a pod for the service.
./ingress.yml ingress resource to route traffic to the different services.

Codesphere recommends adding the generated artifacts to the source code repository.

```
cs generate kubernetes [flags]
```

### Examples

```
# Generate kubernetes for workspace 1234
$ cs generate kubernetes -w 1234

# Generate kubernetes for workspace 1234 based on ci profile ci.prod.yml
$ cs generate kubernetes -w 1234 -i ci.prod.yml
```

### Options

```
  -h, --help                  help for kubernetes
      --hostname string       hostname for the ingress to match (default "localhost")
  -p, --imagePrefix string    Image prefix used for the exported images (should be the same as used in generate images)
      --ingressClass string   ingress class for the ingress resource (default "nginx")
  -n, --namespace string      namespace of generated kubernetes artifacts (default "default")
      --pullsecret string     pullsecret for the pod's images (e.g. for a private registry)
  -r, --registry string       Registry where images are pushed to (should be the same as used in generate images)
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

