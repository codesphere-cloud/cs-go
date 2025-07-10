package export

import (
	"fmt"
	"path/filepath"

	"github.com/codesphere-cloud/cs-go/pkg/ci"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	templates "github.com/codesphere-cloud/cs-go/tmpl/export"
)

type Exporter interface {
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

// ExportDockerArtifacts exports Docker artifacts based on the provided input path, output path, base image, and environment variables.
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
			fmt.Printf("Updated old service %s: %v\n", serviceName, service)
		}
	}

	// Create Dockerfiles and entrypoints for each service
	for serviceName, service := range ymlContent.Run {
		fmt.Printf("Creating dockerfile and entrypoint for service %s\n", serviceName)
		servicePath := filepath.Join(outputPath, serviceName)

		configDocker := templates.DockerTemplateConfig{
			BaseImage:    baseImage,
			PrepareSteps: ymlContent.Prepare.Steps,
		}
		dockerfile, err := templates.CreateDockerfile(configDocker)
		if err != nil {
			return fmt.Errorf("error creating dockerfile for service %s: %w", serviceName, err)
		}
		err = e.fs.WriteFile(servicePath, "Dockerfile", dockerfile)
		if err != nil {
			return fmt.Errorf("error writing dockerfile for service %s: %w", serviceName, err)
		}

		configEntrypoint := templates.EntrypointTemplateConfig{
			RunSteps: service.Steps,
		}
		entrypointFile, err := templates.CreateEntrypoint(configEntrypoint)
		if err != nil {
			return fmt.Errorf("error creating dockerfile for service %s: %w", serviceName, err)
		}
		err = e.fs.WriteFile(servicePath, "entrypoint.sh", entrypointFile)
		if err != nil {
			return fmt.Errorf("error writing entrypoint for service %s: %w", serviceName, err)
		}
	}

	// Create nginx config
	fmt.Printf("Creating nginx config file\n")

	configNginx := templates.NginxConfigTemplateConfig{
		Services: ymlContent.Run,
	}
	nginxFile, err := templates.CreateNginxConfig(configNginx)
	if err != nil {
		return fmt.Errorf("error creating nginx config file: %s", err)
	}
	err = e.fs.WriteFile(outputPath, "nginx.conf", nginxFile)
	if err != nil {
		return fmt.Errorf("error writing nginx config file: %s", err)
	}

	// Create docker-compose file
	fmt.Printf("Creating docker-compose file\n")

	configDockerCompose := templates.DockerComposeTemplateConfig{
		Services: ymlContent.Run,
		EnvVars:  envVars,
	}
	dockerComposeFile, err := templates.CreateDockerCompose(configDockerCompose)
	if err != nil {
		return fmt.Errorf("error creating docker compose file: %s", err)
	}
	err = e.fs.WriteFile(outputPath, "docker-compose.yml", dockerComposeFile)
	if err != nil {
		return fmt.Errorf("error writing docker compose file: %s", err)
	}

	return nil
}
