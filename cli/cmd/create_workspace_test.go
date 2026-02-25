// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

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
		repo         *string
		vpn          *string
		plan         int
		private      bool
		timeout      time.Duration
		branch       *string
		baseimageStr string
		baseimage    *string
		deployArgs   api.DeployWorkspaceArgs
	)

	BeforeEach(func() {
		env = []string{}
		repo = nil
		vpn = nil
		plan = 8
		private = false
		timeout = 30 * time.Second
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
					TeamId: teamId,
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
		It("Creates workspace with all flags set", func() {
			createCmd := &cobra.Command{Use: "create"}
			opts := &cmd.GlobalOptions{Env: cs.NewEnv()}

			cmd.AddCreateWorkspaceCmd(createCmd, *opts)

			createCmd.SetArgs([]string{
				"workspace",
				"test-workspace",
				"--repository", "https://github.com/test/repo.git",
				"--vpn", "test-vpn",
				"--env", "FOO=bar",
				"--env", "BAZ=qux",
				"--plan", "20",
				"--private",
				"--timeout", "2m",
				"--branch", "develop",
				"--base-image", "ubuntu-24.04",
				"--public-dev-domain=false",
			})

			// Override RunE to avoid actually creating workspace
			createCmd.Commands()[0].RunE = func(cmd *cobra.Command, args []string) error {
				return nil
			}

			err := createCmd.Execute()
			Expect(err).NotTo(HaveOccurred())
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

	Context("Repository URL validation", func() {
		It("validates and sets repository URL when provided", func() {
			repoUrl := "https://github.com/test/repo.git"
			c.Opts.Repo = &repoUrl

			expectedArgs := api.DeployWorkspaceArgs{
				Name:          wsName,
				TeamId:        teamId,
				EnvVars:       map[string]string{},
				GitUrl:        &repoUrl,
				VpnConfigName: nil,
				PlanId:        plan,
				IsPrivateRepo: private,
				Timeout:       timeout,
				Branch:        nil,
				BaseImage:     nil,
			}

			mockClient.EXPECT().DeployWorkspace(expectedArgs).Return(&api.Workspace{Name: wsName}, nil)

			ws, err := c.CreateWorkspace(mockClient, teamId, wsName)
			Expect(err).ToNot(HaveOccurred())
			Expect(ws.Name).To(Equal(wsName))
		})

		It("fails when repository URL is invalid", func() {
			invalidUrl := "not-a-valid-url"
			c.Opts.Repo = &invalidUrl

			ws, err := c.CreateWorkspace(mockClient, teamId, wsName)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("validation of repository URL failed"))
			Expect(ws).To(BeNil())
		})

		It("does not set GitUrl when repository is empty string", func() {
			emptyRepo := ""
			c.Opts.Repo = &emptyRepo

			expectedArgs := api.DeployWorkspaceArgs{
				Name:          wsName,
				TeamId:        teamId,
				EnvVars:       map[string]string{},
				GitUrl:        nil,
				VpnConfigName: nil,
				PlanId:        plan,
				IsPrivateRepo: private,
				Timeout:       timeout,
				Branch:        nil,
				BaseImage:     nil,
			}

			mockClient.EXPECT().DeployWorkspace(expectedArgs).Return(&api.Workspace{Name: wsName}, nil)

			ws, err := c.CreateWorkspace(mockClient, teamId, wsName)
			Expect(err).ToNot(HaveOccurred())
			Expect(ws.Name).To(Equal(wsName))
		})
	})

	Context("VPN configuration", func() {
		It("sets VPN config when provided", func() {
			vpnName := "test-vpn"
			c.Opts.Vpn = &vpnName

			expectedArgs := api.DeployWorkspaceArgs{
				Name:          wsName,
				TeamId:        teamId,
				EnvVars:       map[string]string{},
				GitUrl:        nil,
				VpnConfigName: &vpnName,
				PlanId:        plan,
				IsPrivateRepo: private,
				Timeout:       timeout,
				Branch:        nil,
				BaseImage:     nil,
			}

			mockClient.EXPECT().DeployWorkspace(expectedArgs).Return(&api.Workspace{Name: wsName}, nil)

			ws, err := c.CreateWorkspace(mockClient, teamId, wsName)
			Expect(err).ToNot(HaveOccurred())
			Expect(ws.Name).To(Equal(wsName))
		})

		It("does not set VPN when empty string", func() {
			emptyVpn := ""
			c.Opts.Vpn = &emptyVpn

			expectedArgs := api.DeployWorkspaceArgs{
				Name:          wsName,
				TeamId:        teamId,
				EnvVars:       map[string]string{},
				GitUrl:        nil,
				VpnConfigName: nil,
				PlanId:        plan,
				IsPrivateRepo: private,
				Timeout:       timeout,
				Branch:        nil,
				BaseImage:     nil,
			}

			mockClient.EXPECT().DeployWorkspace(expectedArgs).Return(&api.Workspace{Name: wsName}, nil)

			ws, err := c.CreateWorkspace(mockClient, teamId, wsName)
			Expect(err).ToNot(HaveOccurred())
			Expect(ws.Name).To(Equal(wsName))
		})
	})

	Context("Branch configuration", func() {
		It("sets branch when provided", func() {
			branchName := "feature-branch"
			c.Opts.Branch = &branchName

			expectedArgs := api.DeployWorkspaceArgs{
				Name:          wsName,
				TeamId:        teamId,
				EnvVars:       map[string]string{},
				GitUrl:        nil,
				VpnConfigName: nil,
				PlanId:        plan,
				IsPrivateRepo: private,
				Timeout:       timeout,
				Branch:        &branchName,
				BaseImage:     nil,
			}

			mockClient.EXPECT().DeployWorkspace(expectedArgs).Return(&api.Workspace{Name: wsName}, nil)

			ws, err := c.CreateWorkspace(mockClient, teamId, wsName)
			Expect(err).ToNot(HaveOccurred())
			Expect(ws.Name).To(Equal(wsName))
		})

		It("does not set branch when empty string", func() {
			emptyBranch := ""
			c.Opts.Branch = &emptyBranch

			expectedArgs := api.DeployWorkspaceArgs{
				Name:          wsName,
				TeamId:        teamId,
				EnvVars:       map[string]string{},
				GitUrl:        nil,
				VpnConfigName: nil,
				PlanId:        plan,
				IsPrivateRepo: private,
				Timeout:       timeout,
				Branch:        nil,
				BaseImage:     nil,
			}

			mockClient.EXPECT().DeployWorkspace(expectedArgs).Return(&api.Workspace{Name: wsName}, nil)

			ws, err := c.CreateWorkspace(mockClient, teamId, wsName)
			Expect(err).ToNot(HaveOccurred())
			Expect(ws.Name).To(Equal(wsName))
		})
	})

	Context("Public dev domain flag", func() {
		// Currently we can only test the case when the flag is not set
		// as cmd is not exported and we cannot check the Changed field easily.
		It("does not set restricted when flag not set", func() {
			expectedArgs := api.DeployWorkspaceArgs{
				Name:          wsName,
				TeamId:        teamId,
				EnvVars:       map[string]string{},
				GitUrl:        nil,
				VpnConfigName: nil,
				PlanId:        plan,
				IsPrivateRepo: private,
				Timeout:       timeout,
				Branch:        nil,
				BaseImage:     nil,
				Restricted:    nil,
			}

			mockClient.EXPECT().DeployWorkspace(expectedArgs).Return(&api.Workspace{Name: wsName}, nil)

			ws, err := c.CreateWorkspace(mockClient, teamId, wsName)
			Expect(err).ToNot(HaveOccurred())
			Expect(ws.Name).To(Equal(wsName))
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
