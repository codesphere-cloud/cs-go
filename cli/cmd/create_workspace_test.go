// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
)

var _ = Describe("CreateWorkspace", func() {
	var (
		mockEnv      *cmd.MockEnv
		mockClient   *cmd.MockClient
		c            *cmd.CreateWorkspaceCmd
		teamId       int
		wsName       string
		env          []string
		repoStr      string
		repo         *string
		vpn          *string
		vpnStr       string
		plan         int
		private      bool
		timeout      time.Duration
		branchStr    string
		branch       *string
		baseimageStr string
		baseimage    *string
		deployArgs   api.DeployWorkspaceArgs
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
		baseimageStr = "ubuntu-24.04"
		baseimage = nil
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
				Env:       &env,
				Repo:      repo,
				Vpn:       vpn,
				Plan:      &plan,
				Private:   &private,
				Timeout:   &timeout,
				Branch:    branch,
				Baseimage: baseimage,
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
			BaseImage:     baseimage,
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

	Context("Baseimage is specified", func() {
		BeforeEach(func() {
			baseimage = &baseimageStr
		})

		It("validates baseimage exists and creates workspace", func() {
			supportedUntil, _ := time.Parse("2006-01-02", "2025-12-31")
			availableBaseimages := []api.Baseimage{
				{Id: "ubuntu-20.04", Name: "Ubuntu 20.04", SupportedUntil: supportedUntil},
				{Id: "ubuntu-24.04", Name: "Ubuntu 24.04", SupportedUntil: supportedUntil},
				{Id: "node-18", Name: "Node.js 18", SupportedUntil: supportedUntil},
			}
			mockClient.EXPECT().ListBaseimages().Return(availableBaseimages, nil)
			mockClient.EXPECT().DeployWorkspace(deployArgs).Return(&api.Workspace{Name: wsName}, nil)

			ws, err := c.CreateWorkspace(mockClient, teamId, wsName)
			Expect(err).ToNot(HaveOccurred())
			Expect(ws.Name).To(Equal(wsName))
		})

		It("fails when baseimage does not exist", func() {
			supportedUntil, _ := time.Parse("2006-01-02", "2025-12-31")
			availableBaseimages := []api.Baseimage{
				{Id: "ubuntu-20.04", Name: "Ubuntu 20.04", SupportedUntil: supportedUntil},
				{Id: "node-18", Name: "Node.js 18", SupportedUntil: supportedUntil},
			}
			mockClient.EXPECT().ListBaseimages().Return(availableBaseimages, nil)

			ws, err := c.CreateWorkspace(mockClient, teamId, wsName)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("base image 'ubuntu-24.04' not found"))
			Expect(err.Error()).To(ContainSubstring("available options are: ubuntu-20.04, node-18"))
			Expect(ws).To(BeNil())
		})

		It("fails when ListBaseimages returns error", func() {
			mockClient.EXPECT().ListBaseimages().Return([]api.Baseimage{}, errors.New("API error"))

			ws, err := c.CreateWorkspace(mockClient, teamId, wsName)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to list base images: API error"))
			Expect(ws).To(BeNil())
		})
	})

	Context("Error handling", func() {
		It("fails when environment variables are malformed", func() {
			env = []string{"invalid-env-var"}

			ws, err := c.CreateWorkspace(mockClient, teamId, wsName)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to parse environment variables"))
			Expect(ws).To(BeNil())
		})

		It("fails when DeployWorkspace returns error", func() {
			mockClient.EXPECT().DeployWorkspace(deployArgs).Return(nil, errors.New("deployment failed"))

			ws, err := c.CreateWorkspace(mockClient, teamId, wsName)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to create workspace: deployment failed"))
			Expect(ws).To(BeNil())
		})
	})
})
