// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"path"

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
						Env:         mockEnv,
						WorkspaceId: &wsId,
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
})
