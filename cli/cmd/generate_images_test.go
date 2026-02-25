// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"context"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/codesphere-cloud/cs-go/cli/cmd"
	"github.com/codesphere-cloud/cs-go/pkg/ci"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/codesphere-cloud/cs-go/pkg/exporter"
)

var _ = Describe("GenerateImages", func() {
	var (
		memoryFs     *cs.FileSystem
		mockEnv      *cmd.MockEnv
		mockExporter *exporter.MockExporter
		c            *cmd.GenerateImagesCmd
		wsId         int
		repoRoot     string
	)

	BeforeEach(func() {
		input := "ci.dev.yml"

		memoryFs = cs.NewMemFileSystem()
		mockEnv = cmd.NewMockEnv(GinkgoT())
		mockExporter = exporter.NewMockExporter(GinkgoT())

		repoRoot = "fake-root"
		defaultInput := "ci.yml"
		defaultOutput := "./export"
		wsId = 1
		c = &cmd.GenerateImagesCmd{
			Opts: &cmd.GenerateImagesOpts{
				GenerateOpts: &cmd.GenerateOpts{
					GlobalOptions: cmd.GlobalOptions{
						Env:         mockEnv,
						WorkspaceId: wsId,
					},
					Input:  defaultInput,
					Output: defaultOutput,
				},
			},
		}

		c.Opts.Input = input
		c.Opts.RepoRoot = repoRoot

	})

	Context("The registry is not provided", func() {
		It("should return an error", func() {
			err := c.GenerateImages(memoryFs, mockExporter)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("registry is required"))
		})
	})

	Context("A new input file and registry is provided", func() {
		BeforeEach(func() {
			registry := "my-registry.com"
			c.Opts.Registry = registry
		})

		Context("Initial file exists", func() {
			JustBeforeEach(func() {
				_, err := memoryFs.Create("ci.dev.yml")
				Expect(err).To(Not(HaveOccurred()))
			})
			It("should not return an error", func() {
				ciYmlPath := path.Join(c.Opts.RepoRoot, "ci.dev.yml")
				mockExporter.EXPECT().ReadYmlFile(ciYmlPath).Return(&ci.CiYml{}, nil)
				mockExporter.EXPECT().ExportImages(context.Background(), "my-registry.com", "").Return(nil)
				err := c.GenerateImages(memoryFs, mockExporter)
				Expect(err).To(Not(HaveOccurred()))
			})
		})
	})
})
