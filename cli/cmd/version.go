/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/spf13/cobra"
)

type VersionCmd struct {
	cmd *cobra.Command
}

func (c *VersionCmd) RunE(_ *cobra.Command, args []string) error {
	log.Printf("Codesphere CLI version: %s\n", cs.Version())
	log.Printf("Commit: %s\n", cs.Commit())
	log.Printf("Build Date: %s\n", cs.BuildDate())

	return nil
}

func AddVersionCmd(rootCmd *cobra.Command) {
	version := VersionCmd{
		cmd: &cobra.Command{
			Use:   "version",
			Short: "Print version",
			Long:  `Print current version of Codesphere CLI.`,
		},
	}
	rootCmd.AddCommand(version.cmd)
	version.cmd.RunE = version.RunE
}
