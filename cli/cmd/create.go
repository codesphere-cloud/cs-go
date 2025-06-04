package cmd

import (
	"github.com/spf13/cobra"
)

// CreateCmd represents the create command
type CreateCmd struct {
	cmd *cobra.Command
}

func AddCreateCmd(rootCmd *cobra.Command, opts GlobalOptions) {
	create := CreateCmd{
		cmd: &cobra.Command{
			Use:   "create",
			Short: "Create codesphere resource",
			Long:  `Create codesphere resources like workspaces.`,
		},
	}
	rootCmd.AddCommand(create.cmd)

	// Add child commands here
	AddCreateWorkspaceCmd(create.cmd, opts)
}
