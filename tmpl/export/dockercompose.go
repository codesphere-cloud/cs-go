package export

import (
	_ "embed"
	"fmt"
	"text/template"

	"github.com/codesphere-cloud/cs-go/pkg/ci"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
)

//go:embed dockercompose.tmpl
var DockerComposeTemplateFile string

type DockerComposeTemplateConfig struct {
	OutputPath string
	// Docker compose configuration
	Services map[string]ci.Service
	EnvVars  []string
}

func CreateDockerCompose(fs *cs.FileSystem, config DockerComposeTemplateConfig) error {
	err := fs.CreateDirectory(config.OutputPath)
	if err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	// Create the docker compose file
	t, err := template.New("dockercompose.tmpl").Parse(DockerComposeTemplateFile)
	if err != nil {
		return fmt.Errorf("error parsing docker compose template: %w", err)
	}

	f, err := fs.CreateFile(config.OutputPath + "/docker-compose.yml")
	if err != nil {
		return fmt.Errorf("error creating docker compose file: %w", err)
	}

	err = t.Execute(f, config)
	if err != nil {
		return fmt.Errorf("error executing docker compose template: %w", err)
	}

	return f.Close()
}
