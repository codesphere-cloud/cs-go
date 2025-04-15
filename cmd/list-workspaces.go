/*
Copyright © 2025 Codesphere Inc. <support@codesphere.com>
*/
package cmd

import (
	"fmt"

	"github.com/codesphere-cloud/cs-go/pkg/api"
	"github.com/codesphere-cloud/cs-go/pkg/out"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/spf13/cobra"
)

type ListWorkspacesCmd struct {
	cmd  *cobra.Command
	opts ListWorkspacesOptions
}

type ListWorkspacesOptions struct {
	GlobalOptions
	TeamId *int
}

func addListWorkspacesCmd(p *cobra.Command, opts GlobalOptions) {
	l := ListWorkspacesCmd{
		cmd: &cobra.Command{
			Use:   "workspaces",
			Short: "list workspaces",
			Long:  `list workspaces available in Codesphere`,
			Example: `
List all workspaces:

$ cs list workspaces --team-id <team-id>
			`,
		},
		opts: ListWorkspacesOptions{GlobalOptions: opts},
	}
	l.cmd.RunE = l.RunE
	l.parseLogCmdFlags()
	p.AddCommand(l.cmd)
}

func (l *ListWorkspacesCmd) parseLogCmdFlags() {
	l.opts.TeamId = l.cmd.Flags().IntP("team-id", "t", -1, "ID of team to query")
}

func (l *ListWorkspacesCmd) RunE(_ *cobra.Command, args []string) (err error) {
	client, err := NewClient(l.opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %e", err)
	}

	teams, err := l.getTeamIds(client)
	if err != nil {
		return fmt.Errorf("failed to get teams: %e", err)
	}
	workspaces := []api.Workspace{}
	for _, team := range teams {
		teamWorkspaces, err := client.ListWorkspaces(team)
		if err != nil {
			return fmt.Errorf("failed to list workspaces: %e", err)
		}
		workspaces = append(workspaces, teamWorkspaces...)
	}

	t := out.GetTableWriter()
	t.AppendHeader(table.Row{"Team ID", "ID", "Name", "Repository"})
	for _, w := range workspaces {
		gitUrl := ""
		if w.GitUrl.Get() != nil {
			gitUrl = *w.GitUrl.Get()
		}
		t.AppendRow(table.Row{w.TeamId, w.Id, w.Name, gitUrl})
	}
	t.Render()

	return nil
}

func (l *ListWorkspacesCmd) getTeamIds(client *api.Client) (teams []int, err error) {
	if l.opts.TeamId != nil && *l.opts.TeamId >= 0 {
		teams = append(teams, *l.opts.TeamId)
		return
	}
	var allTeams []api.Team
	allTeams, err = client.ListTeams()
	if err != nil {
		return
	}
	for _, t := range allTeams {
		teams = append(teams, int(t.Id))
	}
	return
}
