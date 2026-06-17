// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"errors"

	"github.com/codesphere-cloud/cs-go/cli/cmd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RemoveTeam", func() {
	var (
		mockEnv    *cmd.MockEnv
		mockClient *cmd.MockClient
		c          *cmd.RemoveTeamCmd
		teamId     int
	)

	BeforeEach(func() {
		mockClient = cmd.NewMockClient(GinkgoT())
		mockEnv = cmd.NewMockEnv(GinkgoT())
		teamId = 42
		c = &cmd.RemoveTeamCmd{
			Opts: cmd.RemoveTeamOpts{
				GlobalOptions: &cmd.GlobalOptions{
					Env:    mockEnv,
					TeamId: teamId,
				},
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

	Context("RunE execution flow", func() {
		It("should successfully remove a team", func() {
			mockEnv.EXPECT().GetOrgId().Return("").Once()
			mockClient.EXPECT().DeleteTeam("", teamId).Return(nil).Once()

			err := c.RunE(nil, []string{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail when the token is not allowed to remove a team", func() {
			mockEnv.EXPECT().GetOrgId().Return("").Once()
			mockClient.EXPECT().DeleteTeam("", teamId).Return(errors.New("failed")).Once()

			err := c.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to delete team: "))
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
			mockEnv.EXPECT().GetOrgId().Return("").Once()
			mockEnv.EXPECT().GetTeamId().Return(-1, errors.New("CS_TEAM_ID env var required, but not set")).Once()

			err := c.RunE(nil, []string{})
			Expect(err).To(MatchError("team ID not set, use -t or CS_TEAM_ID to set it"))
		})
	})
})
