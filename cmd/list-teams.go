/*
Copyright © 2025 Codesphere Inc. <support@codesphere.com>
*/
package cmd

import (
	"fmt"

	"github.com/codesphere-cloud/cs-go/pkg/out"
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
			Short: "list teams",
			Long:  `list teams available in Codesphere`,
			Example: `
List all teams:

$ cs list teams
			`,
		},
		opts: opts,
	}
	l.cmd.RunE = l.RunE
	l.parseLogCmdFlags()
	p.AddCommand(l.cmd)
}

func (l *ListTeamsCmd) parseLogCmdFlags() {

}

func (l *ListTeamsCmd) RunE(_ *cobra.Command, args []string) (err error) {
	client, err := NewClient(l.opts)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %e", err)
	}

	teams, err := client.ListTeams()
	if err != nil {
		return fmt.Errorf("failed to list teams: %e", err)
	}

	t := out.GetTableWriter()
	t.AppendHeader(table.Row{"P", "ID", "Name", "Role", "Default DC"})
	for _, team := range teams {
		first := ""
		if team.IsFirst != nil && *team.IsFirst {
			first = "*"
		}
		role := "Admin"
		if team.Role == 1 {
			role = "Member"
		}
		t.AppendRow(table.Row{first, team.Id, team.Name, role, team.DefaultDataCenterId})
	}
	t.Render()

	return nil
}
