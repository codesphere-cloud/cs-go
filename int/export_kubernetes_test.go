// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package int_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	intutil "github.com/codesphere-cloud/cs-go/int/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Sample ci.yml for testing - simulates the flask-demo structure from the blog post
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

var _ = Describe("Kubernetes Export Integration Tests", func() {
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
			By("Creating ci.yml in temp directory")
			ciYmlPath := filepath.Join(tempDir, "ci.yml")
			err := os.WriteFile(ciYmlPath, []byte(flaskDemoCiYml), 0644)
			Expect(err).NotTo(HaveOccurred())

			By("Running generate docker command")
			output := intutil.RunCommand(
				"generate", "docker",
				"--reporoot", tempDir,
				"-b", "ubuntu:latest",
				"-i", "ci.yml",
				"-o", "export",
			)
			fmt.Printf("Generate docker output: %s\n", output)

			Expect(output).To(ContainSubstring("docker artifacts created"))
			Expect(output).To(ContainSubstring("docker compose up"))

			By("Verifying frontend-service Dockerfile was created")
			frontendDockerfile := filepath.Join(tempDir, "export", "frontend-service", "Dockerfile")
			Expect(frontendDockerfile).To(BeAnExistingFile())
			content, err := os.ReadFile(frontendDockerfile)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("FROM ubuntu:latest"))
			Expect(string(content)).To(ContainSubstring("pip install"))

			By("Verifying frontend-service entrypoint was created")
			frontendEntrypoint := filepath.Join(tempDir, "export", "frontend-service", "entrypoint.sh")
			Expect(frontendEntrypoint).To(BeAnExistingFile())
			content, err = os.ReadFile(frontendEntrypoint)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("python app.py"))

			By("Verifying backend-service Dockerfile was created")
			backendDockerfile := filepath.Join(tempDir, "export", "backend-service", "Dockerfile")
			Expect(backendDockerfile).To(BeAnExistingFile())

			By("Verifying backend-service entrypoint was created")
			backendEntrypoint := filepath.Join(tempDir, "export", "backend-service", "entrypoint.sh")
			Expect(backendEntrypoint).To(BeAnExistingFile())
			content, err = os.ReadFile(backendEntrypoint)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("python backend.py"))

			By("Verifying docker-compose.yml was created")
			dockerComposePath := filepath.Join(tempDir, "export", "docker-compose.yml")
			Expect(dockerComposePath).To(BeAnExistingFile())
			content, err = os.ReadFile(dockerComposePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("frontend-service"))
			Expect(string(content)).To(ContainSubstring("backend-service"))

			By("Verifying nginx config was created")
			nginxConfigPath := filepath.Join(tempDir, "export", "nginx.conf")
			Expect(nginxConfigPath).To(BeAnExistingFile())
		})

		It("should generate Docker artifacts with different base image", func() {
			By("Creating ci.yml in temp directory")
			ciYmlPath := filepath.Join(tempDir, "ci.yml")
			err := os.WriteFile(ciYmlPath, []byte(simpleCiYml), 0644)
			Expect(err).NotTo(HaveOccurred())

			By("Running generate docker with alpine base image")
			output := intutil.RunCommand(
				"generate", "docker",
				"--reporoot", tempDir,
				"-b", "alpine:latest",
				"-i", "ci.yml",
				"-o", "export",
			)
			fmt.Printf("Generate docker output: %s\n", output)

			Expect(output).To(ContainSubstring("docker artifacts created"))

			By("Verifying Dockerfile uses alpine base image")
			dockerfile := filepath.Join(tempDir, "export", "web", "Dockerfile")
			Expect(dockerfile).To(BeAnExistingFile())
			content, err := os.ReadFile(dockerfile)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("FROM alpine:latest"))
		})

		It("should fail when baseimage is not provided", func() {
			By("Creating ci.yml in temp directory")
			ciYmlPath := filepath.Join(tempDir, "ci.yml")
			err := os.WriteFile(ciYmlPath, []byte(simpleCiYml), 0644)
			Expect(err).NotTo(HaveOccurred())

			By("Running generate docker without baseimage")
			output, exitCode := intutil.RunCommandWithExitCode(
				"generate", "docker",
				"--reporoot", tempDir,
				"-i", "ci.yml",
			)
			fmt.Printf("Generate docker without baseimage output: %s (exit code: %d)\n", output, exitCode)

			Expect(exitCode).NotTo(Equal(0))
			Expect(output).To(ContainSubstring("baseimage is required"))
		})

		It("should fail when ci.yml does not exist", func() {
			By("Running generate docker without ci.yml")
			output, exitCode := intutil.RunCommandWithExitCode(
				"generate", "docker",
				"--reporoot", tempDir,
				"-b", "ubuntu:latest",
				"-i", "nonexistent.yml",
			)
			fmt.Printf("Generate docker with nonexistent file output: %s (exit code: %d)\n", output, exitCode)

			Expect(exitCode).NotTo(Equal(0))
		})
	})

	Context("Generate Kubernetes Command", func() {
		BeforeEach(func() {
			By("Creating ci.yml and generating docker artifacts first")
			ciYmlPath := filepath.Join(tempDir, "ci.yml")
			err := os.WriteFile(ciYmlPath, []byte(flaskDemoCiYml), 0644)
			Expect(err).NotTo(HaveOccurred())

			output := intutil.RunCommand(
				"generate", "docker",
				"--reporoot", tempDir,
				"-b", "ubuntu:latest",
				"-i", "ci.yml",
				"-o", "export",
			)
			Expect(output).To(ContainSubstring("docker artifacts created"))
		})

		It("should generate Kubernetes artifacts with registry and namespace", func() {
			By("Running generate kubernetes command")
			output := intutil.RunCommand(
				"generate", "kubernetes",
				"--reporoot", tempDir,
				"-r", "ghcr.io/codesphere-cloud/flask-demo",
				"-p", "cs-demo",
				"-i", "ci.yml",
				"-o", "export",
				"-n", "flask-demo",
				"--hostname", "flask-demo.local",
			)
			fmt.Printf("Generate kubernetes output: %s\n", output)

			Expect(output).To(ContainSubstring("Kubernetes artifacts export successful"))
			Expect(output).To(ContainSubstring("kubectl apply"))

			By("Verifying kubernetes directory was created")
			kubernetesDir := filepath.Join(tempDir, "export", "kubernetes")
			info, err := os.Stat(kubernetesDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())

			By("Verifying frontend-service deployment was created")
			frontendService := filepath.Join(kubernetesDir, "service-frontend-service.yml")
			Expect(frontendService).To(BeAnExistingFile())
			content, err := os.ReadFile(frontendService)
			Expect(err).NotTo(HaveOccurred())
			contentStr := string(content)
			Expect(contentStr).To(ContainSubstring("kind: Deployment"))
			Expect(contentStr).To(ContainSubstring("kind: Service"))
			Expect(contentStr).To(ContainSubstring("namespace: flask-demo"))
			Expect(contentStr).To(ContainSubstring("ghcr.io/codesphere-cloud/flask-demo/cs-demo-frontend-service:latest"))

			By("Verifying backend-service deployment was created")
			backendService := filepath.Join(kubernetesDir, "service-backend-service.yml")
			Expect(backendService).To(BeAnExistingFile())
			content, err = os.ReadFile(backendService)
			Expect(err).NotTo(HaveOccurred())
			contentStr = string(content)
			Expect(contentStr).To(ContainSubstring("kind: Deployment"))
			Expect(contentStr).To(ContainSubstring("namespace: flask-demo"))
			Expect(contentStr).To(ContainSubstring("cs-demo-backend-service:latest"))

			By("Verifying ingress was created")
			ingressPath := filepath.Join(kubernetesDir, "ingress.yml")
			Expect(ingressPath).To(BeAnExistingFile())
			content, err = os.ReadFile(ingressPath)
			Expect(err).NotTo(HaveOccurred())
			contentStr = string(content)
			Expect(contentStr).To(ContainSubstring("kind: Ingress"))
			Expect(contentStr).To(ContainSubstring("namespace: flask-demo"))
			Expect(contentStr).To(ContainSubstring("host: flask-demo.local"))
			Expect(contentStr).To(ContainSubstring("ingressClassName: nginx"))
		})

		It("should generate Kubernetes artifacts with custom ingress class", func() {
			By("Running generate kubernetes with custom ingress class")
			output := intutil.RunCommand(
				"generate", "kubernetes",
				"--reporoot", tempDir,
				"-r", "docker.io/myorg",
				"-i", "ci.yml",
				"-o", "export",
				"-n", "production",
				"--hostname", "myapp.example.com",
				"--ingressClass", "traefik",
			)
			fmt.Printf("Generate kubernetes with traefik output: %s\n", output)

			Expect(output).To(ContainSubstring("Kubernetes artifacts export successful"))

			By("Verifying ingress uses traefik class")
			ingressPath := filepath.Join(tempDir, "export", "kubernetes", "ingress.yml")
			content, err := os.ReadFile(ingressPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("ingressClassName: traefik"))
		})

		It("should generate Kubernetes artifacts with pull secret", func() {
			By("Running generate kubernetes with pull secret")
			output := intutil.RunCommand(
				"generate", "kubernetes",
				"--reporoot", tempDir,
				"-r", "private-registry.io/myorg",
				"-i", "ci.yml",
				"-o", "export",
				"-n", "staging",
				"--hostname", "staging.myapp.com",
				"--pullsecret", "my-registry-secret",
			)
			fmt.Printf("Generate kubernetes with pull secret output: %s\n", output)

			Expect(output).To(ContainSubstring("Kubernetes artifacts export successful"))

			By("Verifying deployment includes pull secret")
			frontendService := filepath.Join(tempDir, "export", "kubernetes", "service-frontend-service.yml")
			content, err := os.ReadFile(frontendService)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("my-registry-secret"))
		})

		It("should fail when registry is not provided", func() {
			By("Running generate kubernetes without registry")
			output, exitCode := intutil.RunCommandWithExitCode(
				"generate", "kubernetes",
				"--reporoot", tempDir,
				"-i", "ci.yml",
				"-o", "export",
			)
			fmt.Printf("Generate kubernetes without registry output: %s (exit code: %d)\n", output, exitCode)

			Expect(exitCode).NotTo(Equal(0))
			Expect(output).To(ContainSubstring("registry is required"))
		})
	})

	Context("Full Export Workflow", func() {
		It("should complete the full export workflow from ci.yml to Kubernetes artifacts", func() {
			By("Step 1: Creating ci.yml with multi-service application")
			ciYmlPath := filepath.Join(tempDir, "ci.yml")
			err := os.WriteFile(ciYmlPath, []byte(flaskDemoCiYml), 0644)
			Expect(err).NotTo(HaveOccurred())

			By("Step 2: Generate Docker artifacts")
			dockerOutput := intutil.RunCommand(
				"generate", "docker",
				"--reporoot", tempDir,
				"-b", "ubuntu:latest",
				"-i", "ci.yml",
				"-o", "export",
			)
			fmt.Printf("Docker generation output: %s\n", dockerOutput)
			Expect(dockerOutput).To(ContainSubstring("docker artifacts created"))

			// Verify Docker artifacts
			Expect(filepath.Join(tempDir, "export", "frontend-service", "Dockerfile")).To(BeAnExistingFile())
			Expect(filepath.Join(tempDir, "export", "backend-service", "Dockerfile")).To(BeAnExistingFile())
			Expect(filepath.Join(tempDir, "export", "docker-compose.yml")).To(BeAnExistingFile())

			By("Step 3: Generate Kubernetes artifacts")
			k8sOutput := intutil.RunCommand(
				"generate", "kubernetes",
				"--reporoot", tempDir,
				"-r", "ghcr.io/codesphere-cloud/flask-demo",
				"-p", "cs-demo",
				"-i", "ci.yml",
				"-o", "export",
				"-n", "flask-demo-ns",
				"--hostname", "colima-cluster",
			)
			fmt.Printf("Kubernetes generation output: %s\n", k8sOutput)
			Expect(k8sOutput).To(ContainSubstring("Kubernetes artifacts export successful"))

			By("Step 4: Verify all expected files exist")
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
				fullPath := filepath.Join(tempDir, file)
				Expect(fullPath).To(BeAnExistingFile(), fmt.Sprintf("Expected file %s to exist", file))
			}

			By("Step 5: Verify Kubernetes manifests are valid YAML with correct content")
			kubernetesDir := filepath.Join(tempDir, "export", "kubernetes")

			// Check ingress contains all services
			ingressContent, err := os.ReadFile(filepath.Join(kubernetesDir, "ingress.yml"))
			Expect(err).NotTo(HaveOccurred())
			ingressStr := string(ingressContent)
			Expect(ingressStr).To(ContainSubstring("host: colima-cluster"))
			Expect(ingressStr).To(ContainSubstring("frontend-service"))
			Expect(ingressStr).To(ContainSubstring("backend-service"))
			Expect(ingressStr).To(ContainSubstring("path: /"))
			Expect(ingressStr).To(ContainSubstring("path: /api"))

			// Check frontend service has correct image
			frontendContent, err := os.ReadFile(filepath.Join(kubernetesDir, "service-frontend-service.yml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(frontendContent)).To(ContainSubstring("image: ghcr.io/codesphere-cloud/flask-demo/cs-demo-frontend-service:latest"))

			// Check backend service has correct image
			backendContent, err := os.ReadFile(filepath.Join(kubernetesDir, "service-backend-service.yml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(backendContent)).To(ContainSubstring("image: ghcr.io/codesphere-cloud/flask-demo/cs-demo-backend-service:latest"))
		})

		It("should handle different ci.yml profiles", func() {
			By("Creating multiple ci.yml profiles")
			// Dev profile
			devCiYml := strings.Replace(simpleCiYml, "npm start", "npm run dev", 1)
			err := os.WriteFile(filepath.Join(tempDir, "ci.dev.yml"), []byte(devCiYml), 0644)
			Expect(err).NotTo(HaveOccurred())

			// Prod profile
			prodCiYml := strings.Replace(simpleCiYml, "npm start", "npm run prod", 1)
			err = os.WriteFile(filepath.Join(tempDir, "ci.prod.yml"), []byte(prodCiYml), 0644)
			Expect(err).NotTo(HaveOccurred())

			By("Generating Docker artifacts for dev profile")
			devDockerOutput := intutil.RunCommand(
				"generate", "docker",
				"--reporoot", tempDir,
				"-b", "node:18",
				"-i", "ci.dev.yml",
				"-o", "export-dev",
			)
			Expect(devDockerOutput).To(ContainSubstring("docker artifacts created"))

			By("Generating Docker artifacts for prod profile")
			prodDockerOutput := intutil.RunCommand(
				"generate", "docker",
				"--reporoot", tempDir,
				"-b", "node:18-alpine",
				"-i", "ci.prod.yml",
				"-o", "export-prod",
			)
			Expect(prodDockerOutput).To(ContainSubstring("docker artifacts created"))

			By("Verifying dev and prod have different configurations")
			devEntrypoint, err := os.ReadFile(filepath.Join(tempDir, "export-dev", "web", "entrypoint.sh"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(devEntrypoint)).To(ContainSubstring("npm run dev"))

			prodEntrypoint, err := os.ReadFile(filepath.Join(tempDir, "export-prod", "web", "entrypoint.sh"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(prodEntrypoint)).To(ContainSubstring("npm run prod"))

			devDockerfile, err := os.ReadFile(filepath.Join(tempDir, "export-dev", "web", "Dockerfile"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(devDockerfile)).To(ContainSubstring("FROM node:18"))

			prodDockerfile, err := os.ReadFile(filepath.Join(tempDir, "export-prod", "web", "Dockerfile"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(prodDockerfile)).To(ContainSubstring("FROM node:18-alpine"))
		})
	})

	Context("Legacy ci.yml Format Support", func() {
		It("should handle legacy ci.yml with path directly in network", func() {
			By("Creating legacy format ci.yml")
			ciYmlPath := filepath.Join(tempDir, "ci.yml")
			err := os.WriteFile(ciYmlPath, []byte(legacyCiYml), 0644)
			Expect(err).NotTo(HaveOccurred())

			By("Generating Docker artifacts")
			dockerOutput := intutil.RunCommand(
				"generate", "docker",
				"--reporoot", tempDir,
				"-b", "ubuntu:latest",
				"-i", "ci.yml",
				"-o", "export",
			)
			fmt.Printf("Legacy Docker generation output: %s\n", dockerOutput)
			Expect(dockerOutput).To(ContainSubstring("docker artifacts created"))

			By("Generating Kubernetes artifacts")
			k8sOutput := intutil.RunCommand(
				"generate", "kubernetes",
				"--reporoot", tempDir,
				"-r", "docker.io/myorg",
				"-i", "ci.yml",
				"-o", "export",
				"-n", "legacy-app",
				"--hostname", "legacy.local",
			)
			fmt.Printf("Legacy Kubernetes generation output: %s\n", k8sOutput)
			Expect(k8sOutput).To(ContainSubstring("Kubernetes artifacts export successful"))

			By("Verifying artifacts were created correctly")
			Expect(filepath.Join(tempDir, "export", "app", "Dockerfile")).To(BeAnExistingFile())
			Expect(filepath.Join(tempDir, "export", "kubernetes", "service-app.yml")).To(BeAnExistingFile())
			Expect(filepath.Join(tempDir, "export", "kubernetes", "ingress.yml")).To(BeAnExistingFile())
		})
	})

	Context("Environment Variables in Docker Artifacts", func() {
		It("should include environment variables in generated artifacts", func() {
			By("Creating ci.yml")
			ciYmlPath := filepath.Join(tempDir, "ci.yml")
			err := os.WriteFile(ciYmlPath, []byte(simpleCiYml), 0644)
			Expect(err).NotTo(HaveOccurred())

			By("Generating Docker artifacts with environment variables")
			output := intutil.RunCommand(
				"generate", "docker",
				"--reporoot", tempDir,
				"-b", "node:18",
				"-i", "ci.yml",
				"-o", "export",
				"-e", "NODE_ENV=production",
				"-e", "API_URL=https://api.example.com",
			)
			fmt.Printf("Docker generation with envs output: %s\n", output)
			Expect(output).To(ContainSubstring("docker artifacts created"))

			By("Verifying docker-compose contains environment variables")
			dockerCompose, err := os.ReadFile(filepath.Join(tempDir, "export", "docker-compose.yml"))
			Expect(err).NotTo(HaveOccurred())
			content := string(dockerCompose)
			Expect(content).To(ContainSubstring("NODE_ENV"))
			Expect(content).To(ContainSubstring("API_URL"))
		})
	})

	Context("Force Overwrite Behavior", func() {
		It("should overwrite existing files when --force is specified", func() {
			By("Creating ci.yml")
			ciYmlPath := filepath.Join(tempDir, "ci.yml")
			err := os.WriteFile(ciYmlPath, []byte(simpleCiYml), 0644)
			Expect(err).NotTo(HaveOccurred())

			By("First generation")
			output := intutil.RunCommand(
				"generate", "docker",
				"--reporoot", tempDir,
				"-b", "ubuntu:latest",
				"-i", "ci.yml",
				"-o", "export",
			)
			Expect(output).To(ContainSubstring("docker artifacts created"))

			By("Second generation with --force")
			output = intutil.RunCommand(
				"generate", "docker",
				"--reporoot", tempDir,
				"-b", "alpine:latest",
				"-i", "ci.yml",
				"-o", "export",
				"--force",
			)
			Expect(output).To(ContainSubstring("docker artifacts created"))

			By("Verifying files were overwritten with new base image")
			dockerfile, err := os.ReadFile(filepath.Join(tempDir, "export", "web", "Dockerfile"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(dockerfile)).To(ContainSubstring("FROM alpine:latest"))
		})
	})

	Context("Generate Command Help", func() {
		It("should display help for generate docker command", func() {
			output := intutil.RunCommand("generate", "docker", "--help")
			fmt.Printf("Generate docker help: %s\n", output)

			Expect(output).To(ContainSubstring("generated artifacts"))
			Expect(output).To(ContainSubstring("-b, --baseimage"))
			Expect(output).To(ContainSubstring("-i, --input"))
			Expect(output).To(ContainSubstring("-o, --output"))
		})

		It("should display help for generate kubernetes command", func() {
			output := intutil.RunCommand("generate", "kubernetes", "--help")
			fmt.Printf("Generate kubernetes help: %s\n", output)

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
			fmt.Printf("Generate images help: %s\n", output)

			Expect(output).To(ContainSubstring("generated images will be pushed"))
			Expect(output).To(ContainSubstring("-r, --registry"))
			Expect(output).To(ContainSubstring("-p, --imagePrefix"))
		})
	})
})
