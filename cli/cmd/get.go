// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

func AddGetCmd(rootCmd *cobra.Command, opts GlobalOptions) {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get Codesphere resources",
		Long:  `Get information about specific Codesphere resources`,
		Example: io.FormatExampleCommands("get", []io.Example{
			{Cmd: "domains", Desc: "Get domains for a workspace"},
		}),
	}
	rootCmd.AddCommand(getCmd)

	addGetDomainsCmd(getCmd, opts)
}
