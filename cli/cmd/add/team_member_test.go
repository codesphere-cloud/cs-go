// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package add_test

import (
	"errors"

	"github.com/codesphere-cloud/cs-go/cli/cmd"
	addcmd "github.com/codesphere-cloud/cs-go/cli/cmd/add"
	shared "github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AddTeamMember", func() {
	var (
		mockEnv    *cmd.MockEnv
		mockClient *cmd.MockClient
		globalOpts *cmd.GlobalOptions
		c          *addcmd.AddTeamMemberCmd
		teamId     int
		role       string
		email      string
	)

	BeforeEach(func() {
		mockClient = cmd.NewMockClient(GinkgoT())
		mockEnv = cmd.NewMockEnv(GinkgoT())
		teamId = 42
		email = "test@test.com"
		role = "member"
		globalOpts = &cmd.GlobalOptions{
			Env:    mockEnv,
			TeamId: teamId,
		}
		c = &addcmd.AddTeamMemberCmd{
			Opts: addcmd.AddTeamMemberOpts{
				RootOptions: globalOpts,
				Email:       email,
				TeamId:      teamId,
				Role:        "admin",
			},
			ClientFactory: func(opts shared.RootOptions) (cmd.Client, error) {
				return mockClient, nil
			},
		}
		// Mock common environment calls needed for client creation
	})

	AfterEach(func() {
		mockEnv.AssertExpectations(GinkgoT())
		mockClient.AssertExpectations(GinkgoT())
	})

	Context("Role flag", func() {
		It("defaults to member", func() {
			leaf, _, err := cmd.GetRootCmd().Find([]string{"add", "team-member"})
			Expect(err).NotTo(HaveOccurred())
			role, err := leaf.Flags().GetString("role")
			Expect(err).NotTo(HaveOccurred())
			Expect(role).To(Equal("member"))
			c.Opts.Role = role
			mockClient.EXPECT().AddTeamMember(teamId, email, 1).Return(nil).Once()
			Expect(c.RunE(nil, nil)).To(Succeed())
		})

		It("accepts admin through the short flag", func() {
			leaf, _, err := cmd.GetRootCmd().Find([]string{"add", "team-member"})
			Expect(err).NotTo(HaveOccurred())
			Expect(leaf.ParseFlags([]string{"-r", "admin"})).To(Succeed())
			c.Opts.Role, err = leaf.Flags().GetString("role")
			Expect(err).NotTo(HaveOccurred())
			mockClient.EXPECT().AddTeamMember(teamId, email, 0).Return(nil).Once()
			Expect(c.RunE(nil, nil)).To(Succeed())
		})

		It("accepts member through the long flag", func() {
			leaf, _, err := cmd.GetRootCmd().Find([]string{"add", "team-member"})
			Expect(err).NotTo(HaveOccurred())
			Expect(leaf.ParseFlags([]string{"--role", "member"})).To(Succeed())
			c.Opts.Role, err = leaf.Flags().GetString("role")
			Expect(err).NotTo(HaveOccurred())
			mockClient.EXPECT().AddTeamMember(teamId, email, 1).Return(nil).Once()
			Expect(c.RunE(nil, nil)).To(Succeed())
		})

		It("requires a value after the role flag", func() {
			leaf, _, err := cmd.GetRootCmd().Find([]string{"add", "team-member"})
			Expect(err).NotTo(HaveOccurred())
			Expect(leaf.ParseFlags([]string{"-r"})).To(MatchError(ContainSubstring("flag needs an argument")))
		})
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

		It("should fail if the role is unknown", func() {
			err := c.AddTeamMember(mockClient, teamId, "user@example.com", "owner")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid role: must be member or admin"))
		})

		It("should fail if the role is empty", func() {
			err := c.AddTeamMember(mockClient, teamId, "user@example.com", "")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid role: must be member or admin"))
		})
	})

	It("rejects the old numeric member role", func() {
		Expect(c.AddTeamMember(mockClient, teamId, email, "1")).To(MatchError("invalid role: must be member or admin"))
	})

	It("rejects the old numeric admin role", func() {
		Expect(c.AddTeamMember(mockClient, teamId, email, "-1")).To(MatchError("invalid role: must be member or admin"))
	})

	Context("RunE execution flow", func() {
		It("should successfully add a member to a team with role admin", func() {
			c.Opts.Role = "admin"
			mockClient.EXPECT().AddTeamMember(teamId, email, 0).Return(nil).Once()

			err := c.RunE(nil, []string{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("should successfully add a member to a team with role member", func() {
			c.Opts.Role = "member"
			mockClient.EXPECT().AddTeamMember(teamId, email, 1).Return(nil).Once()

			err := c.RunE(nil, []string{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail when the token is not allowed to add a member", func() {
			mockClient.EXPECT().AddTeamMember(teamId, email, 0).Return(errors.New("failed")).Once()

			err := c.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to add member to team: "))
		})

		It("should fail when client creation fails", func() {
			c.ClientFactory = func(opts shared.RootOptions) (cmd.Client, error) {
				return nil, errors.New("client init failed")
			}

			err := c.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to create Codesphere client: client init failed"))
		})

		It("should fail when team ID is unavailable", func() {
			globalOpts.TeamId = -1
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

		It("should fail when role is unknown", func() {
			c.Opts.Role = "owner"
			err := c.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid role: must be member or admin"))
		})

		It("should fail when role is empty", func() {
			c.Opts.Role = ""
			err := c.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid role: must be member or admin"))
		})
	})

})
