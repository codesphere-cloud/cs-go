package export

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/codesphere-cloud/cs-go/pkg/ci"
)

//go:embed docker.tmpl
var dockerTemplateFile string

type DockerTemplateConfig struct {
	BaseImage    string
	PrepareSteps []ci.Step
}

func CreateDockerfile(config DockerTemplateConfig) ([]byte, error) {
	if len(strings.TrimSpace(config.BaseImage)) == 0 {
		return nil, fmt.Errorf("base image is required")
	}

	dockerTemplate, err := template.New("dockerTemplate").Parse(dockerTemplateFile)
	if err != nil {
		return nil, fmt.Errorf("error parsing docker template: %w", err)
	}

	var buf bytes.Buffer
	err = dockerTemplate.Execute(&buf, config)
	if err != nil {
		return nil, fmt.Errorf("error executing docker template: %w", err)
	}

	return buf.Bytes(), nil
}
