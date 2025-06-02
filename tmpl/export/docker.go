package export

import (
	_ "embed"
	"fmt"
	"os"
	"text/template"

	"github.com/codesphere-cloud/cs-go/pkg/ci"
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

func CreateDockerfile(config DockerTemplateConfig) error {
	err := CreateDirectory(config.OutputPath)
	if err != nil {
		return fmt.Errorf("error creating directory: %w\n", err)
	}

	// Create Dockerfile
	f, err := os.Create(config.OutputPath + "/Dockerfile")
	if err != nil {
		return fmt.Errorf("error creating docker file: %w\n", err)
	}

	dockerTemplate, err := template.New("dockerTemplate").Parse(dockerTemplateFile)
	if err != nil {
		return fmt.Errorf("error parsing docker template: %w\n", err)
	}

	err = dockerTemplate.Execute(f, config)
	if err != nil {
		return fmt.Errorf("error executing docker template: %w\n", err)
	}

	// Create shell script for entrypoint
	f, err = os.Create(config.OutputPath + "/entrypoint.sh")
	if err != nil {
		return fmt.Errorf("error creating entrypoint.sh: %w\n", err)
	}

	entrypointTemplate, err := template.New("entrypoint").Parse(entrypointTemplateFile)
	if err != nil {
		return fmt.Errorf("error parsing shell template: %w\n", err)
	}

	err = entrypointTemplate.Execute(f, config)
	if err != nil {
		return fmt.Errorf("error executing shell template: %w\n", err)
	}

	return f.Close()
}
