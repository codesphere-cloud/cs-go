// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"errors"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
	mcpserver "github.com/codesphere-cloud/cs-go/mcp-server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("update_workspace", func() {
	const wsId = 1057556
	var (
		client    *cmd.MockClient
		session   *mcp.ClientSession
		defaultWs int
		devDomain string
	)

	BeforeEach(func() {
		client = cmd.NewMockClient(GinkgoT())
		defaultWs = wsId
		devDomain = "1057556.7.ws.dev-codesphere.com"
		session = connect(func(s *mcp.Server) {
			mcpserver.RegisterAccessTools(s, mcpserver.Deps{
				Client:             client,
				DefaultWorkspaceId: func() (int, error) { return defaultWs, nil },
			})
		})
	})

	It("is registered under the exact name with a description that names the approval", func() {
		tool := findTool(session, "update_workspace")

		Expect(tool.Description).To(ContainSubstring("changes who can reach the workspace"))
		Expect(tool.Description).To(ContainSubstring("The user is asked to approve every call"))
		Expect(tool.Description).To(ContainSubstring("defaults to CS_WORKSPACE_ID"))
		Expect(requiredArgs(tool)).To(BeEmpty())
		Expect(*tool.Annotations.DestructiveHint).To(BeTrue())
	})

	It("makes the session's workspace public and returns it as read back", func() {
		client.EXPECT().UpdateWorkspace(wsId, api.UpdateWorkspaceArgs{Restricted: new(false)}).Return(nil).Once()
		client.EXPECT().GetWorkspace(wsId).Return(api.Workspace{Id: wsId, Name: "docs", Restricted: false, DevDomain: &devDomain}, nil).Once()

		out := structured(callTool(session, "update_workspace", map[string]any{"restricted": false}))

		Expect(out["restricted"]).To(BeFalse())
		Expect(out["devDomainUrl"]).To(Equal("https://1057556.7.ws.dev-codesphere.com/"))
		Expect(out["access"]).To(Equal("public: reachable without sign-in"))
		Expect(out["workspace"]).To(HaveKeyWithValue("name", "docs"))
	})

	It("uses an explicit workspace ID and changes name and plan", func() {
		client.EXPECT().UpdateWorkspace(42, api.UpdateWorkspaceArgs{Name: new("docs"), PlanId: new(21)}).Return(nil).Once()
		client.EXPECT().GetWorkspace(42).Return(api.Workspace{Id: 42, Name: "docs", PlanId: 21, Restricted: true}, nil).Once()

		out := structured(callTool(session, "update_workspace", map[string]any{"workspaceId": 42, "name": "docs", "planId": 21}))

		Expect(out["access"]).To(Equal("restricted: visitors are redirected to the Codesphere sign-in"))
	})

	It("rejects a call without changes", func() {
		res := callTool(session, "update_workspace", map[string]any{})

		Expect(res.IsError).To(BeTrue())
		Expect(text(res)).To(Equal("nothing to change: set at least one of restricted, name or planId"))
	})

	It("asks for a workspace ID when CS_WORKSPACE_ID is not set", func() {
		defaultWs = -1

		res := callTool(session, "update_workspace", map[string]any{"restricted": false})

		Expect(res.IsError).To(BeTrue())
		Expect(text(res)).To(Equal("workspaceId is required: CS_WORKSPACE_ID is not set"))
	})

	It("passes the API error on", func() {
		client.EXPECT().UpdateWorkspace(wsId, api.UpdateWorkspaceArgs{Restricted: new(false)}).
			Return(errors.New("codesphere API returned error 403 (Forbidden): access change not approved")).Once()

		res := callTool(session, "update_workspace", map[string]any{"restricted": false})

		Expect(res.IsError).To(BeTrue())
		Expect(text(res)).To(Equal("failed to update workspace 1057556: codesphere API returned error 403 (Forbidden): access change not approved"))
	})

	It("reports a failed read-back", func() {
		client.EXPECT().UpdateWorkspace(wsId, api.UpdateWorkspaceArgs{Restricted: new(true)}).Return(nil).Once()
		client.EXPECT().GetWorkspace(wsId).Return(api.Workspace{}, errors.New("boom")).Once()

		res := callTool(session, "update_workspace", map[string]any{"restricted": true})

		Expect(res.IsError).To(BeTrue())
		Expect(text(res)).To(Equal("workspace 1057556 was updated, but reading it back failed: boom"))
	})
})
