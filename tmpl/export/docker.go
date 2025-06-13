package export

import (
	_ "embed"
	"fmt"
	"text/template"

	"github.com/codesphere-cloud/cs-go/pkg/ci"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
)

//go:embed docker.tmpl
var dockerTemplateFile string

//go:embed shell.tmpl
var entrypointTemplateFile string

type DockerTemplateConfig struct {
	OutputPath string
	// Dockerfile congifuration
	BaseImage    string
	PrepareSteps []ci.Step
	RunSteps     []ci.Step
}

func CreateDockerfile(fs *cs.FileSystem, config DockerTemplateConfig) error {
	err := fs.CreateDirectory(config.OutputPath)
	if err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	// Create Dockerfile
	f, err := fs.CreateFile(config.OutputPath + "/Dockerfile")
	if err != nil {
		return fmt.Errorf("error creating docker file: %w", err)
	}

	dockerTemplate, err := template.New("dockerTemplate").Parse(dockerTemplateFile)
	if err != nil {
		return fmt.Errorf("error parsing docker template: %w", err)
	}

	err = dockerTemplate.Execute(f, config)
	if err != nil {
		return fmt.Errorf("error executing docker template: %w", err)
	}

	// Create shell script for entrypoint
	f, err = fs.CreateFile(config.OutputPath + "/entrypoint.sh")
	if err != nil {
		return fmt.Errorf("error creating entrypoint.sh: %w", err)
	}

	entrypointTemplate, err := template.New("entrypoint").Parse(entrypointTemplateFile)
	if err != nil {
		return fmt.Errorf("error parsing shell template: %w", err)
	}

	err = entrypointTemplate.Execute(f, config)
	if err != nil {
		return fmt.Errorf("error executing shell template: %w", err)
	}

	return f.Close()
}
