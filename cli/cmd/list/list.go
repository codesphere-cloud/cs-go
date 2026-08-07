// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"

	shared "github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	shared.RootOptions
	OutputFormat shared.OutputFormat
}

type ListCmd struct {
	cmd *cobra.Command
}

func AddListCmd(rootCmd *cobra.Command, opts shared.RootOptions) {
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

	listOpts := &ListOptions{RootOptions: opts}
	l.cmd.PersistentFlags().StringVarP((*string)(&listOpts.OutputFormat), "output", "o", "table", "Output format (table, json, yaml)")
	l.cmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		if listOpts.OutputFormat != shared.OutputFormatTable && listOpts.OutputFormat != shared.OutputFormatJSON && listOpts.OutputFormat != shared.OutputFormatYAML {
			return fmt.Errorf("invalid output format: %s", listOpts.OutputFormat)
		}
		return nil
	}

	shared.AddCmd(rootCmd, l.cmd)
	addListWorkspacesCmd(l.cmd, listOpts)
	AddListBaseimagesCmd(l.cmd, listOpts)
	addListTeamsCmd(l.cmd, listOpts)
	AddListOrgCmd(l.cmd, listOpts)
	AddListPlansCmd(l.cmd, listOpts)
	addListTeamMembersCmd(l.cmd, listOpts)
	AddListLandscapeLogsCmd(l.cmd, listOpts)
}
