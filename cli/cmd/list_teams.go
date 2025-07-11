// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/spf13/cobra"
)

type ListTeamsCmd struct {
	cmd  *cobra.Command
	opts GlobalOptions
}

func addListTeamsCmd(p *cobra.Command, opts GlobalOptions) {
	l := ListTeamsCmd{
		cmd: &cobra.Command{
			Use:   "teams",
			Short: "List teams",
			Long:  `List teams available in Codesphere`,
			Example: io.FormatExampleCommands("list teams", []io.Example{
				{Desc: "List all teams"},
			}),
		},
		opts: opts,
	}
	l.cmd.RunE = l.RunE
	p.AddCommand(l.cmd)
}

func (l *ListTeamsCmd) RunE(_ *cobra.Command, args []string) (err error) {
	client, err := NewClient(l.opts)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	teams, err := client.ListTeams()
	if err != nil {
		return fmt.Errorf("failed to list teams: %w", err)
	}

	t := io.GetTableWriter()
	t.AppendHeader(table.Row{"P", "ID", "Name", "Role", "Default DC"})
	for _, team := range teams {
		first := ""
		if team.IsFirst != nil && *team.IsFirst {
			first = "*"
		}
		t.AppendRow(table.Row{first, team.Id, team.Name, cs.GetRoleName(int(team.Role)), team.DefaultDataCenterId})
	}
	t.Render()

	return nil
}
