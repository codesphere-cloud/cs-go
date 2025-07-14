// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package export_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/codesphere-cloud/cs-go/pkg/ci"
	"github.com/codesphere-cloud/cs-go/tmpl/export"
)

var _ = Describe("CreateDockerCompose", func() {
	var (
		dockerComposeConfig export.DockerComposeTemplateConfig
	)

	Context("No services are provided", func() {
		JustBeforeEach(func() {
			dockerComposeConfig = export.DockerComposeTemplateConfig{
				Services: map[string]ci.Service{},
				EnvVars:  []string{"NODE_ENV=production"},
			}
		})
		It("should return an error", func() {
			_, err := export.CreateDockerCompose(dockerComposeConfig)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("at least one service is required"))
		})
	})

	Context("Empty service name is provided", func() {
		JustBeforeEach(func() {
			dockerComposeConfig = export.DockerComposeTemplateConfig{
				Services: map[string]ci.Service{
					"": {},
				},
				EnvVars: []string{"NODE_ENV=production"},
			}
		})
		It("should return an error", func() {
			_, err := export.CreateDockerCompose(dockerComposeConfig)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("service name cannot be empty"))
		})
	})

	Context("All values are provided", func() {
		JustBeforeEach(func() {
			dockerComposeConfig = export.DockerComposeTemplateConfig{
				// We test only with empty services as only the key is used in the template
				Services: map[string]ci.Service{
					"web": {},
				},
				EnvVars: []string{"NODE_ENV=production"},
			}
		})
		It("Creates a Docker Compose file with the correct services and environment variables", func() {
			dockerCompose, err := export.CreateDockerCompose(dockerComposeConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(dockerCompose)).To(ContainSubstring("services:"))
			Expect(string(dockerCompose)).To(ContainSubstring("web:"))
			Expect(string(dockerCompose)).To(ContainSubstring("context: ./web"))
			Expect(string(dockerCompose)).To(ContainSubstring("environment:"))
			Expect(string(dockerCompose)).To(ContainSubstring("- NODE_ENV=production"))
			// Nginx service depends on web service
			Expect(string(dockerCompose)).To(ContainSubstring("- web"))
		})
	})
})
