// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

type PipelinesCmd struct {
	cmd *cobra.Command
}

func addPipelinesCmd(rootCmd *cobra.Command, opts GlobalOptions) {
	s := PipelinesCmd{
		cmd: &cobra.Command{
			Use:   "pipelines",
			Short: "pipelines start/stop",
			Long:  "start and stop pipeline stages",
			Example: `
			`,
		},
	}
	rootCmd.AddCommand(s.cmd)
	addPipelinesStartCmd(s.cmd, opts)
}
