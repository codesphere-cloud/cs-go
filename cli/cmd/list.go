// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/codesphere-cloud/cs-go/pkg/out"
	"github.com/spf13/cobra"
)

type ListCmd struct {
	cmd *cobra.Command
}

func AddListCmd(rootCmd *cobra.Command, opts GlobalOptions) {
	l := ListCmd{
		cmd: &cobra.Command{
			Use:   "list",
			Short: "list resources",
			Long:  `list resources available in Codesphere`,
			Example: out.FormatExampleCommands("list", []out.Example{
				{Cmd: "workspaces", Desc: "List all workspaces"},
			}),
		},
	}
	rootCmd.AddCommand(l.cmd)
	addListWorkspacesCmd(l.cmd, opts)
	addListTeamsCmd(l.cmd, opts)
}
