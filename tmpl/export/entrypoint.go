package export

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/codesphere-cloud/cs-go/pkg/ci"
)

//go:embed entrypoint.tmpl
var entrypointTemplateFile string

type EntrypointTemplateConfig struct {
	RunSteps []ci.Step
}

func CreateEntrypoint(config EntrypointTemplateConfig) ([]byte, error) {
	if len(config.RunSteps) == 0 {
		return nil, fmt.Errorf("at least one run step is required")
	}

	entrypointTemplate, err := template.New("entrypoint").Parse(entrypointTemplateFile)
	if err != nil {
		return nil, fmt.Errorf("error parsing shell template: %w", err)
	}

	var buf bytes.Buffer
	err = entrypointTemplate.Execute(&buf, config)
	if err != nil {
		return nil, fmt.Errorf("error executing shell template: %w", err)
	}

	return buf.Bytes(), nil
}
