// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package exporter_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	mock "github.com/stretchr/testify/mock"

	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/codesphere-cloud/cs-go/pkg/exporter"
)

const ymlContent = `
schemaVersion: v0.2
prepare:
  steps:
test:
  steps: []
run:
  frontend:
    steps:
      - run: echo "Hello World"
    plan: 21
    replicas: 1
    isPublic: true
    network:
      path: /
      stripPath: true
`

var _ = Describe("GenerateDockerfile", func() {
	var (
		memoryFs         *cs.FileSystem
		e                exporter.Exporter
		defaultInput     string
		defaultOutput    string
		defaultBaseImage string
	)

	BeforeEach(func() {
		defaultInput = "ci.yml"
		defaultOutput = "./export"
		defaultBaseImage = "alpine:latest"
		memoryFs = cs.NewMemFileSystem()
		e = exporter.NewExporterService(memoryFs, defaultOutput, defaultBaseImage, []string{}, "workspace-repo", false)
	})

	Context("The exporter is not set up", func() {
		It("should return an error", func() {
			err := e.ExportDockerArtifacts()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("call ReadYmlFile first"))

			err = e.ExportKubernetesArtifacts("", "", "", "", "", "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("call ReadYmlFile first"))
		})
	})

	Context("The exporter is set up", func() {
		Context("Valid ci file", func() {
			JustBeforeEach(func() {
				err := memoryFs.WriteFile(".", defaultInput, []byte(ymlContent), false)
				Expect(err).To(Not(HaveOccurred()))
			})
			It("should generate files and don't return an error", func() {
				_, err := e.ReadYmlFile(defaultInput)
				Expect(err).To(Not(HaveOccurred()))
				err = e.ExportDockerArtifacts()
				Expect(err).To(Not(HaveOccurred()))

				Expect(memoryFs.DirExists("workspace-repo/export")).To(BeTrue())
				Expect(memoryFs.FileExists("workspace-repo/export/docker-compose.yml")).To(BeTrue())

				Expect(memoryFs.DirExists("workspace-repo/export/frontend")).To(BeTrue())
				Expect(memoryFs.FileExists("workspace-repo/export/frontend/Dockerfile")).To(BeTrue())
				Expect(memoryFs.FileExists("workspace-repo/export/frontend/entrypoint.sh")).To(BeTrue())

				err = e.ExportKubernetesArtifacts("registry", "image", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				Expect(err).To(Not(HaveOccurred()))

				Expect(memoryFs.DirExists("workspace-repo/export/kubernetes")).To(BeTrue())
				Expect(memoryFs.FileExists("workspace-repo/export/kubernetes/ingress.yml")).To(BeTrue())

				Expect(memoryFs.FileExists("workspace-repo/export/kubernetes/service-frontend.yml")).To(BeTrue())
			})
		})
	})
})
