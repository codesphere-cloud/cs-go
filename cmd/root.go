// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"

	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/spf13/cobra"
)

type GlobalOptions struct {
	ApiUrl *string
}

func (o GlobalOptions) GetApiUrl() string {
	if o.ApiUrl != nil {
		return *o.ApiUrl
	}
	return cs.GetApiUrl()
}

func Execute() {
	var rootCmd = &cobra.Command{
		Use:   "cs",
		Short: "The codesphere CLI",
		Long:  `Manage and debug resources deployed in Codesphere via command line.`,
	}

	opts := GlobalOptions{}

	addLogCmd(rootCmd, opts)
	addListCmd(rootCmd, opts)
	addPipelinesCmd(rootCmd, opts)

	opts.ApiUrl = rootCmd.PersistentFlags().StringP("api", "a", "https://codesphere.com/api", "URL of Codesphere API (can also be CS_API)")

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
