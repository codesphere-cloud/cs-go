package export

import (
	"fmt"
	"os"

	"github.com/codesphere-cloud/cs-go/pkg/ci"
	templates "github.com/codesphere-cloud/cs-go/tmpl/export"
)

func ExportDockerArtifacts(inputPath, outputPath, baseImage string, envVars []string) {
	// Get map from yml file
	ymlContent, err := ci.ReadYmlFile(inputPath)
	if err != nil {
		fmt.Printf("error getting map from yml file: %s\n", err)
		os.Exit(1)
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
		err = templates.CreateDockerfile(config)
		if err != nil {
			fmt.Printf("error creating dockerfile: %s\n", err)
			os.Exit(1)
		}
	}

	// Create nginx config
	fmt.Printf("creating nginx config file\n")

	configNginx := templates.NginxConfigTemplateConfig{
		OutputPath: outputPath,
		Services:   ymlContent.Run,
	}
	err = templates.CreateNginxConfig(configNginx)
	if err != nil {
		fmt.Printf("error creating docker compose file: %s\n", err)
		os.Exit(1)
	}

	// Create Docker compose file
	fmt.Printf("creating docker compose file\n")

	configDockerCompose := templates.DockerComposeTemplateConfig{
		OutputPath: outputPath,
		Services:   ymlContent.Run,
		EnvVars:    envVars,
	}
	err = templates.CreateDockerCompose(configDockerCompose)
	if err != nil {
		fmt.Printf("error creating docker compose file: %s\n", err)
		os.Exit(1)
	}
}
