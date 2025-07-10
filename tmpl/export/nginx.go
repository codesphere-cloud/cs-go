package export

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/codesphere-cloud/cs-go/pkg/ci"
)

//go:embed nginx.tmpl
var nginxTemplateFile string

type NginxConfigTemplateConfig struct {
	Services map[string]ci.Service
}

func CreateNginxConfig(config NginxConfigTemplateConfig) ([]byte, error) {
	if len(config.Services) == 0 {
		return nil, fmt.Errorf("at least one service is required")
	}
	for serviceName, service := range config.Services {
		if len(strings.TrimSpace(serviceName)) == 0 {
			return nil, fmt.Errorf("service name cannot be empty")
		}
		for _, path := range service.Network.Paths {
			if len(strings.TrimSpace(path.Path)) == 0 {
				return nil, fmt.Errorf("path cannot be empty")
			}
			if path.Port == 0 {
				return nil, fmt.Errorf("port must be specified for path %s", path.Path)
			}
		}
	}

	t, err := template.New("nginx").Parse(nginxTemplateFile)
	if err != nil {
		return nil, fmt.Errorf("error parsing nginx template: %w", err)
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, config)
	if err != nil {
		return nil, fmt.Errorf("error executing nginx template: %w", err)
	}

	return buf.Bytes(), nil
}
