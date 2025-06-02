/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// GenerateCmd represents the generate command
type GenerateCmd struct {
	cmd *cobra.Command
}

func AddGenerateCmd(rootCmd *cobra.Command) {
	generate := GenerateCmd{
		cmd: &cobra.Command{
			Use:   "generate",
			Short: "Generate codesphere artifacts",
			Long:  `Collection of commands to generate codesphere related artifacts, such as dockerfiles based on a specific workspace.`,
		},
	}
	rootCmd.AddCommand(generate.cmd)

	AddGenerateDockerfileCmd(generate.cmd)
}
