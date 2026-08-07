// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"

	shared "github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/spf13/cobra"
)

type ListTeamsCmd struct {
	cmd  *cobra.Command
	opts *ListOptions
}

func addListTeamsCmd(p *cobra.Command, opts *ListOptions) {
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
	shared.AddCmd(p, l.cmd)
}

func (l *ListTeamsCmd) RunE(_ *cobra.Command, args []string) (err error) {
	client, err := l.opts.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	orgId, err := l.opts.GetOrgId()
	if err != nil {
		return fmt.Errorf("failed to get organization ID: %w", err)
	}
	teams, err := client.ListTeams(orgId)
	if err != nil {
		return fmt.Errorf("failed to list teams: %w", err)
	}
	switch l.opts.OutputFormat {
	case shared.OutputFormatJSON:
		return io.PrintJSON(teams)
	case shared.OutputFormatYAML:
		return io.PrintYAML(teams)
	}

	t := io.GetTableWriter()
	t.AppendHeader(table.Row{"P", "ID", "Name", "Role", "Default DC"})
	for _, team := range teams {
		first := ""
		if team.IsFirst != nil && *team.IsFirst {
			first = "*"
		}
		roleName := "N/A"
		if team.Role != nil {
			roleName = cs.GetRoleName(*team.Role)
		}
		t.AppendRow(table.Row{first, team.Id, team.Name, roleName, team.DefaultDataCenterId})
	}
	t.Render()

	return nil
}
