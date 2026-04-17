// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cs_test

import (
	"fmt"
	"os"
	"time"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/codesphere-cloud/cs-go/pkg/git"
	"github.com/codesphere-cloud/cs-go/pkg/testutil"
	"github.com/codesphere-cloud/cs-go/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var (
	fakeTime *api.MockTime
)
var _ = Describe("cs up deployment", func() {
	var (
		state      *cs.UpState
		ask        bool
		token      string
		fs         *util.FileSystem
		mockClient *api.MockClient
		mockGit    *git.MockGit
		mockTime   *api.MockTime
		devDomain  string
		wsName     string
		branchName string
		remote     string
	)
	BeforeEach(func() {
		mockClient = api.NewMockClient(GinkgoT())
		mockGit = git.NewMockGit(GinkgoT())
		mockTime = testutil.MockTime()
		fs = util.NewMemFileSystem()
		ask = false
		token = "fake-token"
		fakeTime = testutil.MockTimeAt(time.Date(2026, 01, 01, 11, 00, 00, 00, time.UTC))
		devDomain = "cs-up-20260101110000-cs-dev.codesphere.com"
		wsName = "cs-up-20260101110000"
		branchName = wsName
		remote = "origin"

		// Default state for tests, can be overridden in specific contexts
		state = &cs.UpState{
			TeamId:     30,
			RepoAccess: cs.PublicRepo,
			DomainType: cs.PublicDevDomain,
			Env:        []string{},
			Remote:     remote,
			Plan:       -1,
			Profile:    "",
		}

		// initialize state with defaults and save to file. This is done by the cmd before calling cs.Up
		err := state.Load(".cs-up.yaml", fakeTime, fs)
		Expect(err).ToNot(HaveOccurred())
		err = state.Save()
		Expect(err).ToNot(HaveOccurred())
	})
	AfterEach(func() {
		mockClient.AssertExpectations(GinkgoT())
		mockGit.AssertExpectations(GinkgoT())
		fakeTime.AssertExpectations(GinkgoT())
	})
	Context("Up", func() {
		JustBeforeEach(func() {
			mockGit.EXPECT().Checkout(branchName, false).Return(nil)
			mockGit.EXPECT().AddAll().Return(nil)
			mockGit.EXPECT().HasChanges("origin", branchName).Return(false, nil)

			mockClient.EXPECT().GitPull(42, "origin", branchName).Return(nil)

			mockClient.EXPECT().StartPipelineStage(42, "ci.yml", "prepare").Return(nil)
			mockClient.EXPECT().GetPipelineState(42, "prepare").Return([]api.PipelineStatus{
				{Server: "codesphere-ide", State: "success", Replica: "1"},
				{Server: "backend", State: "success", Replica: "1"},
			}, nil)

			mockClient.EXPECT().StartPipelineStage(42, "ci.yml", "test").Return(nil)
			mockClient.EXPECT().GetPipelineState(42, "test").Return([]api.PipelineStatus{
				{Server: "codesphere-ide", State: "success", Replica: "1"},
				{Server: "backend", State: "success", Replica: "1"},
			}, nil)

			mockClient.EXPECT().StartPipelineStage(42, "ci.yml", "run").Return(nil)
			mockClient.EXPECT().GetPipelineState(42, "run").Return([]api.PipelineStatus{
				{Server: "codesphere-ide", State: "success", Replica: "1"},
				{Server: "backend", State: "running", Replica: "1"},
			}, nil)

			mockClient.EXPECT().GetWorkspace(42).Return(api.Workspace{
				Id:        42,
				Name:      wsName,
				DevDomain: &devDomain,
			}, nil)
		})
		Context("default parameters", func() {
			Context("when the workspace doesn't exist", func() {
				It("creates a new workspace", func() {
					mockGit.EXPECT().GetRemoteUrl("origin").Return("https://myrepo.git", nil)

					mockClient.EXPECT().ListWorkspacePlans().Return([]api.WorkspacePlan{{Id: 8}}, nil)
					mockClient.EXPECT().DeployWorkspace(mock.Anything).Return(&api.Workspace{
						Id:   42,
						Name: wsName,
					}, nil)
					mockClient.EXPECT().DeployLandscape(42, "ci.yml").Return(nil)

					err := cs.Up(mockClient, mockGit, mockTime, fs, state, token, ask, false)
					Expect(err).To(Not(HaveOccurred()))

					validateState(fs, &cs.UpState{
						WorkspaceId:   42,
						WorkspaceName: wsName,
						Profile:       "ci.yml",
						Timeout:       state.Timeout,
						Branch:        wsName,
						TeamId:        30,
						Plan:          8,
						BaseImage:     "",
						Env:           []string{},
						DomainType:    cs.PublicDevDomain,
						RepoAccess:    cs.PublicRepo,
						Remote:        remote,
						StateFile:     ".cs-up.yaml",
					})
				})
			})
			Context("when the workspace already exists", func() {
				JustBeforeEach(func() {
					state.Plan = 8
					state.WorkspaceId = 42
					state.WorkspaceName = wsName
					state.Branch = branchName
					state.Profile = "ci.yml"
					err := state.Save()
					Expect(err).ToNot(HaveOccurred())

					mockClient.EXPECT().GetWorkspace(42).Return(api.Workspace{
						Id:   42,
						Name: wsName,
					}, nil)

					// Workspace exists, so wake it up instead of creating new
					mockClient.EXPECT().WakeUpWorkspace(42, token, "ci.yml", state.Timeout).Return(nil)
					mockClient.EXPECT().SetEnvVarOnWorkspace(42, mock.Anything).Return(nil)
				})
				It("reuses the existing workspace", func() {
					mockClient.EXPECT().DeployLandscape(42, "ci.yml").Return(nil)
					err := cs.Up(mockClient, mockGit, mockTime, fs, state, token, ask, false)
					Expect(err).To(Not(HaveOccurred()))

					validateState(fs, &cs.UpState{
						WorkspaceId:   42,
						WorkspaceName: wsName,
						Profile:       "ci.yml",
						Timeout:       state.Timeout,
						Branch:        wsName,
						TeamId:        30,
						Plan:          8,
						BaseImage:     "",
						Env:           []string{},
						DomainType:    cs.PublicDevDomain,
						RepoAccess:    cs.PublicRepo,
						Remote:        remote,
						StateFile:     ".cs-up.yaml",
					})
				})
				Context("when the branch name provided", func() {
					BeforeEach(func() {
						branchName = "my-feature-branch"
					})
					It("uses the provided branch name instead of the generated one", func() {
						mockClient.EXPECT().DeployLandscape(42, "ci.yml").Return(nil)

						err := cs.Up(mockClient, mockGit, mockTime, fs, state, token, ask, false)
						Expect(err).To(Not(HaveOccurred()))

						validateState(fs, &cs.UpState{
							WorkspaceId:   42,
							WorkspaceName: wsName,
							Profile:       "ci.yml",
							Timeout:       state.Timeout,
							Branch:        branchName,
							TeamId:        30,
							Plan:          8,
							BaseImage:     "",
							Env:           []string{},
							DomainType:    cs.PublicDevDomain,
							RepoAccess:    cs.PublicRepo,
							Remote:        remote,
							StateFile:     ".cs-up.yaml",
						})
					})

				})
				Context("when the deployment request fails the first time", func() {
					It("should retry waking up the workspace", func() {
						mockClient.EXPECT().DeployLandscape(42, "ci.yml").Return(fmt.Errorf("deployment error 500")).Once()
						mockClient.EXPECT().DeployLandscape(42, "ci.yml").Return(nil).Once()
						err := cs.Up(mockClient, mockGit, mockTime, fs, state, token, ask, false)
						Expect(err).To(Not(HaveOccurred()))

						validateState(fs, &cs.UpState{
							WorkspaceId:   42,
							WorkspaceName: wsName,
							Profile:       "ci.yml",
							Timeout:       state.Timeout,
							Branch:        branchName,
							TeamId:        30,
							Plan:          8,
							BaseImage:     "",
							Env:           []string{},
							DomainType:    cs.PublicDevDomain,
							RepoAccess:    cs.PublicRepo,
							Remote:        remote,
							StateFile:     ".cs-up.yaml",
						})

						Expect(err).ToNot(HaveOccurred())
					})

				})
			})
		})
	})

	Describe("CodesphereDeploymentManager", func() {
		var (
			mgr *cs.CodesphereDeploymentManager
		)
		JustBeforeEach(func() {
			mgr = &cs.CodesphereDeploymentManager{
				Client:          mockClient,
				GitSvc:          mockGit,
				FileSys:         fs,
				State:           state,
				Verbose:         false,
				AskConfirmation: ask,
				ApiToken:        token,
			}
		})

		Describe("UpdateGitIgnore", func() {
			It("should add .cs-up.yaml to .gitignore if it doesn't exist", func() {
				err := mgr.UpdateGitIgnore()
				Expect(err).ToNot(HaveOccurred())

				// Verify that .cs-up.yaml is added to .gitignore
				gitIgnoreFile, err := fs.OpenFile(".gitignore", os.O_RDWR|os.O_CREATE, 0644)
				Expect(err).ToNot(HaveOccurred())
				content, err := fs.ReadFile(".gitignore")
				Expect(err).ToNot(HaveOccurred())
				_ = gitIgnoreFile.Close()

				Expect(string(content)).To(ContainSubstring(".cs-up.yaml"))

				Expect(err).ToNot(HaveOccurred())
			})
			It("should not add .cs-up.yaml to .gitignore if it already exists", func() {
				// Simulate .cs-up.yaml already being in .gitignore
				content := []byte("other\nstuff\n.cs-up.yaml\nignored\n")
				_, err := fs.CreateFile(".gitignore")
				Expect(err).ToNot(HaveOccurred())
				gitIgnoreFile, err := fs.OpenFile(".gitignore", os.O_RDWR|os.O_CREATE, 0644)
				Expect(err).ToNot(HaveOccurred())
				_, err = gitIgnoreFile.Write(content)
				Expect(err).ToNot(HaveOccurred())
				_ = gitIgnoreFile.Close()

				err = mgr.UpdateGitIgnore()
				Expect(err).ToNot(HaveOccurred())

				// Verify that .cs-up.yaml is not duplicated in .gitignore
				gitIgnoreFile, err = fs.OpenFile(".gitignore", os.O_RDONLY, 0644)
				Expect(err).ToNot(HaveOccurred())
				newContent, err := fs.ReadFile(".gitignore")
				Expect(err).ToNot(HaveOccurred())
				_ = gitIgnoreFile.Close()

				Expect(newContent).To(Equal(content))
			})
		})

		Describe("PushChanges", func() {
			JustBeforeEach(func() {
				state.WorkspaceId = 42
				state.Branch = branchName

				err := state.Save()
				Expect(err).ToNot(HaveOccurred())

				mockGit.EXPECT().Checkout(branchName, false).Return(nil)
				mockGit.EXPECT().AddAll().Return(nil)
			})
			Context("when there are changes to push", func() {
				It("should push changes to the correct branch", func() {
					mockGit.EXPECT().HasChanges(remote, branchName).Return(true, nil)
					mockGit.EXPECT().Commit("cs up commit").Return(nil)
					mockGit.EXPECT().Push(remote, branchName).Return(nil)
					mockGit.EXPECT().PrintStatus().Return(nil)

					err := mgr.PushChanges(false)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("when there are no changes to push", func() {
				It("should not attempt to push", func() {
					mockGit.EXPECT().HasChanges(remote, branchName).Return(false, nil)

					err := mgr.PushChanges(false)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Describe("EnsureWorkspace", func() {
			Context("when the workspace exists", func() {
				It("should wake up the existing workspace", func() {
					state.WorkspaceId = 42
					err := state.Save()
					Expect(err).ToNot(HaveOccurred())

					mockClient.EXPECT().GetWorkspace(42).Return(api.Workspace{
						Id:   42,
						Name: wsName,
					}, nil)

					mockClient.EXPECT().WakeUpWorkspace(42, token, "ci.yml", state.Timeout).Return(nil)
					mockClient.EXPECT().SetEnvVarOnWorkspace(42, mock.Anything).Return(nil)

					err = mgr.EnsureWorkspace()
					Expect(err).ToNot(HaveOccurred())
				})

			})

			Context("when the workspace does not exist", func() {
				It("should create a new workspace", func() {
					mockGit.EXPECT().GetRemoteUrl("origin").Return("https://myrepo.git", nil)

					mockClient.EXPECT().ListWorkspacePlans().Return([]api.WorkspacePlan{{Id: 8}}, nil)
					mockClient.EXPECT().DeployWorkspace(mock.Anything).Return(&api.Workspace{
						Id:   42,
						Name: wsName,
					}, nil)

					err := mgr.EnsureWorkspace()
					Expect(err).ToNot(HaveOccurred())
				})

				Context("when the domain type is private", func() {
					BeforeEach(func() {
						state.DomainType = cs.PrivateDevDomain
					})
					It("should create a new workspace with the correct domain type", func() {
						mockGit.EXPECT().GetRemoteUrl("origin").Return("https://myrepo.git", nil)

						mockClient.EXPECT().ListWorkspacePlans().Return([]api.WorkspacePlan{{Id: 8}}, nil)
						mockClient.EXPECT().DeployWorkspace(mock.MatchedBy(func(args api.DeployWorkspaceArgs) bool {
							return *args.Restricted == true
						})).Return(&api.Workspace{
							Id:   42,
							Name: wsName,
						}, nil)

						err := mgr.EnsureWorkspace()
						Expect(err).ToNot(HaveOccurred())
					})
				})
				Context("when the git repo is private", func() {
					BeforeEach(func() {
						state.RepoAccess = cs.PrivateRepo
					})
					It("should create a new workspace with the correct domain type", func() {
						mockGit.EXPECT().GetRemoteUrl("origin").Return("https://myrepo.git", nil)

						mockClient.EXPECT().ListWorkspacePlans().Return([]api.WorkspacePlan{{Id: 8}}, nil)
						mockClient.EXPECT().DeployWorkspace(mock.MatchedBy(func(args api.DeployWorkspaceArgs) bool {
							return args.IsPrivateRepo == true
						})).Return(&api.Workspace{
							Id:   42,
							Name: wsName,
						}, nil)

						err := mgr.EnsureWorkspace()
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})

		})
	})
})

func validateState(fs *util.FileSystem, expected *cs.UpState) {
	actual := &cs.UpState{}
	err := actual.Load(".cs-up.yaml", fakeTime, fs)
	Expect(err).ToNot(HaveOccurred())

	Expect(actual.WorkspaceId).To(Equal(expected.WorkspaceId))
	Expect(actual.WorkspaceName).To(Equal(expected.WorkspaceName))
	Expect(actual.Profile).To(Equal(expected.Profile))
	Expect(actual.Timeout).To(Equal(expected.Timeout))
	Expect(actual.Branch).To(Equal(expected.Branch))
	Expect(actual.TeamId).To(Equal(expected.TeamId))
	Expect(actual.Plan).To(Equal(expected.Plan))
	Expect(actual.BaseImage).To(Equal(expected.BaseImage))
	Expect(actual.Env).To(Equal(expected.Env))
	Expect(actual.DomainType).To(Equal(expected.DomainType))
	Expect(actual.RepoAccess).To(Equal(expected.RepoAccess))
	Expect(actual.Remote).To(Equal(expected.Remote))
}
