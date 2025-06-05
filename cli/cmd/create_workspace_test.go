// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
)

var _ = Describe("CreateWorkspace", func() {
	var (
		mockEnv    *cmd.MockEnv
		mockClient *cmd.MockClient
		c          *cmd.CreateWorkspaceCmd
		teamId     int
		wsName     string
		env        []string
		repoStr    string
		repo       *string
		vpn        *string
		vpnStr     string
		plan       int
		private    bool
		timeout    time.Duration
		branchStr  string
		branch     *string
		deployArgs api.DeployWorkspaceArgs
	)

	BeforeEach(func() {
		env = []string{}
		repoStr = "https://fake-git.com/my/repo.git"
		repo = nil
		vpnStr = "MyVpn"
		vpn = nil
		plan = 8
		private = false
		timeout = 30 * time.Second
		branchStr = "fake-branch"
		branch = nil
	})

	JustBeforeEach(func() {
		mockClient = cmd.NewMockClient(GinkgoT())
		mockEnv = cmd.NewMockEnv(GinkgoT())
		wsName = "foo-workspace"
		teamId = 21
		c = &cmd.CreateWorkspaceCmd{
			Opts: cmd.CreateWorkspaceOpts{
				GlobalOptions: cmd.GlobalOptions{
					Env:    mockEnv,
					TeamId: &teamId,
				},
				Env:     &env,
				Repo:    repo,
				Vpn:     vpn,
				Plan:    &plan,
				Private: &private,
				Timeout: &timeout,
				Branch:  branch,
			},
		}
		envMap, err := cs.ArgToEnvVarMap(env)
		Expect(err).ToNot(HaveOccurred())
		deployArgs = api.DeployWorkspaceArgs{
			Name:          wsName,
			TeamId:        teamId,
			EnvVars:       envMap,
			GitUrl:        repo,
			VpnConfigName: vpn,
			PlanId:        plan,
			IsPrivateRepo: private,
			Timeout:       timeout,
			Branch:        branch,
		}
	})

	Context("Minimal values are set", func() {

		It("Creates the workspace", func() {
			mockClient.EXPECT().DeployWorkspace(deployArgs).Return(&api.Workspace{Name: wsName}, nil)
			ws, err := c.CreateWorkspace(mockClient, teamId, wsName)
			Expect(err).ToNot(HaveOccurred())
			Expect(ws.Name).To(Equal(wsName))
		})
	})

	Context("All values are set", func() {
		BeforeEach(func() {
			env = []string{"foo=bla", "blib=blub"}
			repo = &repoStr
			vpn = &vpnStr
			private = true
			timeout = 120 * time.Second
			branch = &branchStr
			wsName = "different-name"
		})
		It("Creates the workspace and passes expected arguments", func() {
			mockClient.EXPECT().DeployWorkspace(deployArgs).Return(&api.Workspace{Name: wsName}, nil)
			ws, err := c.CreateWorkspace(mockClient, teamId, wsName)
			Expect(err).ToNot(HaveOccurred())
			Expect(ws.Name).To(Equal(wsName))
		})
	})

})
