// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
)

var _ = Describe("WorkspaceURL", func() {
	Context("ConstructWorkspaceServiceURL", func() {
		It("should construct URL correctly with all parameters", func() {
			devDomain := "team-slug.codesphere.com"
			workspace := api.Workspace{
				Id:        1234,
				DevDomain: &devDomain,
			}

			url, err := cmd.ConstructWorkspaceServiceURL(workspace, 3000, "/api/health")

			Expect(err).NotTo(HaveOccurred())
			Expect(url).To(Equal("https://1234-3000.team-slug.codesphere.com/api/health"))
		})

		It("should construct URL correctly with root path", func() {
			devDomain := "team-slug.codesphere.com"
			workspace := api.Workspace{
				Id:        5678,
				DevDomain: &devDomain,
			}

			url, err := cmd.ConstructWorkspaceServiceURL(workspace, 8080, "/")

			Expect(err).NotTo(HaveOccurred())
			Expect(url).To(Equal("https://5678-8080.team-slug.codesphere.com/"))
		})

		It("should construct URL correctly with empty path", func() {
			devDomain := "dev.example.com"
			workspace := api.Workspace{
				Id:        999,
				DevDomain: &devDomain,
			}

			url, err := cmd.ConstructWorkspaceServiceURL(workspace, 3001, "")

			Expect(err).NotTo(HaveOccurred())
			Expect(url).To(Equal("https://999-3001.dev.example.com"))
		})

		It("should return error when dev domain is nil", func() {
			workspace := api.Workspace{
				Id:        1234,
				DevDomain: nil,
			}

			_, err := cmd.ConstructWorkspaceServiceURL(workspace, 3000, "/")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("does not have a development domain configured"))
		})
	})
})
