// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package start

import (
	shared "github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/spf13/cobra"
)

type StartCmd struct {
	cmd *cobra.Command
}

func AddStartCmd(rootCmd *cobra.Command, opts shared.RootOptions) {
	start := StartCmd{
		cmd: &cobra.Command{
			Use:   "start",
			Short: "Start workspace pipeline",
			Long:  `Start pipeline of a workspace using the pipeline subcommand`,
		},
	}
	shared.AddCmd(rootCmd, start.cmd)
	AddStartPipelineCmd(start.cmd, opts)
}
