// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"path"

	"github.com/codesphere-cloud/cs-go/cli/cmd"
	"github.com/codesphere-cloud/cs-go/pkg/ci"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/codesphere-cloud/cs-go/pkg/exporter"
)

var _ = Describe("GenerateKubernetes", func() {
	var (
		memoryFs     *cs.FileSystem
		mockEnv      *cmd.MockEnv
		mockExporter *exporter.MockExporter
		c            *cmd.GenerateKubernetesCmd
		wsId         int
		repoRoot     string
	)

	BeforeEach(func() {
		memoryFs = cs.NewMemFileSystem()
		mockEnv = cmd.NewMockEnv(GinkgoT())
		mockExporter = exporter.NewMockExporter(GinkgoT())
		repoRoot = "workspace-repo"

		defaultInput := "ci.yml"
		defaultOutput := "./export"
		wsId = 1
		c = &cmd.GenerateKubernetesCmd{
			Opts: &cmd.GenerateKubernetesOpts{
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
		input := "ci.dev.yml"
		c.Opts.Input = input
		c.Opts.RepoRoot = repoRoot
	})

	Context("The registry is not provided", func() {
		It("should return an error", func() {
			err := c.GenerateKubernetes(memoryFs, mockExporter)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("registry is required"))
		})
	})

	Context("A new input file and registry is provided", func() {
		var ciYmlPath string
		BeforeEach(func() {
			c.Opts.Registry = "my-registry.com"
		})
		Context("Initial file exists", func() {
			JustBeforeEach(func() {
				_, err := memoryFs.Create("ci.dev.yml")
				Expect(err).To(Not(HaveOccurred()))
			})
			It("should not return an error", func() {
				ciYmlPath = path.Join(c.Opts.RepoRoot, "ci.dev.yml")
				mockExporter.EXPECT().ReadYmlFile(ciYmlPath).Return(&ci.CiYml{}, nil)
				mockExporter.EXPECT().ExportKubernetesArtifacts("my-registry.com", "", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				err := c.GenerateKubernetes(memoryFs, mockExporter)
				Expect(err).To(Not(HaveOccurred()))
			})
		})
	})
})
