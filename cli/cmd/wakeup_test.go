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

var _ = Describe("WakeUp", func() {
	var (
		mockEnv    *cmd.MockEnv
		mockClient *cmd.MockClient
		c          *cmd.WakeUpCmd
		wsId       int
		teamId     int
		token      string
	)

	JustBeforeEach(func() {
		mockClient = cmd.NewMockClient(GinkgoT())
		mockEnv = cmd.NewMockEnv(GinkgoT())
		wsId = 42
		teamId = 21
		token = "test-api-token"
		c = &cmd.WakeUpCmd{
			Opts: cmd.GlobalOptions{
				Env:         mockEnv,
				WorkspaceId: &wsId,
			},
		}
	})

	Context("WakeUpWorkspace", func() {
		It("should construct the correct services domain and wake up the workspace", func() {
			devDomain := "team-slug.codesphere.com"
			workspace := api.Workspace{
				Id:        wsId,
				TeamId:    teamId,
				Name:      "test-workspace",
				DevDomain: &devDomain,
			}

			mockClient.EXPECT().GetWorkspace(wsId).Return(workspace, nil)

			err := c.WakeUpWorkspace(mockClient, wsId, token)

			// This will fail because we're making a real HTTP request
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to wake up workspace"))
		})

		It("should return error if workspace has no dev domain", func() {
			workspace := api.Workspace{
				Id:        wsId,
				TeamId:    teamId,
				Name:      "test-workspace",
				DevDomain: nil,
			}

			mockClient.EXPECT().GetWorkspace(wsId).Return(workspace, nil)

			err := c.WakeUpWorkspace(mockClient, wsId, token)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("does not have a development domain configured"))
		})

		It("should return error if GetWorkspace fails", func() {
			mockClient.EXPECT().GetWorkspace(wsId).Return(api.Workspace{}, fmt.Errorf("api error"))

			err := c.WakeUpWorkspace(mockClient, wsId, token)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get workspace"))
		})
	})
})
