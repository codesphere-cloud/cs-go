// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package update_test

import (
	"errors"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
	"github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	subject "github.com/codesphere-cloud/cs-go/cli/cmd/update"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("UpdateWorkspace", func() {
	const wsId = 1057556
	var (
		client  *cmd.MockClient
		env     *cmd.MockEnv
		opts    *cmd.GlobalOptions
		c       *subject.UpdateWorkspaceCmd
		devHost string
	)

	BeforeEach(func() {
		client = cmd.NewMockClient(GinkgoT())
		env = cmd.NewMockEnv(GinkgoT())
		opts = &cmd.GlobalOptions{WorkspaceId: wsId, Env: env}
		devHost = "1057556.7.ws.dev-codesphere.com"
		c = &subject.UpdateWorkspaceCmd{
			Opts:          subject.UpdateWorkspaceOpts{RootOptions: opts},
			ClientFactory: func(shared.RootOptions) (cmd.Client, error) { return client, nil },
		}
	})

	It("makes the workspace public and returns the workspace read back afterwards", func() {
		c.Opts.Restricted = new(false)
		client.EXPECT().UpdateWorkspace(wsId, api.UpdateWorkspaceArgs{Restricted: new(false)}).Return(nil).Once()
		client.EXPECT().GetWorkspace(wsId).Return(api.Workspace{Id: wsId, Name: "docs", Restricted: false, DevDomain: &devHost}, nil).Once()

		ws, err := c.UpdateWorkspace(client, wsId)

		Expect(err).NotTo(HaveOccurred())
		Expect(ws.Restricted).To(BeFalse())
	})

	It("sends name and plan without touching access", func() {
		c.Opts.Name = new("docs")
		c.Opts.PlanId = new(21)
		client.EXPECT().UpdateWorkspace(wsId, api.UpdateWorkspaceArgs{Name: new("docs"), PlanId: new(21)}).Return(nil).Once()
		client.EXPECT().GetWorkspace(wsId).Return(api.Workspace{Id: wsId, Name: "docs", PlanId: 21}, nil).Once()

		ws, err := c.UpdateWorkspace(client, wsId)

		Expect(err).NotTo(HaveOccurred())
		Expect(ws.PlanId).To(Equal(21))
	})

	It("refuses to call the API without any change", func() {
		_, err := c.UpdateWorkspace(client, wsId)

		Expect(err).To(MatchError("nothing to change: set --public, --private, --restricted, --name or --plan"))
	})

	It("rejects a negative plan ID", func() {
		c.Opts.PlanId = new(-1)

		_, err := c.UpdateWorkspace(client, wsId)

		Expect(err).To(MatchError("plan ID must be non-negative"))
	})

	It("explains that refused access changes need the user's approval", func() {
		c.Opts.Restricted = new(false)
		client.EXPECT().UpdateWorkspace(wsId, api.UpdateWorkspaceArgs{Restricted: new(false)}).
			Return(errors.New("codesphere API returned error 403 (Forbidden)")).Once()

		_, err := c.UpdateWorkspace(client, wsId)

		Expect(err).To(MatchError(ContainSubstring("failed to update workspace 1057556: codesphere API returned error 403 (Forbidden)")))
		Expect(err).To(MatchError(ContainSubstring("use the MCP tool update_workspace")))
	})

	It("does not add the approval hint to other errors", func() {
		c.Opts.Name = new("docs")
		client.EXPECT().UpdateWorkspace(wsId, api.UpdateWorkspaceArgs{Name: new("docs")}).
			Return(errors.New("codesphere API returned error 404 (Workspace is not found.)")).Once()

		_, err := c.UpdateWorkspace(client, wsId)

		Expect(err).To(MatchError("failed to update workspace 1057556: codesphere API returned error 404 (Workspace is not found.)"))
	})

	It("reports a failed read-back after a successful update", func() {
		c.Opts.Restricted = new(true)
		client.EXPECT().UpdateWorkspace(wsId, api.UpdateWorkspaceArgs{Restricted: new(true)}).Return(nil).Once()
		client.EXPECT().GetWorkspace(wsId).Return(api.Workspace{}, errors.New("boom")).Once()

		_, err := c.UpdateWorkspace(client, wsId)

		Expect(err).To(MatchError("workspace 1057556 was updated, but reading it back failed: boom"))
	})

	It("reads --public from the command line", func() {
		parent := &cobra.Command{Use: "update"}
		subject.AddUpdateWorkspaceCmd(parent, opts)
		wsCmd, _, err := parent.Find([]string{"workspace"})
		Expect(err).NotTo(HaveOccurred())
		Expect(wsCmd.Flags().Parse([]string{"--public"})).To(Succeed())
		c.Opts.Name = new("stale value from an earlier run")
		client.EXPECT().UpdateWorkspace(wsId, api.UpdateWorkspaceArgs{Restricted: new(false)}).Return(nil).Once()
		client.EXPECT().GetWorkspace(wsId).Return(api.Workspace{Id: wsId, DevDomain: &devHost}, nil).Once()

		Expect(c.RunE(wsCmd, nil)).To(Succeed())
	})

	It("reads --restricted=false from the command line", func() {
		parent := &cobra.Command{Use: "update"}
		subject.AddUpdateWorkspaceCmd(parent, opts)
		wsCmd, _, err := parent.Find([]string{"workspace"})
		Expect(err).NotTo(HaveOccurred())
		Expect(wsCmd.Flags().Parse([]string{"--restricted=false", "--plan", "8"})).To(Succeed())
		client.EXPECT().UpdateWorkspace(wsId, api.UpdateWorkspaceArgs{Restricted: new(false), PlanId: new(8)}).Return(nil).Once()
		client.EXPECT().GetWorkspace(wsId).Return(api.Workspace{Id: wsId}, nil).Once()

		Expect(c.RunE(wsCmd, nil)).To(Succeed())
	})

	It("fails before creating a client when the workspace ID is missing", func() {
		opts.WorkspaceId = -1
		env.EXPECT().GetWorkspaceId().Return(-1, nil).Once()

		Expect(c.RunE(nil, nil)).To(MatchError(ContainSubstring("workspace ID not set")))
	})

	It("reports client creation errors", func() {
		c.ClientFactory = func(shared.RootOptions) (cmd.Client, error) { return nil, errors.New("no token") }

		Expect(c.RunE(nil, nil)).To(MatchError("failed to create Codesphere client: no token"))
	})

	It("registers the command with mutually exclusive access flags", func() {
		parent := &cobra.Command{Use: "update"}
		subject.AddUpdateWorkspaceCmd(parent, opts)
		wsCmd, _, err := parent.Find([]string{"workspace"})

		Expect(err).NotTo(HaveOccurred())
		Expect(wsCmd.Use).To(Equal("workspace"))
		Expect(wsCmd.RunE).NotTo(BeNil())
		Expect(wsCmd.Flags().Lookup("public")).NotTo(BeNil())
		Expect(wsCmd.Flags().Lookup("private")).NotTo(BeNil())
		Expect(wsCmd.Flags().Lookup("restricted")).NotTo(BeNil())
	})
})

var _ = Describe("FormatWorkspaceAccess", func() {
	It("names the sign-in redirect for restricted workspaces", func() {
		host := "1.7.ws.example.com"
		out := subject.FormatWorkspaceAccess(&api.Workspace{Id: 1, Name: "a", Restricted: true, DevDomain: &host})
		Expect(out).To(ContainSubstring("Access: restricted, visitors of https://1.7.ws.example.com/ are redirected to the Codesphere sign-in"))
	})

	It("shows the public URL for public workspaces", func() {
		host := "1.7.ws.example.com"
		out := subject.FormatWorkspaceAccess(&api.Workspace{Id: 1, Name: "a", Restricted: false, DevDomain: &host})
		Expect(out).To(ContainSubstring("Access: public, https://1.7.ws.example.com/ is reachable without sign-in"))
	})
})
