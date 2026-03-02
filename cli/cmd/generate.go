// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type GenerateCmd struct {
	cmd  *cobra.Command
	Opts *GenerateOpts
}

type GenerateOpts struct {
	*GlobalOptions
	Input    string
	Branch   string
	Output   string
	Force    bool
	RepoRoot string
}

func AddGenerateCmd(rootCmd *cobra.Command, opts *GlobalOptions) {
	generate := GenerateCmd{
		cmd: &cobra.Command{
			Use:   "generate",
			Short: "Generate codesphere artifacts",
			Long: io.Long(`Collection of commands to generate codesphere related artifacts, such as dockerfiles based on a specific workspace.
			If the input file is not found, cs will attempt to clone a branch (default is 'main') of the repository of the workspace
			on your local machine to run the artifact generation.`),
		},
		Opts: &GenerateOpts{GlobalOptions: opts},
	}
	generate.cmd.PersistentFlags().StringVarP(&generate.Opts.Input, "input", "i", "ci.yml", "CI profile to use as input for generation, relative to repository root")
	generate.cmd.PersistentFlags().StringVar(&generate.Opts.Branch, "branch", "main", "Branch of the repository to clone if the input file is not found")
	generate.cmd.PersistentFlags().StringVarP(&generate.Opts.Output, "output", "o", "export", "Output path of the folder including generated artifacts, relative to repository root")
	generate.cmd.PersistentFlags().BoolVarP(&generate.Opts.Force, "force", "f", false, "Overwrite any files if existing")
	generate.cmd.PersistentFlags().StringVar(&generate.Opts.RepoRoot, "reporoot", "./workspace-repo", "root directory of the workspace repository to export. Will be used to clone the repository if it doesn't exist.")

	rootCmd.AddCommand(generate.cmd)

	AddGenerateDockerCmd(generate.cmd, generate.Opts)
	AddGenerateKubernetesCmd(generate.cmd, generate.Opts)
	AddGenerateImagesCmd(generate.cmd, generate.Opts)
}
