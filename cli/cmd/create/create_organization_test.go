// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package create_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
	createcmd "github.com/codesphere-cloud/cs-go/cli/cmd/create"
)

var _ = Describe("CreateOrganization", func() {
	var (
		mockEnv    *cmd.MockEnv
		mockClient *cmd.MockClient
		c          *createcmd.CreateOrganizationCmd
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
		c = &createcmd.CreateOrganizationCmd{
			Opts: createcmd.CreateOrganizationOpts{
				RootOptions: &cmd.GlobalOptions{
					Env: mockEnv,
				},
				Name:       orgName,
				AdminEmail: adminEmail,
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

		It("should fail if the admin email is invalid", func() {
			org, err := c.CreateOrganization(mockClient, orgName, "not-an-email")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("admin email is invalid"))
			Expect(org).To(BeNil())
		})
	})

	Context("CreateOrganization execution flow", func() {
		It("should successfully create an organization", func() {
			expectedOrg := api.Organization{
				Id: orgId,
			}
			mockClient.EXPECT().CreateOrganization(orgName, adminEmail).Return(&expectedOrg, nil).Once()

			org, err := c.CreateOrganization(mockClient, orgName, adminEmail)
			Expect(err).ToNot(HaveOccurred())
			Expect(org.Id).To(Equal(orgId))
		})

		It("should return error if API call fails", func() {
			mockClient.EXPECT().CreateOrganization(orgName, adminEmail).Return(nil, fmt.Errorf("api error")).Once()

			org, err := c.CreateOrganization(mockClient, orgName, adminEmail)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to create organization: api error"))
			Expect(org).To(BeNil())
		})
	})
})
