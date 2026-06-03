// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"errors"

	"github.com/codesphere-cloud/cs-go/cli/cmd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RemoveTeamMember", func() {
	var (
		mockEnv    *cmd.MockEnv
		mockClient *cmd.MockClient
		c          *cmd.RemoveTeamMemberCmd
		teamId     int
		userId     int
	)

	BeforeEach(func() {
		mockClient = cmd.NewMockClient(GinkgoT())
		mockEnv = cmd.NewMockEnv(GinkgoT())
		teamId = 42
		userId = 100
		c = &cmd.RemoveTeamMemberCmd{
			Opts: cmd.RemoveTeamMemberOpts{
				GlobalOptions: &cmd.GlobalOptions{
					Env:    mockEnv,
					TeamId: teamId,
				},
				UserId: userId,
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
		It("should fail if the user ID is empty", func() {
			err := c.RemoveTeamMember(mockClient, teamId, 0)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("user ID cannot be empty"))
		})
	})

	Context("RunE execution flow", func() {
		It("should successfully remove a member from a team", func() {
			mockClient.EXPECT().RemoveTeamMember(teamId, userId).Return(nil).Once()

			err := c.RunE(nil, []string{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail when the token is not allowed to remove a member", func() {
			mockClient.EXPECT().RemoveTeamMember(teamId, userId).Return(errors.New("failed")).Once()

			err := c.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to remove member from team: "))
		})

		It("should fail when client creation fails", func() {
			c.ClientFactory = func(opts cmd.GlobalOptions) (cmd.Client, error) {
				return nil, errors.New("client init failed")
			}

			err := c.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to create Codespehre client: client init failed"))
		})

		It("should fail when team ID is unavailable", func() {
			c.Opts.TeamId = -1
			mockEnv.EXPECT().GetTeamId().Return(-1, errors.New("CS_TEAM_ID env var required, but not set")).Once()

			err := c.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("CS_TEAM_ID env var required, but not set"))
		})

		It("should fail when user ID is empty", func() {
			c.Opts.UserId = 0
			err := c.RunE(nil, []string{})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("user ID cannot be empty"))
		})
	})
})
