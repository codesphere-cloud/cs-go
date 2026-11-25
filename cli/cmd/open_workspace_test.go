// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("GenerateWorkspacePath", func() {
	var (
		mockEnv     *cmd.MockEnv
		mockClient  *cmd.MockClient
		mockBrowser *cmd.MockBrowser
		o           *cmd.OpenWorkspaceCmd
		wsId        int
		teamId      int
	)

	JustBeforeEach(func() {
		mockClient = cmd.NewMockClient(GinkgoT())
		mockBrowser = cmd.NewMockBrowser(GinkgoT())
		mockEnv = cmd.NewMockEnv(GinkgoT())
		wsId = 42
		teamId = 21
		o = &cmd.OpenWorkspaceCmd{
			Opts: &cmd.GlobalOptions{
				Env:         mockEnv,
				WorkspaceId: wsId,
			},
		}
	})

	It("queries the workspace and opens the IDE path", func() {
		mockClient.EXPECT().GetWorkspace(mock.Anything, wsId).Return(api.Workspace{
			Id:     wsId,
			TeamId: teamId,
		}, nil)
		mockBrowser.EXPECT().OpenIde(fmt.Sprintf("teams/%d/workspaces/%d", teamId, wsId)).Return(nil)
		err := o.OpenWorkspace(context.Background(), mockBrowser, mockClient, wsId)
		Expect(err).ToNot(HaveOccurred())

	})

})
