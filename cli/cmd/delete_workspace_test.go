// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
)

var _ = Describe("CreateWorkspace", func() {
	var (
		mockEnv    *cmd.MockEnv
		mockClient *cmd.MockClient
		mockPrompt *cmd.MockPrompt
		c          *cmd.DeleteWorkspaceCmd
		wsId       int
		wsName     string
		confirmed  bool
		ws         api.Workspace
	)

	JustBeforeEach(func() {
		mockClient = cmd.NewMockClient(GinkgoT())
		mockEnv = cmd.NewMockEnv(GinkgoT())
		mockPrompt = cmd.NewMockPrompt(GinkgoT())
		c = &cmd.DeleteWorkspaceCmd{
			Opts: cmd.DeleteWorkspaceOpts{
				GlobalOptions: cmd.GlobalOptions{
					Env:         mockEnv,
					WorkspaceId: &wsId,
				},
				Confirmed: &confirmed,
			},
			Prompt: mockPrompt,
		}
	})
	BeforeEach(func() {
		ws = api.Workspace{
			Id:   wsId,
			Name: wsName,
		}
	})
	Context("Unconfirmed", func() {
		BeforeEach(func() {
			confirmed = false
			wsName = "fake-ws"
		})
		Context("Workspace exists", func() {
			Context("Workspace name entered in confirmation prompt", func() {
				It("deletes the workspace", func() {
					mockClient.EXPECT().GetWorkspace(wsId).Return(ws, nil)
					mockPrompt.EXPECT().InputPrompt("Confirmation delete").Return(ws.Name)
					mockClient.EXPECT().DeleteWorkspace(wsId).Return(nil)
					err := c.DeleteWorkspace(mockClient, wsId)
					Expect(err).ToNot(HaveOccurred())
				})
			})
			Context("Wrong input entered in confirmation prompt", func() {
				It("Returns an error", func() {
					mockClient.EXPECT().GetWorkspace(wsId).Return(ws, nil)
					mockPrompt.EXPECT().InputPrompt("Confirmation delete").Return("other-workspace")
					err := c.DeleteWorkspace(mockClient, wsId)
					Expect(err).To(MatchError("confirmation failed"))
				})
			})
		})

	})

	Context("Confirmed via CLI flag", func() {
		var (
			getWsErr error
		)
		BeforeEach(func() {
			wsId = 42
			ws = api.Workspace{}
			confirmed = true
			getWsErr = nil
		})
		JustBeforeEach(func() {
			mockClient.EXPECT().GetWorkspace(wsId).Return(ws, getWsErr)
		})
		Context("Workspace exists", func() {
			BeforeEach(func() {
				ws = api.Workspace{
					Id: wsId,
				}
			})
			It("Deletes the workspace", func() {
				mockClient.EXPECT().DeleteWorkspace(wsId).Return(nil)
				err := c.DeleteWorkspace(mockClient, wsId)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("Workspace doesn't exist", func() {
			BeforeEach(func() {
				ws = api.Workspace{}
				getWsErr = errors.New("404")
			})

			It("Returns an error", func() {
				err := c.DeleteWorkspace(mockClient, wsId)
				Expect(err).To(MatchError("failed to get workspace 42: 404"))
			})
		})
	})

})
