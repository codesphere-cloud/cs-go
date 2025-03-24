/*
Copyright © 2025 Codesphere Inc. <support@codesphere.com>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func Execute() {
	var rootCmd = &cobra.Command{
		Use:   "cs",
		Short: "The codesphere CLI",
		Long: `Manage and debug resources deployed in Codesphere
	via command line.`,
	}

	addLogCmd(rootCmd)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
