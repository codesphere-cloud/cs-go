package export

import (
	"fmt"

	"github.com/codesphere-cloud/cs-go/pkg/ci"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	templates "github.com/codesphere-cloud/cs-go/tmpl/export"
)

type Exporter interface {
	// ExportDockerArtifacts exports Docker artifacts based on the provided input path, output path, base image, and environment variables.
	ExportDockerArtifacts(inputPath string, outputPath string, baseImage string, envVars []string) error
}

type ExporterService struct {
	fs *cs.FileSystem
}

func NewExporterService(fs *cs.FileSystem) Exporter {
	return &ExporterService{
		fs: fs,
	}
}

func (e *ExporterService) ExportDockerArtifacts(inputPath string, outputPath string, baseImage string, envVars []string) error {
	// Get map from yml file
	ymlContent, err := ci.ReadYmlFile(inputPath)
	if err != nil {
		return fmt.Errorf("error reading yml file: %s", err)
	}

	// Update old services (path directly in network) to network with array of paths
	for serviceName, service := range ymlContent.Run {
		if service.Network.Path != "" {
			service.Network.Paths = []ci.Path{{
				Port:      3000,
				Path:      service.Network.Path,
				StripPath: service.Network.StripPath,
			}}
			service.Network.Ports = []ci.Port{{
				Port:     3000,
				IsPublic: service.IsPublic,
			}}
			ymlContent.Run[serviceName] = service
			fmt.Printf("updated old service %s: %v\n", serviceName, service)
		}
	}

	// Create Dockerfiles
	for serviceName, service := range ymlContent.Run {
		fmt.Printf("creating dockerfile for service %s\n", serviceName)

		config := templates.DockerTemplateConfig{
			OutputPath:   outputPath + "/" + serviceName,
			BaseImage:    baseImage,
			PrepareSteps: ymlContent.Prepare.Steps,
			RunSteps:     service.Steps,
		}
		err = templates.CreateDockerfile(e.fs, config)
		if err != nil {
			return fmt.Errorf("error creating dockerfile for service %s: %s", serviceName, err)
		}
	}

	// Create nginx config
	fmt.Printf("creating nginx config file\n")

	configNginx := templates.NginxConfigTemplateConfig{
		OutputPath: outputPath,
		Services:   ymlContent.Run,
	}
	err = templates.CreateNginxConfig(e.fs, configNginx)
	if err != nil {
		return fmt.Errorf("error creating nginx config file: %s", err)
	}

	// Create Docker compose file
	fmt.Printf("creating docker compose file\n")

	configDockerCompose := templates.DockerComposeTemplateConfig{
		OutputPath: outputPath,
		Services:   ymlContent.Run,
		EnvVars:    envVars,
	}
	err = templates.CreateDockerCompose(e.fs, configDockerCompose)
	if err != nil {
		return fmt.Errorf("error creating docker compose file: %s", err)
	}

	return nil
}
