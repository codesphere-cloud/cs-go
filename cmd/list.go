/*
Copyright © 2025 Codesphere Inc. <support@codesphere.com>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

type ListCmd struct {
	cmd *cobra.Command
}

func addListCmd(rootCmd *cobra.Command, opts GlobalOptions) {
	l := ListCmd{
		cmd: &cobra.Command{
			Use:   "list",
			Short: "list resources",
			Long:  `list resources available in Codesphere`,
			Example: `
				List all workspaces:

			  $ cs list workspaces
			`,
		},
	}
	rootCmd.AddCommand(l.cmd)
	addListWorkspacesCmd(l.cmd, opts)
	addListTeamsCmd(l.cmd, opts)
}
