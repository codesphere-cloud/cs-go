// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package delete_test

import (
	"context"
	"errors"

	"github.com/codesphere-cloud/cs-go/cli/cmd"
	deletecmd "github.com/codesphere-cloud/cs-go/cli/cmd/delete"
	shared "github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("DeleteTeam", func() {
	var (
		mockEnv    *cmd.MockEnv
		mockClient *cmd.MockClient
		globalOpts *cmd.GlobalOptions
		c          *deletecmd.DeleteTeamCmd
		teamId     int
	)

	BeforeEach(func() {
		mockClient = cmd.NewMockClient(GinkgoT())
		mockEnv = cmd.NewMockEnv(GinkgoT())
		teamId = 42
		globalOpts = &cmd.GlobalOptions{
			Env:    mockEnv,
			TeamId: teamId,
		}
		c = &deletecmd.DeleteTeamCmd{
			Opts: deletecmd.DeleteTeamOpts{
				RootOptions: globalOpts,
			},
			ClientFactory: func(opts shared.RootOptions) (cmd.Client, error) {
				return mockClient, nil
			},
		}
	})

	AfterEach(func() {
		mockEnv.AssertExpectations(GinkgoT())
		mockClient.AssertExpectations(GinkgoT())
	})

	Context("RunE execution flow", func() {
		It("should successfully delete a team", func() {
			mockClient.EXPECT().DeleteTeam(mock.Anything, teamId).Return(nil).Once()

			err := c.RunE(newTestCommand(), []string{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("passes the command context to the API call", func() {
			type ctxKey struct{}
			ctx := context.WithValue(context.Background(), ctxKey{}, "delete-team")
			command := &cobra.Command{}
			command.SetContext(ctx)
			mockClient.EXPECT().DeleteTeam(ctx, teamId).Return(nil).Once()

			Expect(c.RunE(command, []string{})).To(Succeed())
		})

		It("should fail when the token is not allowed to delete a team", func() {
			mockClient.EXPECT().DeleteTeam(mock.Anything, teamId).Return(errors.New("failed")).Once()

			err := c.RunE(newTestCommand(), []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to delete team: "))
		})

		It("should fail when client creation fails", func() {
			c.ClientFactory = func(opts shared.RootOptions) (cmd.Client, error) {
				return nil, errors.New("client init failed")
			}

			err := c.RunE(newTestCommand(), []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to create Codesphere client: client init failed"))
		})

		It("should fail when team ID is unavailable", func() {
			globalOpts.TeamId = -1
			mockEnv.EXPECT().GetTeamId().Return(-1, errors.New("CS_TEAM_ID env var required, but not set")).Once()

			err := c.RunE(newTestCommand(), []string{})
			Expect(err).To(MatchError("resource group ID not set, use --resource-group or CS_RESOURCE_GROUP_ID to set it"))
		})
	})
})
