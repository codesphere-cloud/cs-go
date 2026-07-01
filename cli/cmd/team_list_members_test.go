// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"errors"
	"time"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ListTeamMembers", func() {
	var (
		mockEnv    *cmd.MockEnv
		mockClient *cmd.MockClient
		l          *cmd.ListTeamMembersCmd
		teamId     int
	)

	BeforeEach(func() {
		mockClient = cmd.NewMockClient(GinkgoT())
		mockEnv = cmd.NewMockEnv(GinkgoT())
		teamId = 42
		l = &cmd.ListTeamMembersCmd{
			Opts: cmd.ListTeamMembersOpts{
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

	Context("ListTeamMembers", func() {
		It("should successfully list team members", func() {
			name := "Test User"
			email := "test@example.com"
			expectedMembers := []api.TeamMember{
				{
					UserId:    1,
					TeamId:    teamId,
					Role:      0,
					Pending:   false,
					CreatedAt: time.Now(),
					Name:      &name,
					Email:     &email,
				},
			}
			mockClient.EXPECT().ListTeamMembers(teamId).Return(expectedMembers, nil).Once()

			err := l.ListTeamMembers(mockClient, teamId)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return error when API call fails", func() {
			mockClient.EXPECT().ListTeamMembers(teamId).Return(nil, errors.New("api error")).Once()

			err := l.ListTeamMembers(mockClient, teamId)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to list team members: api error"))
		})
	})

	Context("RunE execution flow", func() {
		It("should successfully run the command", func() {
			mockClient.EXPECT().ListTeamMembers(teamId).Return([]api.TeamMember{}, nil).Once()

			err := l.RunE(nil, []string{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail when client creation fails", func() {
			l.ClientFactory = func(opts cmd.GlobalOptions) (cmd.Client, error) {
				return nil, errors.New("client init failed")
			}

			err := l.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to create Codesphere client: client init failed"))
		})

		It("should fail when team ID is unavailable", func() {
			l.Opts.TeamId = -1
			mockEnv.EXPECT().GetTeamId().Return(-1, errors.New("CS_TEAM_ID env var required, but not set")).Once()

			err := l.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("CS_TEAM_ID env var required, but not set"))
		})
	})
})
