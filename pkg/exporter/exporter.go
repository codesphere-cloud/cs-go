// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package exporter

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/url"
	"path/filepath"

	"github.com/codesphere-cloud/cs-go/pkg/ci"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	templates "github.com/codesphere-cloud/cs-go/tmpl/docker"
	"github.com/codesphere-cloud/cs-go/tmpl/k8s"
)

type Exporter interface {
	ReadYmlFile(path string) (*ci.CiYml, error)
	ExportDockerArtifacts() error
	ExportKubernetesArtifacts(registry string, image string, namespace string, pullSecret string, hostname string, ingressClass string) error
	ExportImages(ctx context.Context, registry string, imagePrefix string) error
}

type ExporterService struct {
	fs         *cs.FileSystem
	ymlContent *ci.CiYml
	outputPath string
	baseImage  string
	envVars    []string
	repoRoot   string
	force      bool
}

func NewExporterService(fs *cs.FileSystem, outputPath string, baseImage string, envVars []string, repoRoot string, force bool) Exporter {
	return &ExporterService{
		fs:         fs,
		outputPath: outputPath,
		baseImage:  baseImage,
		envVars:    envVars,
		repoRoot:   repoRoot,
		force:      force,
	}
}

// ReadYmlFile reads the CI YML file from the given path.
func (e *ExporterService) ReadYmlFile(path string) (*ci.CiYml, error) {
	ymlContent, err := ci.ReadYmlFile(e.fs, path)
	if err != nil {
		return nil, fmt.Errorf("error reading yml file: %w", err)
	}

	e.ymlContent = ymlContent

	return ymlContent, nil
}

func (e *ExporterService) GetExportDir() string {
	return filepath.Join(e.repoRoot, e.outputPath)

}

func (e *ExporterService) GetKubernetesDir() string {
	return filepath.Join(e.repoRoot, e.outputPath, "kubernetes")

}

// ExportDockerArtifacts exports Docker artifacts based on the provided input path, output path, base image, and environment variables.
// ReadYmlFile has to be called before this method.
func (e *ExporterService) ExportDockerArtifacts() error {
	if e.baseImage == "" {
		return fmt.Errorf("baseimage is not set, call Setup first")
	}
	if e.ymlContent == nil {
		return fmt.Errorf("yml content is not set, call ReadYmlFile first")
	}

	// Create Dockerfiles and entrypoints for each service
	for serviceName, service := range e.ymlContent.Run {
		log.Printf("Creating dockerfile and entrypoint for service %s\n", serviceName)

		configDocker := templates.DockerTemplateConfig{
			BaseImage:    e.baseImage,
			PrepareSteps: e.ymlContent.Prepare.Steps,
			Entrypoint:   filepath.Join(e.outputPath, serviceName, "entrypoint.sh"),
		}
		dockerfile, err := templates.CreateDockerfile(configDocker)
		if err != nil {
			return fmt.Errorf("error creating dockerfile for service %s: %w", serviceName, err)
		}
		log.Println(e.outputPath)
		log.Println(e.GetExportDir())
		log.Println(filepath.Join(e.GetExportDir(), serviceName))
		err = e.fs.WriteFile(filepath.Join(e.GetExportDir(), serviceName), "Dockerfile", dockerfile, e.force)
		if err != nil {
			return fmt.Errorf("error writing dockerfile for service %s: %w", serviceName, err)
		}

		configEntrypoint := templates.EntrypointTemplateConfig{
			RunSteps: service.Steps,
		}
		entrypointFile, err := templates.CreateEntrypoint(configEntrypoint)
		if err != nil {
			return fmt.Errorf("error creating entrypoint for service %s: %w", serviceName, err)
		}
		err = e.fs.WriteFile(filepath.Join(e.GetExportDir(), serviceName), "entrypoint.sh", entrypointFile, e.force)
		if err != nil {
			return fmt.Errorf("error writing entrypoint for service %s: %w", serviceName, err)
		}
	}

	log.Printf("Creating nginx config file and nginx dockerfile\n")
	// Create nginx config
	configNginx := templates.NginxConfigTemplateConfig{
		Services: e.ymlContent.Run,
	}
	nginxFile, err := templates.CreateNginxConfig(configNginx)
	if err != nil {
		return fmt.Errorf("error creating nginx config file: %s", err)
	}
	err = e.fs.WriteFile(e.GetExportDir(), "nginx.conf", nginxFile, e.force)
	if err != nil {
		return fmt.Errorf("error writing nginx config file: %s", err)
	}

	// Create nginx Dockerfile
	nginxDockerfile := templates.CreateNginxDockerfile()
	err = e.fs.WriteFile(e.GetExportDir(), "Dockerfile.nginx", nginxDockerfile, e.force)
	if err != nil {
		return fmt.Errorf("error writing nginx dockerfile: %s", err)
	}

	// Create docker-compose file
	log.Printf("Creating docker-compose file\n")

	configDockerCompose := templates.DockerComposeTemplateConfig{
		Services: e.ymlContent.Run,
		EnvVars:  e.envVars,
	}
	dockerComposeFile, err := templates.CreateDockerCompose(configDockerCompose)
	if err != nil {
		return fmt.Errorf("error creating docker compose file: %s", err)
	}
	err = e.fs.WriteFile(e.GetExportDir(), "docker-compose.yml", dockerComposeFile, e.force)
	if err != nil {
		return fmt.Errorf("error writing docker compose file: %s", err)
	}

	return nil
}

// ExportKubernetesArtifacts generates Kubernetes artifacts for each service defined in the CI YML file.
// ExportDockerArtifacts has to be called before this method.
func (e *ExporterService) ExportKubernetesArtifacts(registry string, imagePrefix string, namespace string, pullSecret string, hostname string, ingressClass string) error {
	if e.ymlContent == nil {
		return fmt.Errorf("yml content is not set, call ReadYmlFile first")
	}

	// Create deployment and service for each service
	for serviceName, service := range e.ymlContent.Run {
		log.Printf("Creating deployment for service %s\n", serviceName)

		tag, err := e.CreateImageTag(registry, imagePrefix, serviceName)
		if err != nil {
			return fmt.Errorf("error creating image tag from registry and image prefix: %w", err)
		}

		deployment, err := k8s.GenerateDeploymentTemplate(serviceName, namespace, tag, pullSecret)
		if err != nil {
			return fmt.Errorf("error creating deployment for service %s: %w", serviceName, err)
		}

		if len(service.Network.Ports) == 0 {
			service.Network.Ports = []ci.Port{
				{
					Port:     3000,
					IsPublic: true,
				},
			}
		}
		service, err := k8s.GenerateServiceTemplate(serviceName, namespace, service.Network.Ports)
		if err != nil {
			return fmt.Errorf("error creating service for service %s: %w", serviceName, err)
		}

		var b bytes.Buffer
		b.Write(deployment)
		b.WriteString("\n---\n")
		b.Write(service)
		filename := fmt.Sprintf("service-%s.yml", serviceName)
		err = e.fs.WriteFile(e.GetKubernetesDir(), filename, b.Bytes(), e.force)
		if err != nil {
			return fmt.Errorf("error writing service deployment file for service %s: %w", serviceName, err)
		}
	}

	// Create kubernetes ingress
	ingress, err := k8s.GenerateIngressTemplate(e.ymlContent, namespace, hostname, ingressClass)
	if err != nil {
		return fmt.Errorf("error creating ingress: %w", err)
	}
	err = e.fs.WriteFile(e.GetKubernetesDir(), "ingress.yml", ingress, e.force)
	if err != nil {
		return fmt.Errorf("error writing ingress file: %w", err)
	}

	return nil
}

// ExportImages builds and pushes Docker images for each service defined in the CI YML file.
// ExportDockerArtifacts has to be called before this method.
func (e *ExporterService) ExportImages(ctx context.Context, registry string, imagePrefix string) error {
	// Build and push service docker images
	for serviceName := range e.ymlContent.Run {
		tag, err := e.CreateImageTag(registry, imagePrefix, serviceName)
		if err != nil {
			return fmt.Errorf("error creating image tag from registry and image prefix: %w", err)
		}

		servicePath := filepath.Join(e.GetExportDir(), serviceName)
		log.Printf("Building image in %v\n", servicePath)
		err = BuildImage(ctx, filepath.Join(e.outputPath, serviceName, "Dockerfile"), tag, e.repoRoot)
		if err != nil {
			return fmt.Errorf("error building %v image: %s", serviceName, err)
		}

		log.Printf("Pushing image %s\n", tag)
		err = PushImage(ctx, tag)
		if err != nil {
			return fmt.Errorf("error pushing %v image: %s", serviceName, err)
		}
	}

	return nil
}

// CreateImageTag creates a Docker image tag from the registry, image prefix and service name.
// It returns the full image tag in the format: <registry>/<imagePrefix>-<serviceName>:latest.
func (e *ExporterService) CreateImageTag(registry string, imagePrefix string, serviceName string) (string, error) {
	log.Println(imagePrefix)
	if imagePrefix == "" {
		tag, err := url.JoinPath(registry, fmt.Sprintf("%s:latest", serviceName))
		if err != nil {
			return "", err
		}
		return tag, nil
	}

	return fmt.Sprintf("%s/%s-%s:latest", registry, imagePrefix, serviceName), nil
}
