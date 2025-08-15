// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/codesphere-cloud/cs-go/pkg/ci"
)

//go:embed dockercompose.tmpl
var dockerComposeTemplateFile string

type DockerComposeTemplateConfig struct {
	Services map[string]ci.Service
	EnvVars  []string
}

func CreateDockerCompose(config DockerComposeTemplateConfig) ([]byte, error) {
	if len(config.Services) == 0 {
		return nil, fmt.Errorf("at least one service is required")
	}
	for serviceName := range config.Services {
		if len(strings.TrimSpace(serviceName)) == 0 {
			return nil, fmt.Errorf("service name cannot be empty")
		}
	}

	t, err := template.New("dockercompose.tmpl").Parse(dockerComposeTemplateFile)
	if err != nil {
		return nil, fmt.Errorf("error parsing docker compose template: %w", err)
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, config)
	if err != nil {
		return nil, fmt.Errorf("error executing docker compose template: %w", err)
	}

	return buf.Bytes(), nil
}
