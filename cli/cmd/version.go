/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/spf13/cobra"
)

// VersionCmd represents the version command
type VersionCmd struct {
	cmd *cobra.Command
}

func (c *VersionCmd) RunE(_ *cobra.Command, args []string) error {
	fmt.Printf("Codesphere CLI version: %s\n", cs.Version())
	fmt.Printf("Commit: %s\n", cs.Commit())
	fmt.Printf("Build Date: %s\n", cs.BuildDate())

	return nil
}

func AddVersionCmd(rootCmd *cobra.Command) {
	version := VersionCmd{
		cmd: &cobra.Command{
			Use:   "version",
			Short: "A brief description of your command",
			Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		},
	}
	rootCmd.AddCommand(version.cmd)
	version.cmd.RunE = version.RunE
}
