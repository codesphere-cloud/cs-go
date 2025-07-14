// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/api/openapi_client"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/codesphere-cloud/cs-go/pkg/export"
	"github.com/codesphere-cloud/cs-go/pkg/git"
)

var _ = Describe("GenerateDockerfile", func() {
	var (
		memoryFs     *cs.FileSystem
		mockClient   *cmd.MockClient
		mockEnv      *cmd.MockEnv
		mockExporter *export.MockExporter
		mockGit      *git.MockGit
		c            *cmd.GenerateDockerfileCmd
		wsId         int
		repoStr      string
		ws           api.Workspace
	)

	BeforeEach(func() {
		mockClient = cmd.NewMockClient(GinkgoT())
		memoryFs = cs.NewMemFileSystem()
		mockEnv = cmd.NewMockEnv(GinkgoT())
		mockExporter = export.NewMockExporter(GinkgoT())
		mockGit = git.NewMockGit(GinkgoT())

		defaultInput := "ci.yml"
		defaultOutput := "./export"
		wsId = 1
		repoStr = "https://fake-git.com/my/repo.git"
		c = &cmd.GenerateDockerfileCmd{
			Opts: cmd.GenerateDockerfileOpts{
				GlobalOptions: cmd.GlobalOptions{
					Env:         mockEnv,
					WorkspaceId: &wsId,
				},
				Input:  &defaultInput,
				Output: &defaultOutput,
			},
		}
	})

	Context("the baseimage is not provided", func() {
		It("should return an error", func() {
			err := c.GenerateDockerfile(mockClient, memoryFs, mockExporter, mockGit)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("baseimage is required"))
		})
	})

	Context("a new input file and baseimage is provided", func() {
		BeforeEach(func() {
			input := "ci.dev.yml"
			baseImage := "alpine:latest"
			c.Opts.BaseImage = &baseImage
			c.Opts.Input = &input
		})

		Context("initial file exists", func() {
			JustBeforeEach(func() {
				_, err := memoryFs.Create("ci.dev.yml")
				Expect(err).To(Not(HaveOccurred()))
			})
			It("should not return an error", func() {
				mockExporter.EXPECT().ExportDockerArtifacts("ci.dev.yml", "./export", "alpine:latest", []string{}).Return(nil)
				err := c.GenerateDockerfile(mockClient, memoryFs, mockExporter, mockGit)
				Expect(err).To(Not(HaveOccurred()))
			})
		})

		Context("the workspace does not have a git repo", func() {
			JustBeforeEach(func() {
				emptyRepoStr := ""
				ws = api.Workspace{
					Id:     wsId,
					GitUrl: *openapi_client.NewNullableString(&emptyRepoStr),
				}
			})
			It("should return an error", func() {
				mockClient.EXPECT().GetWorkspace(wsId).Return(ws, nil)
				err := c.GenerateDockerfile(mockClient, memoryFs, mockExporter, mockGit)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("workspace does not have a git repository"))
			})
		})

		Context("the workspace has a git repo but repo does not contain input file", func() {
			JustBeforeEach(func() {
				ws = api.Workspace{
					Id:     wsId,
					GitUrl: *openapi_client.NewNullableString(&repoStr),
				}
			})
			It("should return an error", func() {
				mockClient.EXPECT().GetWorkspace(wsId).Return(ws, nil)
				mockGit.EXPECT().CloneRepository(memoryFs, repoStr, "main").Return(nil, nil)
				err := c.GenerateDockerfile(mockClient, memoryFs, mockExporter, mockGit)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("input file ci.dev.yml not found after cloning repository"))
			})
		})
	})
})
