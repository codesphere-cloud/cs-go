// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package deploy_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/pkg/deploy"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Deployer", func() {
	var (
		mockClient *deploy.MockClient
		deployer   *deploy.Deployer
		teamId     int
		wsName     string
	)

	BeforeEach(func() {
		teamId = 5
		wsName = "my-app-#42"
	})

	JustBeforeEach(func() {
		mockClient = deploy.NewMockClient(GinkgoT())
		deployer = deploy.NewDeployer(mockClient)
	})

	Describe("FindWorkspace", func() {
		Context("when workspace exists", func() {
			It("returns the matching workspace", func() {
				workspaces := []api.Workspace{
					{Id: 100, Name: "other-ws"},
					{Id: 200, Name: wsName},
				}
				mockClient.EXPECT().ListWorkspaces(teamId).Return(workspaces, nil)

				ws, err := deployer.FindWorkspace(teamId, wsName)
				Expect(err).ToNot(HaveOccurred())
				Expect(ws).ToNot(BeNil())
				Expect(ws.Id).To(Equal(200))
				Expect(ws.Name).To(Equal(wsName))
			})
		})

		Context("when workspace does not exist", func() {
			It("returns nil without error", func() {
				workspaces := []api.Workspace{
					{Id: 100, Name: "other-ws"},
				}
				mockClient.EXPECT().ListWorkspaces(teamId).Return(workspaces, nil)

				ws, err := deployer.FindWorkspace(teamId, wsName)
				Expect(err).ToNot(HaveOccurred())
				Expect(ws).To(BeNil())
			})
		})

		Context("when listing fails", func() {
			It("returns the error", func() {
				mockClient.EXPECT().ListWorkspaces(teamId).Return(nil, errors.New("api error"))

				ws, err := deployer.FindWorkspace(teamId, wsName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("listing workspaces"))
				Expect(ws).To(BeNil())
			})
		})
	})

	Describe("CreateWorkspace", func() {
		var cfg deploy.Config

		BeforeEach(func() {
			cfg = deploy.Config{
				TeamId:  teamId,
				PlanId:  8,
				Name:    wsName,
				EnvVars: map[string]string{"KEY": "val"},
				Branch:  "feature-branch",
				RepoUrl: "https://github.com/org/repo.git",
				Timeout: 5 * time.Minute,
			}
		})

		It("creates workspace with correct args", func() {
			branch := "feature-branch"
			repoUrl := "https://github.com/org/repo.git"
			mockClient.EXPECT().DeployWorkspace(api.DeployWorkspaceArgs{
				TeamId:        teamId,
				PlanId:        8,
				Name:          wsName,
				EnvVars:       map[string]string{"KEY": "val"},
				IsPrivateRepo: true,
				GitUrl:        &repoUrl,
				Branch:        &branch,
				Timeout:       5 * time.Minute,
			}).Return(&api.Workspace{Id: 300, Name: wsName}, nil)

			ws, err := deployer.CreateWorkspace(cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(ws.Id).To(Equal(300))
		})

		It("returns error when deploy fails", func() {
			mockClient.EXPECT().DeployWorkspace(mock.Anything).Return(nil, errors.New("deploy failed"))

			ws, err := deployer.CreateWorkspace(cfg)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("creating workspace"))
			Expect(ws).To(BeNil())
		})
	})

	Describe("UpdateWorkspace", func() {
		var (
			ws  *api.Workspace
			cfg deploy.Config
		)

		BeforeEach(func() {
			ws = &api.Workspace{Id: 200, Name: wsName}
			cfg = deploy.Config{
				Branch:  "feature-branch",
				EnvVars: map[string]string{"KEY": "val"},
				Timeout: 5 * time.Minute,
			}
		})

		It("waits for running, pulls, and sets env vars", func() {
			mockClient.EXPECT().WaitForWorkspaceRunning(ws, 5*time.Minute).Return(nil)
			mockClient.EXPECT().GitPull(200, "origin", "feature-branch").Return(nil)
			mockClient.EXPECT().SetEnvVarOnWorkspace(200, map[string]string{"KEY": "val"}).Return(nil)

			err := deployer.UpdateWorkspace(ws, cfg)
			Expect(err).ToNot(HaveOccurred())
		})

		It("skips env vars when none provided", func() {
			cfg.EnvVars = map[string]string{}
			mockClient.EXPECT().WaitForWorkspaceRunning(ws, 5*time.Minute).Return(nil)
			mockClient.EXPECT().GitPull(200, "origin", "feature-branch").Return(nil)
			// SetEnvVarOnWorkspace should NOT be called

			err := deployer.UpdateWorkspace(ws, cfg)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error when wait fails", func() {
			mockClient.EXPECT().WaitForWorkspaceRunning(ws, 5*time.Minute).Return(errors.New("timeout"))

			err := deployer.UpdateWorkspace(ws, cfg)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("timeout"))
		})

		It("returns error when git pull fails", func() {
			mockClient.EXPECT().WaitForWorkspaceRunning(ws, 5*time.Minute).Return(nil)
			mockClient.EXPECT().GitPull(200, "origin", "feature-branch").Return(errors.New("pull failed"))

			err := deployer.UpdateWorkspace(ws, cfg)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("git pull"))
		})
	})

	Describe("DeleteWorkspace", func() {
		It("deletes the workspace", func() {
			mockClient.EXPECT().DeleteWorkspace(200).Return(nil)

			err := deployer.DeleteWorkspace(200)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error on failure", func() {
			mockClient.EXPECT().DeleteWorkspace(200).Return(errors.New("delete failed"))

			err := deployer.DeleteWorkspace(200)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Deploy", func() {
		var cfg deploy.Config

		BeforeEach(func() {
			cfg = deploy.Config{
				TeamId:  teamId,
				PlanId:  8,
				Name:    wsName,
				EnvVars: map[string]string{},
				Branch:  "feature-branch",
				RepoUrl: "https://github.com/org/repo.git",
				Stages:  []string{},
				Timeout: 5 * time.Minute,
			}
		})

		Context("delete mode", func() {
			It("finds and deletes existing workspace", func() {
				workspaces := []api.Workspace{{Id: 200, Name: wsName}}
				mockClient.EXPECT().ListWorkspaces(teamId).Return(workspaces, nil)
				mockClient.EXPECT().DeleteWorkspace(200).Return(nil)

				result, err := deployer.Deploy(cfg, true)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("does nothing when workspace not found", func() {
				mockClient.EXPECT().ListWorkspaces(teamId).Return([]api.Workspace{}, nil)

				result, err := deployer.Deploy(cfg, true)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})
		})

		Context("create mode (no existing workspace)", func() {
			It("creates a new workspace and returns result", func() {
				// FindWorkspace returns nothing
				mockClient.EXPECT().ListWorkspaces(teamId).Return([]api.Workspace{}, nil)
				// CreateWorkspace
				mockClient.EXPECT().DeployWorkspace(mock.Anything).Return(&api.Workspace{Id: 300, Name: wsName}, nil)

				result, err := deployer.Deploy(cfg, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(result.WorkspaceId).To(Equal(300))
				Expect(result.WorkspaceURL).To(ContainSubstring("300"))
			})
		})

		Context("update mode (existing workspace)", func() {
			It("updates existing workspace and returns result", func() {
				existing := &api.Workspace{Id: 200, Name: wsName}
				workspaces := []api.Workspace{*existing}
				// FindWorkspace
				mockClient.EXPECT().ListWorkspaces(teamId).Return(workspaces, nil)
				// UpdateWorkspace
				mockClient.EXPECT().WaitForWorkspaceRunning(mock.Anything, 5*time.Minute).Return(nil)
				mockClient.EXPECT().GitPull(200, "origin", "feature-branch").Return(nil)

				result, err := deployer.Deploy(cfg, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(result.WorkspaceId).To(Equal(200))
			})
		})
	})
})
