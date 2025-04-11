/*
Copyright © 2025 Codesphere Inc. <support@codesphere.com>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

type GlobalOptions struct {
	apiUrl *string
}

func (o GlobalOptions) GetApiUrl() string {
	if o.apiUrl != nil {
		return *o.apiUrl
	}
	return GetApiUrl()
}

func Execute() {
	var rootCmd = &cobra.Command{
		Use:   "cs",
		Short: "The codesphere CLI",
		Long: `Manage and debug resources deployed in Codesphere
	via command line.`,
	}

	opts := GlobalOptions{}

	addLogCmd(rootCmd, opts)
	addListCmd(rootCmd, opts)

	opts.apiUrl = rootCmd.PersistentFlags().StringP("api", "a", "", "URL of Codesphere API (can also be CS_API)")

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
