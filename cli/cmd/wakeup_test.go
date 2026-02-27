// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

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
	)

	JustBeforeEach(func() {
		mockClient = cmd.NewMockClient(GinkgoT())
		mockEnv = cmd.NewMockEnv(GinkgoT())
		wsId = 42
		teamId = 21
		c = &cmd.WakeUpCmd{
			Opts: cmd.WakeUpOptions{
				GlobalOptions: &cmd.GlobalOptions{
					Env:         mockEnv,
					WorkspaceId: wsId,
				},
				Timeout: 120 * time.Second,
			},
		}
	})

	Context("WakeUpWorkspace", func() {
		It("should wake up the workspace by scaling to 1 replica", func() {
			workspace := api.Workspace{
				Id:     wsId,
				TeamId: teamId,
				Name:   "test-workspace",
			}

			mockClient.EXPECT().GetWorkspace(wsId).Return(workspace, nil)
			mockClient.EXPECT().WorkspaceStatus(wsId).Return(&api.WorkspaceStatus{IsRunning: false}, nil)
			mockClient.EXPECT().ScaleWorkspace(wsId, 1).Return(nil)
			mockClient.EXPECT().WaitForWorkspaceRunning(mock.Anything, mock.Anything).Return(nil)

			err := c.WakeUpWorkspace(mockClient, wsId)

			Expect(err).ToNot(HaveOccurred())
		})

		It("should return early if workspace is already running", func() {
			workspace := api.Workspace{
				Id:     wsId,
				TeamId: teamId,
				Name:   "test-workspace",
			}

			mockClient.EXPECT().GetWorkspace(wsId).Return(workspace, nil)
			mockClient.EXPECT().WorkspaceStatus(wsId).Return(&api.WorkspaceStatus{IsRunning: true}, nil)

			err := c.WakeUpWorkspace(mockClient, wsId)

			Expect(err).ToNot(HaveOccurred())
		})

		It("should return error if GetWorkspace fails", func() {
			mockClient.EXPECT().GetWorkspace(wsId).Return(api.Workspace{}, fmt.Errorf("api error"))

			err := c.WakeUpWorkspace(mockClient, wsId)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get workspace"))
		})

		It("should return error if ScaleWorkspace fails", func() {
			workspace := api.Workspace{
				Id:     wsId,
				TeamId: teamId,
				Name:   "test-workspace",
			}

			mockClient.EXPECT().GetWorkspace(wsId).Return(workspace, nil)
			mockClient.EXPECT().WorkspaceStatus(wsId).Return(&api.WorkspaceStatus{IsRunning: false}, nil)
			mockClient.EXPECT().ScaleWorkspace(wsId, 1).Return(fmt.Errorf("scale error"))

			err := c.WakeUpWorkspace(mockClient, wsId)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to scale workspace"))
		})

		It("should sync landscape when SyncLandscape flag is set", func() {
			workspace := api.Workspace{
				Id:     wsId,
				TeamId: teamId,
				Name:   "test-workspace",
			}
			c.Opts.SyncLandscape = true

			mockClient.EXPECT().GetWorkspace(wsId).Return(workspace, nil)
			mockClient.EXPECT().WorkspaceStatus(wsId).Return(&api.WorkspaceStatus{IsRunning: false}, nil)
			mockClient.EXPECT().ScaleWorkspace(wsId, 1).Return(nil)
			mockClient.EXPECT().WaitForWorkspaceRunning(mock.Anything, mock.Anything).Return(nil)
			mockClient.EXPECT().DeployLandscape(wsId, "").Return(nil)

			err := c.WakeUpWorkspace(mockClient, wsId)

			Expect(err).ToNot(HaveOccurred())
		})

		It("should sync landscape with custom profile", func() {
			workspace := api.Workspace{
				Id:     wsId,
				TeamId: teamId,
				Name:   "test-workspace",
			}
			c.Opts.SyncLandscape = true
			c.Opts.Profile = "prod"

			mockClient.EXPECT().GetWorkspace(wsId).Return(workspace, nil)
			mockClient.EXPECT().WorkspaceStatus(wsId).Return(&api.WorkspaceStatus{IsRunning: false}, nil)
			mockClient.EXPECT().ScaleWorkspace(wsId, 1).Return(nil)
			mockClient.EXPECT().WaitForWorkspaceRunning(mock.Anything, mock.Anything).Return(nil)
			mockClient.EXPECT().DeployLandscape(wsId, "prod").Return(nil)

			err := c.WakeUpWorkspace(mockClient, wsId)

			Expect(err).ToNot(HaveOccurred())
		})

		It("should return error if DeployLandscape fails", func() {
			workspace := api.Workspace{
				Id:     wsId,
				TeamId: teamId,
				Name:   "test-workspace",
			}
			c.Opts.SyncLandscape = true

			mockClient.EXPECT().GetWorkspace(wsId).Return(workspace, nil)
			mockClient.EXPECT().WorkspaceStatus(wsId).Return(&api.WorkspaceStatus{IsRunning: false}, nil)
			mockClient.EXPECT().ScaleWorkspace(wsId, 1).Return(nil)
			mockClient.EXPECT().WaitForWorkspaceRunning(mock.Anything, mock.Anything).Return(nil)
			mockClient.EXPECT().DeployLandscape(wsId, "").Return(fmt.Errorf("deploy error"))

			err := c.WakeUpWorkspace(mockClient, wsId)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to deploy landscape"))
		})

		It("should sync landscape even when workspace is already running", func() {
			workspace := api.Workspace{
				Id:     wsId,
				TeamId: teamId,
				Name:   "test-workspace",
			}
			c.Opts.SyncLandscape = true

			mockClient.EXPECT().GetWorkspace(wsId).Return(workspace, nil)
			mockClient.EXPECT().WorkspaceStatus(wsId).Return(&api.WorkspaceStatus{IsRunning: true}, nil)
			mockClient.EXPECT().DeployLandscape(wsId, "").Return(nil)

			err := c.WakeUpWorkspace(mockClient, wsId)

			Expect(err).ToNot(HaveOccurred())
		})

		It("should scale landscape services when ScaleServices flag is set", func() {
			workspace := api.Workspace{
				Id:     wsId,
				TeamId: teamId,
				Name:   "test-workspace",
			}
			c.Opts.ScaleServices = "web=1,api=2"

			mockClient.EXPECT().GetWorkspace(wsId).Return(workspace, nil)
			mockClient.EXPECT().WorkspaceStatus(wsId).Return(&api.WorkspaceStatus{IsRunning: false}, nil)
			mockClient.EXPECT().ScaleWorkspace(wsId, 1).Return(nil)
			mockClient.EXPECT().WaitForWorkspaceRunning(mock.Anything, mock.Anything).Return(nil)
			mockClient.EXPECT().ScaleLandscapeServices(wsId, map[string]int{"web": 1, "api": 2}).Return(nil)

			err := c.WakeUpWorkspace(mockClient, wsId)

			Expect(err).ToNot(HaveOccurred())
		})

		It("should return error if ScaleLandscapeServices fails", func() {
			workspace := api.Workspace{
				Id:     wsId,
				TeamId: teamId,
				Name:   "test-workspace",
			}
			c.Opts.ScaleServices = "web=1"

			mockClient.EXPECT().GetWorkspace(wsId).Return(workspace, nil)
			mockClient.EXPECT().WorkspaceStatus(wsId).Return(&api.WorkspaceStatus{IsRunning: false}, nil)
			mockClient.EXPECT().ScaleWorkspace(wsId, 1).Return(nil)
			mockClient.EXPECT().WaitForWorkspaceRunning(mock.Anything, mock.Anything).Return(nil)
			mockClient.EXPECT().ScaleLandscapeServices(wsId, map[string]int{"web": 1}).Return(fmt.Errorf("scale services error"))

			err := c.WakeUpWorkspace(mockClient, wsId)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to scale landscape services"))
		})

		It("should return error for invalid scale-services format", func() {
			workspace := api.Workspace{
				Id:     wsId,
				TeamId: teamId,
				Name:   "test-workspace",
			}
			c.Opts.ScaleServices = "invalid"

			mockClient.EXPECT().GetWorkspace(wsId).Return(workspace, nil)
			mockClient.EXPECT().WorkspaceStatus(wsId).Return(&api.WorkspaceStatus{IsRunning: false}, nil)
			mockClient.EXPECT().ScaleWorkspace(wsId, 1).Return(nil)
			mockClient.EXPECT().WaitForWorkspaceRunning(mock.Anything, mock.Anything).Return(nil)

			err := c.WakeUpWorkspace(mockClient, wsId)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to parse scale-services"))
		})

		It("should use both sync-landscape and scale-services together", func() {
			workspace := api.Workspace{
				Id:     wsId,
				TeamId: teamId,
				Name:   "test-workspace",
			}
			c.Opts.SyncLandscape = true
			c.Opts.ScaleServices = "web=1"

			mockClient.EXPECT().GetWorkspace(wsId).Return(workspace, nil)
			mockClient.EXPECT().WorkspaceStatus(wsId).Return(&api.WorkspaceStatus{IsRunning: false}, nil)
			mockClient.EXPECT().ScaleWorkspace(wsId, 1).Return(nil)
			mockClient.EXPECT().WaitForWorkspaceRunning(mock.Anything, mock.Anything).Return(nil)
			mockClient.EXPECT().DeployLandscape(wsId, "").Return(nil)
			mockClient.EXPECT().ScaleLandscapeServices(wsId, map[string]int{"web": 1}).Return(nil)

			err := c.WakeUpWorkspace(mockClient, wsId)

			Expect(err).ToNot(HaveOccurred())
		})
	})
})
