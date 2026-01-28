// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
)

var _ = Describe("Curl", func() {
	var (
		mockEnv    *cmd.MockEnv
		mockClient *cmd.MockClient
		c          *cmd.CurlCmd
		wsId       int
		teamId     int
		token      string
		port       int
	)

	JustBeforeEach(func() {
		mockClient = cmd.NewMockClient(GinkgoT())
		mockEnv = cmd.NewMockEnv(GinkgoT())
		wsId = 42
		teamId = 21
		token = "test-api-token"
		port = 3000
		c = &cmd.CurlCmd{
			Opts: cmd.GlobalOptions{
				Env:         mockEnv,
				WorkspaceId: &wsId,
			},
			Port: &port,
		}
	})

	Context("CurlWorkspace", func() {
		It("should construct the correct URL with default port", func() {
			workspace := api.Workspace{
				Id:     wsId,
				TeamId: teamId,
				Name:   "test-workspace",
			}
			team := api.Team{
				Id:                  teamId,
				DefaultDataCenterId: 5,
				Name:                "test-team",
			}

			mockClient.EXPECT().GetWorkspace(wsId).Return(workspace, nil)
			mockClient.EXPECT().GetTeam(teamId).Return(&team, nil)

			err := c.CurlWorkspace(mockClient, wsId, token, "/api/health", []string{"-I"})

			Expect(err).To(HaveOccurred())
		})

		It("should construct the correct URL with custom port", func() {
			customPort := 3001
			c.Port = &customPort
			workspace := api.Workspace{
				Id:     wsId,
				TeamId: teamId,
				Name:   "test-workspace",
			}
			team := api.Team{
				Id:                  teamId,
				DefaultDataCenterId: 5,
				Name:                "test-team",
			}

			mockClient.EXPECT().GetWorkspace(wsId).Return(workspace, nil)
			mockClient.EXPECT().GetTeam(teamId).Return(&team, nil)

			err := c.CurlWorkspace(mockClient, wsId, token, "/custom/path", []string{})

			Expect(err).To(HaveOccurred())
		})

		It("should return error if GetTeam fails", func() {
			workspace := api.Workspace{
				Id:     wsId,
				TeamId: teamId,
				Name:   "test-workspace",
			}

			mockClient.EXPECT().GetWorkspace(wsId).Return(workspace, nil)
			mockClient.EXPECT().GetTeam(teamId).Return(nil, fmt.Errorf("team not found"))

			err := c.CurlWorkspace(mockClient, wsId, token, "/", []string{})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get team"))
		})

		It("should return error if GetWorkspace fails", func() {
			mockClient.EXPECT().GetWorkspace(wsId).Return(api.Workspace{}, fmt.Errorf("api error"))

			err := c.CurlWorkspace(mockClient, wsId, token, "/", []string{})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get workspace"))
		})
	})
})
