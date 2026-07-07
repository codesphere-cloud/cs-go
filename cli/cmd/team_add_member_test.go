// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"errors"

	"github.com/codesphere-cloud/cs-go/cli/cmd"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AddTeamMember", func() {
	var (
		mockEnv    *cmd.MockEnv
		mockClient *cmd.MockClient
		c          *cmd.AddTeamMemberCmd
		teamId     int
		role       cs.TeamRole
		email      string
	)

	BeforeEach(func() {
		mockClient = cmd.NewMockClient(GinkgoT())
		mockEnv = cmd.NewMockEnv(GinkgoT())
		teamId = 42
		email = "test@test.com"
		role = cs.RoleMember
		c = &cmd.AddTeamMemberCmd{
			Opts: cmd.AddTeamMemberOpts{
				GlobalOptions: &cmd.GlobalOptions{
					Env:    mockEnv,
					TeamId: teamId,
					// OrgId is intentionally left empty here, will be set in BeforeEach for specific contexts
				},
				Email:  email,
				TeamId: teamId,
				Role:   -1,
			},
			ClientFactory: func(opts cmd.GlobalOptions) (cmd.Client, error) {
				return mockClient, nil
			},
		}
		// Mock common environment calls needed for client creation
	})

	AfterEach(func() {
		mockEnv.AssertExpectations(GinkgoT())
		mockClient.AssertExpectations(GinkgoT())
	})

	Context("Validation", func() {
		It("should fail if the mail is empty", func() {

			err := c.AddTeamMember(mockClient, teamId, "", role)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("email cannot be empty"))
		})

		It("should fail if the email is invalid", func() {

			err := c.AddTeamMember(mockClient, teamId, "invalid-email", role)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid email address"))
		})

		It("should fail if the role is invalid (e.g. 3)", func() {
			err := c.AddTeamMember(mockClient, teamId, "user@example.com", 3)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid role: must be 1 for member or -1 for admin"))
		})

		It("should fail if the role is 0", func() {
			err := c.AddTeamMember(mockClient, teamId, "user@example.com", 0)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid role: must be 1 for member or -1 for admin"))
		})
	})

	Context("RunE execution flow", func() {
		It("should successfully add a member to a team with role -1 (admin)", func() {
			c.Opts.Role = -1
			mockClient.EXPECT().AddTeamMember(teamId, email, -1).Return(nil).Once()

			err := c.RunE(nil, []string{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("should successfully add a member to a team with role 1 (member)", func() {
			c.Opts.Role = 1
			mockClient.EXPECT().AddTeamMember(teamId, email, 1).Return(nil).Once()

			err := c.RunE(nil, []string{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail when the token is not allowed to add a member", func() {
			mockClient.EXPECT().AddTeamMember(teamId, email, -1).Return(errors.New("failed")).Once()

			err := c.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to add member to team: "))
		})

		It("should fail when client creation fails", func() {
			c.ClientFactory = func(opts cmd.GlobalOptions) (cmd.Client, error) {
				return nil, errors.New("client init failed")
			}

			err := c.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to create Codesphere client: client init failed"))
		})

		It("should fail when team ID is unavailable", func() {
			c.Opts.GlobalOptions.TeamId = -1
			mockEnv.EXPECT().GetTeamId().Return(-1, errors.New("CS_TEAM_ID env var required, but not set")).Once()

			err := c.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("CS_TEAM_ID env var required, but not set"))
		})

		It("should fail when email is empty", func() {
			c.Opts.Email = ""
			err := c.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("email cannot be empty"))
		})

		It("should fail when email is invalid", func() {
			c.Opts.Email = "invalid-email"
			err := c.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid email address"))
		})

		It("should fail when role is invalid (e.g. 3)", func() {
			c.Opts.Role = 3
			err := c.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid role: must be 1 for member or -1 for admin"))
		})

		It("should fail when role is 0", func() {
			c.Opts.Role = 0
			err := c.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid role: must be 1 for member or -1 for admin"))
		})
	})

})
