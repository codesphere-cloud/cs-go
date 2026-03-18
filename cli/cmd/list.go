// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type OutputFormat string

const (
	OutputFormatTable OutputFormat = "table"
	OutputFormatJSON  OutputFormat = "json"
	OutputFormatYAML  OutputFormat = "yaml"
)

type ListOptions struct {
	*GlobalOptions
	OutputFormat OutputFormat
}

type ListCmd struct {
	cmd *cobra.Command
}

func AddListCmd(rootCmd *cobra.Command, opts *GlobalOptions) {
	l := ListCmd{
		cmd: &cobra.Command{
			Use:   "list",
			Short: "List resources",
			Long:  `List resources available in Codesphere`,
			Example: io.FormatExampleCommands("list", []io.Example{
				{Cmd: "workspaces", Desc: "List all workspaces"},
			}),
		},
	}

	listOpts := &ListOptions{GlobalOptions: opts}
	l.cmd.PersistentFlags().StringVarP((*string)(&listOpts.OutputFormat), "output", "o", "table", "Output format (table, json, yaml)")
	l.cmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		if listOpts.OutputFormat != OutputFormatTable && listOpts.OutputFormat != OutputFormatJSON && listOpts.OutputFormat != OutputFormatYAML {
			return fmt.Errorf("invalid output format: %s", listOpts.OutputFormat)
		}
		return nil
	}

	rootCmd.AddCommand(l.cmd)
	addListWorkspacesCmd(l.cmd, listOpts)
	AddListBaseimagesCmd(l.cmd, listOpts)
	addListTeamsCmd(l.cmd, listOpts)
	AddListPlansCmd(l.cmd, listOpts)
}
