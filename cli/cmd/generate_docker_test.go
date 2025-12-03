// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"errors"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/codesphere-cloud/cs-go/api"
	openapi_client "github.com/codesphere-cloud/cs-go/api/openapi_client"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
	"github.com/codesphere-cloud/cs-go/pkg/ci"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/codesphere-cloud/cs-go/pkg/exporter"
	"github.com/codesphere-cloud/cs-go/pkg/git"
)

var _ = Describe("GenerateDocker", func() {
	var (
		memoryFs     *cs.FileSystem
		mockEnv      *cmd.MockEnv
		mockExporter *exporter.MockExporter
		mockGit      *git.MockGit
		mockClient   *cmd.MockClient
		c            *cmd.GenerateDockerCmd
		wsId         int
	)

	BeforeEach(func() {
		memoryFs = cs.NewMemFileSystem()
		mockEnv = cmd.NewMockEnv(GinkgoT())
		mockExporter = exporter.NewMockExporter(GinkgoT())
		mockGit = git.NewMockGit(GinkgoT())
		mockClient = cmd.NewMockClient(GinkgoT())

		defaultInput := "ci.yml"
		defaultOutput := "./export"
		wsId = 1
		c = &cmd.GenerateDockerCmd{
			Opts: &cmd.GenerateDockerOpts{
				GenerateOpts: &cmd.GenerateOpts{
					GlobalOptions: cmd.GlobalOptions{
						Env: mockEnv,
					},
					Input:  defaultInput,
					Output: defaultOutput,
				},
			},
		}
	})

	Context("the baseimage is not provided", func() {
		It("should return an error", func() {
			err := c.GenerateDocker(memoryFs, mockExporter, mockGit, mockClient)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("baseimage is required"))
		})
	})

	Context("a new input file and baseimage is provided", func() {
		BeforeEach(func() {
			c.Opts.BaseImage = "alpine:latest"
			c.Opts.Input = "ci.dev.yml"
		})

		Context("initial file exists", func() {
			var ciYmlPath string
			JustBeforeEach(func() {
				ciYmlPath = path.Join(c.Opts.RepoRoot, "ci.dev.yml")

				_, err := memoryFs.Create(ciYmlPath)
				Expect(err).To(Not(HaveOccurred()))
			})
			It("should not return an error", func() {
				mockExporter.EXPECT().ReadYmlFile(ciYmlPath).Return(&ci.CiYml{}, nil)
				mockExporter.EXPECT().ExportDockerArtifacts().Return(nil)
				err := c.GenerateDocker(memoryFs, mockExporter, mockGit, mockClient)
				Expect(err).To(Not(HaveOccurred()))
			})
		})
	})

	Context("CloneRepository", func() {
		var (
			clonedir string
			repoUrl  string
			branch   string
		)

		BeforeEach(func() {
			c.Opts.BaseImage = "alpine:latest"
			clonedir = "./test-clone"
			repoUrl = "https://github.com/test/repo.git"
			branch = "main"
		})

		Context("when workspace ID cannot be retrieved", func() {
			It("should return an error", func() {
				mockEnv.EXPECT().GetWorkspaceId().Return(0, errors.New("workspace ID not found"))

				err := c.CloneRepository(mockClient, memoryFs, mockGit, clonedir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to get workspace ID"))
			})
		})

		Context("when workspace ID is set in the env var", func() {
			It("should use the workspace ID from environment", func() {
				envWsId := 99
				mockEnv.EXPECT().GetWorkspaceId().Return(envWsId, nil)

				cmd := &cmd.GenerateDockerCmd{
					Opts: &cmd.GenerateDockerOpts{
						GenerateOpts: &cmd.GenerateOpts{
							GlobalOptions: cmd.GlobalOptions{
								Env: mockEnv,
								// WorkspaceId is nil, so it will use env var
							},
						},
						BaseImage: "alpine:latest",
					},
				}

				ws := api.Workspace{
					GitUrl: *openapi_client.NewNullableString(&repoUrl),
				}
				mockClient.EXPECT().GetWorkspace(envWsId).Return(ws, nil)
				mockGit.EXPECT().CloneRepository(memoryFs, repoUrl, branch, clonedir).Return(nil, nil)

				err := cmd.CloneRepository(mockClient, memoryFs, mockGit, clonedir)
				Expect(err).To(Not(HaveOccurred()))
			})
		})

		Context("when workspace ID is set in the -w flag", func() {
			It("should use the workspace ID from flag", func() {
				flagWsId := 42
				cmd := &cmd.GenerateDockerCmd{
					Opts: &cmd.GenerateDockerOpts{
						GenerateOpts: &cmd.GenerateOpts{
							GlobalOptions: cmd.GlobalOptions{
								Env:         mockEnv,
								WorkspaceId: &flagWsId,
							},
						},
						BaseImage: "alpine:latest",
					},
				}

				ws := api.Workspace{
					GitUrl: *openapi_client.NewNullableString(&repoUrl),
				}
				mockClient.EXPECT().GetWorkspace(flagWsId).Return(ws, nil)
				mockGit.EXPECT().CloneRepository(memoryFs, repoUrl, branch, clonedir).Return(nil, nil)

				err := cmd.CloneRepository(mockClient, memoryFs, mockGit, clonedir)
				Expect(err).To(Not(HaveOccurred()))
			})
		})

		Context("when workspace cannot be retrieved", func() {
			It("should return an error", func() {
				mockEnv.EXPECT().GetWorkspaceId().Return(wsId, nil)
				mockClient.EXPECT().GetWorkspace(wsId).Return(api.Workspace{}, errors.New("workspace not found"))

				err := c.CloneRepository(mockClient, memoryFs, mockGit, clonedir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to get workspace"))
			})
		})

		Context("when workspace has no git repository", func() {
			It("should return an error", func() {
				ws := api.Workspace{}
				mockEnv.EXPECT().GetWorkspaceId().Return(wsId, nil)
				mockClient.EXPECT().GetWorkspace(wsId).Return(ws, nil)

				err := c.CloneRepository(mockClient, memoryFs, mockGit, clonedir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("workspace does not have a git repository"))
			})
		})

		Context("when git clone fails", func() {
			It("should return an error", func() {
				ws := api.Workspace{
					GitUrl: *openapi_client.NewNullableString(&repoUrl),
				}
				mockEnv.EXPECT().GetWorkspaceId().Return(wsId, nil)
				mockClient.EXPECT().GetWorkspace(wsId).Return(ws, nil)
				mockGit.EXPECT().CloneRepository(memoryFs, repoUrl, branch, clonedir).Return(nil, errors.New("clone failed"))

				err := c.CloneRepository(mockClient, memoryFs, mockGit, clonedir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to clone repository"))
			})
		})

		Context("when cloning succeeds with default branch", func() {
			It("should clone the repository", func() {
				ws := api.Workspace{
					GitUrl: *openapi_client.NewNullableString(&repoUrl),
				}
				mockEnv.EXPECT().GetWorkspaceId().Return(wsId, nil)
				mockClient.EXPECT().GetWorkspace(wsId).Return(ws, nil)
				mockGit.EXPECT().CloneRepository(memoryFs, repoUrl, branch, clonedir).Return(nil, nil)

				err := c.CloneRepository(mockClient, memoryFs, mockGit, clonedir)
				Expect(err).To(Not(HaveOccurred()))
			})
		})

		Context("when workspace has initial branch set", func() {
			It("should use the workspace's initial branch", func() {
				customBranch := "develop"
				ws := api.Workspace{
					GitUrl:        *openapi_client.NewNullableString(&repoUrl),
					InitialBranch: *openapi_client.NewNullableString(&customBranch),
				}
				mockEnv.EXPECT().GetWorkspaceId().Return(wsId, nil)
				mockClient.EXPECT().GetWorkspace(wsId).Return(ws, nil)
				mockGit.EXPECT().CloneRepository(memoryFs, repoUrl, customBranch, clonedir).Return(nil, nil)

				err := c.CloneRepository(mockClient, memoryFs, mockGit, clonedir)
				Expect(err).To(Not(HaveOccurred()))
			})
		})

		Context("when branch is explicitly set in options", func() {
			It("should use the explicitly set branch", func() {
				customBranch := "feature-branch"
				workspaceBranch := "develop"
				c.Opts.Branch = customBranch
				ws := api.Workspace{
					GitUrl:        *openapi_client.NewNullableString(&repoUrl),
					InitialBranch: *openapi_client.NewNullableString(&workspaceBranch),
				}
				mockEnv.EXPECT().GetWorkspaceId().Return(wsId, nil)
				mockClient.EXPECT().GetWorkspace(wsId).Return(ws, nil)
				mockGit.EXPECT().CloneRepository(memoryFs, repoUrl, customBranch, clonedir).Return(nil, nil)

				err := c.CloneRepository(mockClient, memoryFs, mockGit, clonedir)
				Expect(err).To(Not(HaveOccurred()))
			})
		})
	})
})
