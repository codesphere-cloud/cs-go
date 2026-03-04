// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

type ScaleCmd struct {
	cmd *cobra.Command
}

func AddScaleCmd(rootCmd *cobra.Command, opts *GlobalOptions) {
	scale := ScaleCmd{
		cmd: &cobra.Command{
			Use:   "scale",
			Short: "Scale Codesphere resources",
			Long:  `Scale Codesphere resources, like landscape services of a workspace.`,
		},
	}
	rootCmd.AddCommand(scale.cmd)

	AddScaleWorkspaceCmd(scale.cmd, opts)
}
