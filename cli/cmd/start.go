// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

type StartCmd struct {
	cmd *cobra.Command
}

func AddStartCmd(rootCmd *cobra.Command, opts GlobalOptions) {
	start := StartCmd{
		cmd: &cobra.Command{
			Use:   "start",
			Short: "Start workspace pipeline",
			Long:  `Start pipeline of a workspace using the pipeline subcommand`,
		},
	}
	rootCmd.AddCommand(start.cmd)
	AddStartPipelineCmd(start.cmd, opts)
}
