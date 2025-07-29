// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package docker_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/codesphere-cloud/cs-go/pkg/ci"
	"github.com/codesphere-cloud/cs-go/tmpl/docker"
)

var _ = Describe("CreateNginxConfig", func() {
	var (
		nginxConfig docker.NginxConfigTemplateConfig
	)

	BeforeEach(func() {
		nginxConfig = docker.NginxConfigTemplateConfig{
			Services: map[string]ci.Service{},
		}
	})

	Context("No services are provided", func() {
		It("should return an error", func() {
			_, err := docker.CreateNginxConfig(nginxConfig)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("at least one service is required"))
		})
	})

	Context("Empty service name is provided", func() {
		JustBeforeEach(func() {
			nginxConfig.Services = map[string]ci.Service{
				"": {},
			}
		})
		It("should return an error", func() {
			_, err := docker.CreateNginxConfig(nginxConfig)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("service name cannot be empty"))
		})
	})

	Context("Empty path is provided", func() {
		JustBeforeEach(func() {
			nginxConfig.Services = map[string]ci.Service{
				"web": {
					IsPublic: true,
					Network: ci.Network{
						Paths: []ci.Path{{
							Port:      3000,
							Path:      "",
							StripPath: true,
						}},
					},
				},
			}
		})
		It("should return an error", func() {
			_, err := docker.CreateNginxConfig(nginxConfig)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("path cannot be empty"))
		})
	})

	Context("Public path with length 1 is provided", func() {
		JustBeforeEach(func() {
			nginxConfig.Services = map[string]ci.Service{
				"web": {
					IsPublic: true,
					Network: ci.Network{
						Paths: []ci.Path{{
							Port:      3000,
							Path:      "/",
							StripPath: true,
						}},
					},
				},
			}
		})
		It("Creates an Nginx config with the correct services", func() {
			config, err := docker.CreateNginxConfig(nginxConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(config)).To(ContainSubstring("location / {"))
			Expect(string(config)).To(ContainSubstring("proxy_pass http://web:3000/;"))
		})
	})

	Context("Public path with length greater than 1 is provided", func() {
		JustBeforeEach(func() {
			nginxConfig.Services = map[string]ci.Service{
				"web": {
					IsPublic: true,
					Network: ci.Network{
						Paths: []ci.Path{{
							Port:      3000,
							Path:      "/api",
							StripPath: true,
						}},
					},
				},
			}
		})
		It("Creates an Nginx config with the correct services", func() {
			config, err := docker.CreateNginxConfig(nginxConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(config)).To(ContainSubstring("location /api/ {"))
			Expect(string(config)).To(ContainSubstring("proxy_pass http://web:3000/;"))
		})
	})

	Context("Private path is provided", func() {
		JustBeforeEach(func() {
			nginxConfig.Services = map[string]ci.Service{
				"web": {
					IsPublic: false,
					Network: ci.Network{
						Paths: []ci.Path{{
							Port:      3000,
							Path:      "/private",
							StripPath: true,
						}},
					},
				},
			}
		})
		It("Creates an Nginx config without the private service", func() {
			config, err := docker.CreateNginxConfig(nginxConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(config)).ToNot(ContainSubstring("location /private {"))
			Expect(string(config)).ToNot(ContainSubstring("proxy_pass http://web:3000/;"))
		})
	})
})
