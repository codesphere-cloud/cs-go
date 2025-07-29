// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package docker_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/codesphere-cloud/cs-go/pkg/ci"
	"github.com/codesphere-cloud/cs-go/tmpl/docker"
)

var _ = Describe("CreateDockerfile", func() {
	var (
		dockerConfig docker.DockerTemplateConfig
	)

	Context("The baseimage is not provided", func() {
		JustBeforeEach(func() {
			dockerConfig = docker.DockerTemplateConfig{
				PrepareSteps: []ci.Step{
					{
						Name:    "Install dependencies",
						Command: "npm install",
					},
					{
						Name:    "Build project",
						Command: "npm run build",
					},
				},
			}
		})
		It("should return an error", func() {
			_, err := docker.CreateDockerfile(dockerConfig)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("base image is required"))
		})
	})

	Context("All values are provided", func() {
		JustBeforeEach(func() {
			dockerConfig = docker.DockerTemplateConfig{
				BaseImage: "node:20",
				PrepareSteps: []ci.Step{
					{
						Name:    "Install dependencies",
						Command: "npm install",
					},
					{
						Name:    "Build project",
						Command: "npm run build",
					},
				},
			}
		})
		It("Creates a Dockerfile with the correct base image and prepare steps", func() {
			dockerfile, err := docker.CreateDockerfile(dockerConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(dockerfile)).To(ContainSubstring("FROM node:20"))
			Expect(string(dockerfile)).To(ContainSubstring("RUN npm install"))
			Expect(string(dockerfile)).To(ContainSubstring("RUN npm run build"))
		})
	})
})
