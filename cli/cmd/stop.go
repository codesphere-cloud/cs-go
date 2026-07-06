// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

type StopCmd struct {
	cmd *cobra.Command
}

func AddStopCmd(rootCmd *cobra.Command, opts *GlobalOptions) {
	stop := StopCmd{
		cmd: &cobra.Command{
			Use:   "stop",
			Short: "Stop workspace pipeline",
			Long:  `Stop pipeline of a workspace using the pipeline subcommand`,
		},
	}
	AddCmd(rootCmd, stop.cmd)
	AddStopPipelineCmd(stop.cmd, opts)
}
