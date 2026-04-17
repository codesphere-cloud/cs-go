// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
)

var _ = Describe("GenerateWorkspacePath", func() {
	var (
		mockEnv     *cs.MockEnv
		mockClient  *cmd.MockClient
		mockBrowser *cs.MockBrowser
		o           *cmd.OpenWorkspaceCmd
		wsId        int
		teamId      int
	)

	JustBeforeEach(func() {
		mockClient = cmd.NewMockClient(GinkgoT())
		mockBrowser = cs.NewMockBrowser(GinkgoT())
		mockEnv = cs.NewMockEnv(GinkgoT())
		wsId = 42
		teamId = 21
		o = &cmd.OpenWorkspaceCmd{
			Opts: cmd.NewGlobalOptionsWithCustomEnv(cmd.GlobalOptions{
				WorkspaceId: wsId,
			}, mockEnv),
		}
	})

	It("queries the workspace and opens the IDE path", func() {
		mockClient.EXPECT().GetWorkspace(wsId).Return(api.Workspace{
			Id:     wsId,
			TeamId: teamId,
		}, nil)
		mockBrowser.EXPECT().OpenIde(fmt.Sprintf("teams/%d/workspaces/%d", teamId, wsId), mock.Anything).Return(nil)
		err := o.OpenWorkspace(mockBrowser, mockClient, wsId)
		Expect(err).ToNot(HaveOccurred())

	})

})
