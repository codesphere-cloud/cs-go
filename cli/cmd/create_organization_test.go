// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreateOrganization", func() {
	var (
		mockEnv    *cmd.MockEnv
		mockClient *cmd.MockClient
		c          *cmd.CreateOrganizationCmd
		orgId      string
		orgName    string
		adminEmail string
	)

	BeforeEach(func() {
		mockClient = cmd.NewMockClient(GinkgoT())
		mockEnv = cmd.NewMockEnv(GinkgoT())
		orgId = "test-org-id"
		orgName = "test-org"
		adminEmail = "admin@example.com"
		c = &cmd.CreateOrganizationCmd{
			Opts: cmd.CreateOrganizationOpts{
				GlobalOptions: &cmd.GlobalOptions{
					Env: mockEnv,
				},
				Name:       orgName,
				AdminEmail: adminEmail,
			},
			ClientFactory: func(opts cmd.GlobalOptions) (cmd.Client, error) {
				return mockClient, nil
			},
		}
	})

	AfterEach(func() {
		mockEnv.AssertExpectations(GinkgoT())
		mockClient.AssertExpectations(GinkgoT())
	})

	Context("Validation", func() {
		It("should fail if the organization name is empty", func() {
			org, err := c.CreateOrganization(mockClient, "", adminEmail)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("organization name cannot be empty"))
			Expect(org).To(BeNil())
		})

		It("should fail if the admin email is empty", func() {
			org, err := c.CreateOrganization(mockClient, orgName, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("admin email cannot be empty"))
			Expect(org).To(BeNil())
		})
	})

	Context("RunE execution flow", func() {
		It("should successfully create an organization and print the correct message", func() {
			expectedOrg := api.Organization{
				Id: orgId,
			}
			mockClient.EXPECT().CreateOrganization(orgName, adminEmail).Return(&expectedOrg, nil).Once()

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := c.RunE(nil, []string{})
			Expect(err).ToNot(HaveOccurred())

			_ = w.Close()
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			os.Stdout = oldStdout

			Expect(buf.String()).To(ContainSubstring(fmt.Sprintf("Organization created: %v\n", orgId)))
		})

		It("should return error if API call fails", func() {
			mockClient.EXPECT().CreateOrganization(orgName, adminEmail).Return(nil, fmt.Errorf("api error")).Once()

			err := c.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("api error"))
		})
	})
})
