package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/codesphere-cloud/cs-go/pkg/cs"
)

// OpenCmd represents the open command
type OpenCmd struct {
	cmd *cobra.Command
}

func (c *OpenCmd) RunE(_ *cobra.Command, args []string) error {
	//Command execution goes here

	fmt.Println("Opening Codesphere IDE")
	return cs.NewBrowser().OpenIde("")
}

func AddOpenCmd(rootCmd *cobra.Command, opts GlobalOptions) {
	open := OpenCmd{
		cmd: &cobra.Command{
			Use:   "open",
			Short: "Open the codesphere IDE",
			Long:  `Open the codesphere IDE.`,
		},
	}
	rootCmd.AddCommand(open.cmd)
	open.cmd.RunE = open.RunE
	AddOpenWorkspaceCmd(open.cmd, opts)
}
