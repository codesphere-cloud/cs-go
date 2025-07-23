package cmd

import (
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type GitCmd struct {
	cmd *cobra.Command
}

func AddGitCmd(rootCmd *cobra.Command, opts GlobalOptions) {
	git := GitCmd{
		cmd: &cobra.Command{
			Use:   "git",
			Short: "Interacting with the git repository of the workspace",
			Long: io.Long(`Interact with the git repository of the workspace

				Run git commands inside the workspace,
				like pulling or switching to a specific branch.`),
		},
	}
	rootCmd.AddCommand(git.cmd)

	// Add child commands here
	AddGitPullCmd(git.cmd, opts)
}
