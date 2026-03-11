// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package int_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	intutil "github.com/codesphere-cloud/cs-go/int/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	sigsyaml "sigs.k8s.io/yaml"
)

// Sample ci.yml for testing - simulates the flask-demo structure.
// Reference: https://github.com/codesphere-cloud/flask-demo
const flaskDemoCiYml = `schemaVersion: v0.2
prepare:
  steps:
    - name: install dependencies
      command: pip install -r requirements.txt
test:
  steps: []
run:
  frontend-service:
    steps:
      - command: python app.py
    plan: 21
    replicas: 1
    isPublic: true
    network:
      paths:
        - port: 3000
          path: /
          stripPath: false
      ports:
        - port: 3000
          isPublic: true
  backend-service:
    steps:
      - command: python backend.py
    plan: 21
    replicas: 1
    isPublic: true
    network:
      paths:
        - port: 3000
          path: /api
          stripPath: true
      ports:
        - port: 3000
          isPublic: true
`

// Simple ci.yml with a single service
const simpleCiYml = `schemaVersion: v0.2
prepare:
  steps:
    - name: install
      command: npm install
test:
  steps: []
run:
  web:
    steps:
      - command: npm start
    plan: 21
    replicas: 1
    isPublic: true
    network:
      paths:
        - port: 8080
          path: /
          stripPath: false
      ports:
        - port: 8080
          isPublic: true
`

// Legacy ci.yml format with path directly in network
const legacyCiYml = `schemaVersion: v0.2
prepare:
  steps: []
test:
  steps: []
run:
  app:
    steps:
      - command: ./start.sh
    plan: 21
    replicas: 1
    isPublic: true
    network:
      path: /
      stripPath: true
`

const invalidYaml = `this is not valid yaml:
  - missing proper structure
    broken: [indentation
`

const emptyCiYml = `schemaVersion: v0.2
prepare:
  steps: []
test:
  steps: []
run: {}
`

// splitYAMLDocuments splits a multi-document YAML byte slice on "---" separators.
func splitYAMLDocuments(content []byte) [][]byte {
	docs := bytes.Split(content, []byte("\n---\n"))
	var result [][]byte
	for _, doc := range docs {
		trimmed := bytes.TrimSpace(doc)
		if len(trimmed) > 0 {
			result = append(result, trimmed)
		}
	}
	return result
}

// unmarshalDeployment unmarshals YAML into a Kubernetes Deployment and validates its structure.
func unmarshalDeployment(yamlContent []byte) *apps.Deployment {
	GinkgoHelper()
	deployment := &apps.Deployment{}
	err := sigsyaml.Unmarshal(yamlContent, deployment)
	Expect(err).NotTo(HaveOccurred(), "Failed to unmarshal Deployment YAML")
	Expect(deployment.Kind).To(Equal("Deployment"), "Expected kind Deployment")
	Expect(deployment.APIVersion).To(Equal("apps/v1"), "Expected apiVersion apps/v1")
	Expect(deployment.Name).NotTo(BeEmpty(), "Deployment name should not be empty")
	return deployment
}

// unmarshalService unmarshals YAML into a Kubernetes Service and validates its structure.
func unmarshalService(yamlContent []byte) *core.Service {
	GinkgoHelper()
	service := &core.Service{}
	err := sigsyaml.Unmarshal(yamlContent, service)
	Expect(err).NotTo(HaveOccurred(), "Failed to unmarshal Service YAML")
	Expect(service.Kind).To(Equal("Service"), "Expected kind Service")
	Expect(service.APIVersion).To(Equal("v1"), "Expected apiVersion v1")
	Expect(service.Name).NotTo(BeEmpty(), "Service name should not be empty")
	return service
}

// unmarshalIngress unmarshals YAML into a Kubernetes Ingress and validates its structure.
func unmarshalIngress(yamlContent []byte) *networking.Ingress {
	GinkgoHelper()
	ingress := &networking.Ingress{}
	err := sigsyaml.Unmarshal(yamlContent, ingress)
	Expect(err).NotTo(HaveOccurred(), "Failed to unmarshal Ingress YAML")
	Expect(ingress.Kind).To(Equal("Ingress"), "Expected kind Ingress")
	Expect(ingress.APIVersion).To(Equal("networking.k8s.io/v1"), "Expected apiVersion networking.k8s.io/v1")
	Expect(ingress.Name).NotTo(BeEmpty(), "Ingress name should not be empty")
	return ingress
}

// validateServiceFile validates a K8s service YAML file containing a Deployment and Service.
func validateServiceFile(content []byte) (*apps.Deployment, *core.Service) {
	GinkgoHelper()
	docs := splitYAMLDocuments(content)
	Expect(docs).To(HaveLen(2), "Service file should contain a Deployment and a Service document")
	return unmarshalDeployment(docs[0]), unmarshalService(docs[1])
}

// validateDockerfile performs basic structural validation of a Dockerfile.
func validateDockerfile(content string) {
	GinkgoHelper()
	lines := strings.Split(strings.TrimSpace(content), "\n")
	Expect(len(lines)).To(BeNumerically(">", 0), "Dockerfile should not be empty")

	validInstructions := map[string]bool{
		"FROM": true, "RUN": true, "CMD": true, "LABEL": true,
		"EXPOSE": true, "ENV": true, "ADD": true, "COPY": true,
		"ENTRYPOINT": true, "VOLUME": true, "USER": true,
		"WORKDIR": true, "ARG": true, "ONBUILD": true,
		"STOPSIGNAL": true, "HEALTHCHECK": true, "SHELL": true,
	}

	hasFrom := false
	inContinuation := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			inContinuation = false
			continue
		}
		if inContinuation {
			inContinuation = strings.HasSuffix(trimmed, "\\")
			continue
		}

		parts := strings.Fields(trimmed)
		instruction := strings.ToUpper(parts[0])
		Expect(validInstructions).To(HaveKey(instruction),
			fmt.Sprintf("Invalid Dockerfile instruction: '%s' in line: '%s'", parts[0], trimmed))
		if instruction == "FROM" {
			hasFrom = true
		}
		inContinuation = strings.HasSuffix(trimmed, "\\")
	}
	Expect(hasFrom).To(BeTrue(), "Dockerfile must contain a FROM instruction")
}

// validateShellScript validates that a shell script has a proper shebang and is non-empty.
func validateShellScript(content string) {
	GinkgoHelper()
	Expect(content).NotTo(BeEmpty(), "Shell script should not be empty")
	Expect(content).To(HavePrefix("#!/bin/bash"), "Shell script should start with #!/bin/bash shebang")
}

// validateDockerCompose validates docker-compose.yml content is valid YAML with required structure.
func validateDockerCompose(content []byte) {
	GinkgoHelper()
	var compose map[string]interface{}
	err := sigsyaml.Unmarshal(content, &compose)
	Expect(err).NotTo(HaveOccurred(), "docker-compose.yml should be valid YAML")
	Expect(compose).To(HaveKey("services"), "docker-compose.yml should have a 'services' key")
}

// writeCiYml writes content to a file in the given directory.
func writeCiYml(dir, filename, content string) {
	GinkgoHelper()
	err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644)
	Expect(err).NotTo(HaveOccurred())
}

// readFileContent reads a file and fails the test on error.
func readFileContent(path string) []byte {
	GinkgoHelper()
	content, err := os.ReadFile(path)
	Expect(err).NotTo(HaveOccurred())
	return content
}

// generateDocker runs the generate docker command with common flags.
func generateDocker(tempDir, baseImage, input, output string, extraArgs ...string) string {
	GinkgoHelper()
	args := []string{
		"generate", "docker",
		"--reporoot", tempDir,
		"-b", baseImage,
		"-i", input,
		"-o", output,
	}
	args = append(args, extraArgs...)
	return intutil.RunCommand(args...)
}

// generateKubernetes runs the generate kubernetes command with common flags.
func generateKubernetes(tempDir, registry, input, output string, extraArgs ...string) string {
	GinkgoHelper()
	args := []string{
		"generate", "kubernetes",
		"--reporoot", tempDir,
		"-r", registry,
		"-i", input,
		"-o", output,
	}
	args = append(args, extraArgs...)
	return intutil.RunCommand(args...)
}

// readAndValidateServiceFile reads a K8s service YAML file and validates it.
func readAndValidateServiceFile(path string) (*apps.Deployment, *core.Service) {
	GinkgoHelper()
	return validateServiceFile(readFileContent(path))
}

// readAndValidateIngress reads a K8s ingress YAML file and validates it.
func readAndValidateIngress(path string) *networking.Ingress {
	GinkgoHelper()
	return unmarshalIngress(readFileContent(path))
}

var _ = Describe("Kubernetes Export Integration Tests", Label("local"), func() {
	var (
		tempDir string
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "cs-export-test-")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			Expect(os.RemoveAll(tempDir)).NotTo(HaveOccurred())
		}
	})

	Context("Generate Docker Command", func() {
		It("should generate Dockerfiles and docker-compose from flask-demo ci.yml", func() {
			writeCiYml(tempDir, "ci.yml", flaskDemoCiYml)

			output := generateDocker(tempDir, "ubuntu:latest", "ci.yml", "export")
			Expect(output).To(ContainSubstring("docker artifacts created"))
			Expect(output).To(ContainSubstring("docker compose up"))

			By("Verifying service Dockerfiles and entrypoints")
			for _, tc := range []struct {
				service  string
				entryCmd string
			}{
				{"frontend-service", "python app.py"},
				{"backend-service", "python backend.py"},
			} {
				df := string(readFileContent(filepath.Join(tempDir, "export", tc.service, "Dockerfile")))
				validateDockerfile(df)
				if tc.service == "frontend-service" {
					Expect(df).To(ContainSubstring("FROM ubuntu:latest"))
					Expect(df).To(ContainSubstring("pip install"))
				}

				ep := string(readFileContent(filepath.Join(tempDir, "export", tc.service, "entrypoint.sh")))
				validateShellScript(ep)
				Expect(ep).To(ContainSubstring(tc.entryCmd))
			}

			By("Verifying docker-compose.yml")
			dc := readFileContent(filepath.Join(tempDir, "export", "docker-compose.yml"))
			Expect(string(dc)).To(ContainSubstring("frontend-service"))
			Expect(string(dc)).To(ContainSubstring("backend-service"))
			validateDockerCompose(dc)

			Expect(filepath.Join(tempDir, "export", "nginx.conf")).To(BeAnExistingFile())
		})

		It("should generate Docker artifacts with different base image", func() {
			writeCiYml(tempDir, "ci.yml", simpleCiYml)

			output := generateDocker(tempDir, "alpine:latest", "ci.yml", "export")
			Expect(output).To(ContainSubstring("docker artifacts created"))

			df := string(readFileContent(filepath.Join(tempDir, "export", "web", "Dockerfile")))
			Expect(df).To(ContainSubstring("FROM alpine:latest"))
			validateDockerfile(df)
		})

		It("should fail when baseimage is not provided", func() {
			writeCiYml(tempDir, "ci.yml", simpleCiYml)

			output, exitCode := intutil.RunCommandWithExitCode(
				"generate", "docker",
				"--reporoot", tempDir,
				"-i", "ci.yml",
			)
			Expect(exitCode).NotTo(Equal(0))
			Expect(output).To(ContainSubstring("baseimage is required"))
		})

		It("should fail when ci.yml does not exist", func() {
			_, exitCode := intutil.RunCommandWithExitCode(
				"generate", "docker",
				"--reporoot", tempDir,
				"-b", "ubuntu:latest",
				"-i", "nonexistent.yml",
			)
			Expect(exitCode).NotTo(Equal(0))
		})

		It("should fail with invalid YAML content", func() {
			writeCiYml(tempDir, "ci.yml", invalidYaml)

			_, exitCode := intutil.RunCommandWithExitCode(
				"generate", "docker",
				"--reporoot", tempDir,
				"-b", "ubuntu:latest",
				"-i", "ci.yml",
				"-o", "export",
			)
			Expect(exitCode).NotTo(Equal(0))
		})

		It("should fail with ci.yml with no services", func() {
			writeCiYml(tempDir, "ci.yml", emptyCiYml)

			output, exitCode := intutil.RunCommandWithExitCode(
				"generate", "docker",
				"--reporoot", tempDir,
				"-b", "ubuntu:latest",
				"-i", "ci.yml",
				"-o", "export",
			)
			Expect(exitCode).NotTo(Equal(0))
			Expect(output).To(ContainSubstring("at least one service is required"))
		})
	})

	Context("Generate Kubernetes Command", func() {
		BeforeEach(func() {
			writeCiYml(tempDir, "ci.yml", flaskDemoCiYml)
			output := generateDocker(tempDir, "ubuntu:latest", "ci.yml", "export")
			Expect(output).To(ContainSubstring("docker artifacts created"))
		})

		It("should generate Kubernetes artifacts with registry and namespace", func() {
			output := generateKubernetes(tempDir,
				"ghcr.io/codesphere-cloud/flask-demo", "ci.yml", "export",
				"-p", "cs-demo",
				"-n", "flask-demo",
				"--hostname", "flask-demo.local",
			)
			Expect(output).To(ContainSubstring("Kubernetes artifacts export successful"))
			Expect(output).To(ContainSubstring("kubectl apply"))

			kubernetesDir := filepath.Join(tempDir, "export", "kubernetes")
			info, err := os.Stat(kubernetesDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())

			By("Verifying frontend-service")
			frontDep, frontSvc := readAndValidateServiceFile(filepath.Join(kubernetesDir, "service-frontend-service.yml"))
			Expect(frontDep.Namespace).To(Equal("flask-demo"))
			Expect(frontDep.Spec.Template.Spec.Containers).To(HaveLen(1))
			Expect(frontDep.Spec.Template.Spec.Containers[0].Image).To(Equal("ghcr.io/codesphere-cloud/flask-demo/cs-demo-frontend-service:latest"))
			Expect(frontSvc.Namespace).To(Equal("flask-demo"))

			By("Verifying backend-service")
			backDep, backSvc := readAndValidateServiceFile(filepath.Join(kubernetesDir, "service-backend-service.yml"))
			Expect(backDep.Namespace).To(Equal("flask-demo"))
			Expect(backDep.Spec.Template.Spec.Containers).To(HaveLen(1))
			Expect(backDep.Spec.Template.Spec.Containers[0].Image).To(ContainSubstring("cs-demo-backend-service:latest"))
			Expect(backSvc.Namespace).To(Equal("flask-demo"))

			By("Verifying ingress")
			ingress := readAndValidateIngress(filepath.Join(kubernetesDir, "ingress.yml"))
			Expect(ingress.Namespace).To(Equal("flask-demo"))
			Expect(ingress.Spec.Rules).To(HaveLen(1))
			Expect(ingress.Spec.Rules[0].Host).To(Equal("flask-demo.local"))
			Expect(*ingress.Spec.IngressClassName).To(Equal("nginx"))
		})

		It("should generate Kubernetes artifacts with custom ingress class", func() {
			output := generateKubernetes(tempDir,
				"docker.io/myorg", "ci.yml", "export",
				"-n", "production",
				"--hostname", "myapp.example.com",
				"--ingressClass", "traefik",
			)
			Expect(output).To(ContainSubstring("Kubernetes artifacts export successful"))

			ingress := readAndValidateIngress(filepath.Join(tempDir, "export", "kubernetes", "ingress.yml"))
			Expect(*ingress.Spec.IngressClassName).To(Equal("traefik"))
		})

		It("should generate Kubernetes artifacts with pull secret", func() {
			output := generateKubernetes(tempDir,
				"private-registry.io/myorg", "ci.yml", "export",
				"-n", "staging",
				"--hostname", "staging.myapp.com",
				"--pullsecret", "my-registry-secret",
			)
			Expect(output).To(ContainSubstring("Kubernetes artifacts export successful"))

			dep, _ := readAndValidateServiceFile(filepath.Join(tempDir, "export", "kubernetes", "service-frontend-service.yml"))
			Expect(dep.Spec.Template.Spec.ImagePullSecrets).To(HaveLen(1))
			Expect(dep.Spec.Template.Spec.ImagePullSecrets[0].Name).To(Equal("my-registry-secret"))
		})

		It("should fail when registry is not provided", func() {
			output, exitCode := intutil.RunCommandWithExitCode(
				"generate", "kubernetes",
				"--reporoot", tempDir,
				"-i", "ci.yml",
				"-o", "export",
			)
			Expect(exitCode).NotTo(Equal(0))
			Expect(output).To(ContainSubstring("registry is required"))
		})
	})

	Context("Generate Images Command", func() {
		BeforeEach(func() {
			writeCiYml(tempDir, "ci.yml", simpleCiYml)
			output := generateDocker(tempDir, "ubuntu:latest", "ci.yml", "export")
			Expect(output).To(ContainSubstring("docker artifacts created"))
		})

		It("should fail when registry is not provided for generate images", func() {
			output, exitCode := intutil.RunCommandWithExitCode(
				"generate", "images",
				"--reporoot", tempDir,
				"-i", "ci.yml",
				"-o", "export",
			)
			Expect(exitCode).NotTo(Equal(0))
			Expect(output).To(ContainSubstring("registry is required"))
		})
	})

	Context("Full Export Workflow", func() {
		It("should complete the full export workflow from ci.yml to Kubernetes artifacts", func() {
			writeCiYml(tempDir, "ci.yml", flaskDemoCiYml)

			By("Generating Docker and Kubernetes artifacts")
			dockerOutput := generateDocker(tempDir, "ubuntu:latest", "ci.yml", "export")
			Expect(dockerOutput).To(ContainSubstring("docker artifacts created"))

			k8sOutput := generateKubernetes(tempDir,
				"ghcr.io/codesphere-cloud/flask-demo", "ci.yml", "export",
				"-p", "cs-demo",
				"-n", "flask-demo-ns",
				"--hostname", "colima-cluster",
			)
			Expect(k8sOutput).To(ContainSubstring("Kubernetes artifacts export successful"))

			By("Verifying all expected files exist")
			expectedFiles := []string{
				"export/frontend-service/Dockerfile",
				"export/frontend-service/entrypoint.sh",
				"export/backend-service/Dockerfile",
				"export/backend-service/entrypoint.sh",
				"export/docker-compose.yml",
				"export/nginx.conf",
				"export/Dockerfile.nginx",
				"export/kubernetes/service-frontend-service.yml",
				"export/kubernetes/service-backend-service.yml",
				"export/kubernetes/ingress.yml",
			}
			for _, file := range expectedFiles {
				Expect(filepath.Join(tempDir, file)).To(BeAnExistingFile(), fmt.Sprintf("Expected file %s to exist", file))
			}

			By("Validating Docker artifacts")
			for _, svc := range []string{"frontend-service", "backend-service"} {
				validateDockerfile(string(readFileContent(filepath.Join(tempDir, "export", svc, "Dockerfile"))))
				validateShellScript(string(readFileContent(filepath.Join(tempDir, "export", svc, "entrypoint.sh"))))
			}
			validateDockerCompose(readFileContent(filepath.Join(tempDir, "export", "docker-compose.yml")))

			By("Validating Kubernetes manifests")
			kubernetesDir := filepath.Join(tempDir, "export", "kubernetes")

			ingress := readAndValidateIngress(filepath.Join(kubernetesDir, "ingress.yml"))
			Expect(ingress.Spec.Rules).To(HaveLen(1))
			Expect(ingress.Spec.Rules[0].Host).To(Equal("colima-cluster"))
			ingressPaths := ingress.Spec.Rules[0].HTTP.Paths
			pathStrings := make([]string, len(ingressPaths))
			for i, p := range ingressPaths {
				pathStrings[i] = p.Path
			}
			Expect(pathStrings).To(ContainElements("/", "/api"))

			frontDep, _ := readAndValidateServiceFile(filepath.Join(kubernetesDir, "service-frontend-service.yml"))
			Expect(frontDep.Spec.Template.Spec.Containers[0].Image).To(Equal("ghcr.io/codesphere-cloud/flask-demo/cs-demo-frontend-service:latest"))

			backDep, _ := readAndValidateServiceFile(filepath.Join(kubernetesDir, "service-backend-service.yml"))
			Expect(backDep.Spec.Template.Spec.Containers[0].Image).To(Equal("ghcr.io/codesphere-cloud/flask-demo/cs-demo-backend-service:latest"))
		})

		It("should handle different ci.yml profiles", func() {
			devCiYml := strings.Replace(simpleCiYml, "npm start", "npm run dev", 1)
			writeCiYml(tempDir, "ci.dev.yml", devCiYml)

			prodCiYml := strings.Replace(simpleCiYml, "npm start", "npm run prod", 1)
			writeCiYml(tempDir, "ci.prod.yml", prodCiYml)

			devOutput := generateDocker(tempDir, "node:18", "ci.dev.yml", "export-dev")
			Expect(devOutput).To(ContainSubstring("docker artifacts created"))

			prodOutput := generateDocker(tempDir, "node:18-alpine", "ci.prod.yml", "export-prod")
			Expect(prodOutput).To(ContainSubstring("docker artifacts created"))

			By("Verifying dev and prod have different configurations")
			for _, tc := range []struct {
				exportDir string
				entryCmd  string
				baseImage string
			}{
				{"export-dev", "npm run dev", "FROM node:18"},
				{"export-prod", "npm run prod", "FROM node:18-alpine"},
			} {
				ep := string(readFileContent(filepath.Join(tempDir, tc.exportDir, "web", "entrypoint.sh")))
				validateShellScript(ep)
				Expect(ep).To(ContainSubstring(tc.entryCmd))

				df := string(readFileContent(filepath.Join(tempDir, tc.exportDir, "web", "Dockerfile")))
				validateDockerfile(df)
				Expect(df).To(ContainSubstring(tc.baseImage))
			}
		})
	})

	Context("Legacy ci.yml Format Support", func() {
		It("should handle legacy ci.yml with path directly in network", func() {
			writeCiYml(tempDir, "ci.yml", legacyCiYml)

			dockerOutput := generateDocker(tempDir, "ubuntu:latest", "ci.yml", "export")
			Expect(dockerOutput).To(ContainSubstring("docker artifacts created"))

			k8sOutput := generateKubernetes(tempDir,
				"docker.io/myorg", "ci.yml", "export",
				"-n", "legacy-app",
				"--hostname", "legacy.local",
			)
			Expect(k8sOutput).To(ContainSubstring("Kubernetes artifacts export successful"))

			By("Verifying artifacts are valid")
			validateDockerfile(string(readFileContent(filepath.Join(tempDir, "export", "app", "Dockerfile"))))
			readAndValidateServiceFile(filepath.Join(tempDir, "export", "kubernetes", "service-app.yml"))
			readAndValidateIngress(filepath.Join(tempDir, "export", "kubernetes", "ingress.yml"))
		})
	})

	Context("Environment Variables in Docker Artifacts", func() {
		It("should include environment variables in generated artifacts", func() {
			writeCiYml(tempDir, "ci.yml", simpleCiYml)

			output := generateDocker(tempDir, "node:18", "ci.yml", "export",
				"-e", "NODE_ENV=production",
				"-e", "API_URL=https://api.example.com",
			)
			Expect(output).To(ContainSubstring("docker artifacts created"))

			dc := readFileContent(filepath.Join(tempDir, "export", "docker-compose.yml"))
			Expect(string(dc)).To(ContainSubstring("NODE_ENV"))
			Expect(string(dc)).To(ContainSubstring("API_URL"))
			validateDockerCompose(dc)
		})
	})

	Context("Force Overwrite Behavior", func() {
		It("should overwrite existing files when --force is specified", func() {
			writeCiYml(tempDir, "ci.yml", simpleCiYml)

			output := generateDocker(tempDir, "ubuntu:latest", "ci.yml", "export")
			Expect(output).To(ContainSubstring("docker artifacts created"))

			output = generateDocker(tempDir, "alpine:latest", "ci.yml", "export", "--force")
			Expect(output).To(ContainSubstring("docker artifacts created"))

			df := string(readFileContent(filepath.Join(tempDir, "export", "web", "Dockerfile")))
			Expect(df).To(ContainSubstring("FROM alpine:latest"))
			validateDockerfile(df)
		})
	})

	Context("Generate Command Help", func() {
		It("should display help for generate docker command", func() {
			output := intutil.RunCommand("generate", "docker", "--help")
			Expect(output).To(ContainSubstring("generated artifacts"))
			Expect(output).To(ContainSubstring("-b, --baseimage"))
			Expect(output).To(ContainSubstring("-i, --input"))
			Expect(output).To(ContainSubstring("-o, --output"))
		})

		It("should display help for generate kubernetes command", func() {
			output := intutil.RunCommand("generate", "kubernetes", "--help")
			Expect(output).To(ContainSubstring("generated artifacts"))
			Expect(output).To(ContainSubstring("-r, --registry"))
			Expect(output).To(ContainSubstring("-p, --imagePrefix"))
			Expect(output).To(ContainSubstring("-n, --namespace"))
			Expect(output).To(ContainSubstring("--hostname"))
			Expect(output).To(ContainSubstring("--pullsecret"))
			Expect(output).To(ContainSubstring("--ingressClass"))
		})

		It("should display help for generate images command", func() {
			output := intutil.RunCommand("generate", "images", "--help")
			Expect(output).To(ContainSubstring("generated images will be pushed"))
			Expect(output).To(ContainSubstring("-r, --registry"))
			Expect(output).To(ContainSubstring("-p, --imagePrefix"))
		})
	})
})
