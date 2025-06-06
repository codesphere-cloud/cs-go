package cmd

import (
	"github.com/spf13/cobra"
)

type DeleteCmd struct {
	cmd *cobra.Command
}

func AddDeleteCmd(rootCmd *cobra.Command, opt GlobalOptions) {
	delete := DeleteCmd{
		cmd: &cobra.Command{
			Use:   "delete",
			Short: "Delete Codesphere resources",
			Long:  `Delete Codesphere resources, e.g. workspaces.`,
		},
	}
	rootCmd.AddCommand(delete.cmd)

	// Add child commands here
	AddDeleteWorkspaceCmd(delete.cmd, opt)
}
