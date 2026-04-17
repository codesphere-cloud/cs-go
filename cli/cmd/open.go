// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/codesphere-cloud/cs-go/pkg/cs"
)

type OpenCmd struct {
	cmd  *cobra.Command
	Opts *GlobalOptions
}

func (c *OpenCmd) RunE(_ *cobra.Command, args []string) error {
	log.Println("Opening Codesphere IDE")
	return cs.NewBrowser().OpenIde("", c.Opts.StateFile)
}

func AddOpenCmd(rootCmd *cobra.Command, opts *GlobalOptions) {
	open := OpenCmd{
		cmd: &cobra.Command{
			Use:   "open",
			Short: "Open the Codesphere IDE",
			Long:  `Open the Codesphere IDE.`,
		},
		Opts: opts,
	}
	rootCmd.AddCommand(open.cmd)
	open.cmd.RunE = open.RunE
	AddOpenWorkspaceCmd(open.cmd, opts)
}
