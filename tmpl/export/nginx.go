package export

import (
	_ "embed"
	"fmt"
	"text/template"

	"github.com/codesphere-cloud/cs-go/pkg/ci"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
)

//go:embed nginx.tmpl
var nginxTemplateFile string

type NginxConfigTemplateConfig struct {
	OutputPath string
	// Nginx configuration
	Services map[string]ci.Service
}

func CreateNginxConfig(fs *cs.FileSystem, config NginxConfigTemplateConfig) error {
	err := fs.CreateDirectory(config.OutputPath)
	if err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	// Create the nginx config file
	t, err := template.New("nginx").Parse(nginxTemplateFile)
	if err != nil {
		return fmt.Errorf("error parsing nginx template: %w", err)
	}

	f, err := fs.Create(config.OutputPath + "/nginx.conf")
	if err != nil {
		return fmt.Errorf("error creating nginx file: %w", err)
	}

	err = t.Execute(f, config)
	if err != nil {
		return fmt.Errorf("error executing nginx template: %w", err)
	}

	return f.Close()
}
