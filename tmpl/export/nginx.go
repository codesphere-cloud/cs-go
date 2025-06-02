package export

import (
	_ "embed"
	"fmt"
	"os"
	"text/template"

	"github.com/codesphere-cloud/cs-go/pkg/ci"
)

//go:embed nginx.tmpl
var nginxTemplateFile string

type NginxConfigTemplateConfig struct {
	OutputPath string
	// Nginx configuration
	Services map[string]ci.Service
}

func CreateNginxConfig(config NginxConfigTemplateConfig) error {
	t, err := template.New("nginx").Parse(nginxTemplateFile)
	if err != nil {
		return fmt.Errorf("error parsing nginx template: %w\n", err)
	}

	err = CreateDirectory(config.OutputPath)
	if err != nil {
		return fmt.Errorf("error creating directory: %w\n", err)
	}

	f, err := os.Create(config.OutputPath + "/nginx.conf")
	if err != nil {
		return fmt.Errorf("error creating nginx file: %w\n", err)
	}

	err = t.Execute(f, config)
	if err != nil {
		return fmt.Errorf("error executing nginx template: %w\n", err)
	}

	return f.Close()
}
