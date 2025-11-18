// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/spf13/cobra"
)

type ListWorkspacesCmd struct {
	Opts GlobalOptions
	cmd  *cobra.Command
}

func addListWorkspacesCmd(p *cobra.Command, opts GlobalOptions) {
	l := ListWorkspacesCmd{
		cmd: &cobra.Command{
			Use:   "workspaces",
			Short: "List workspaces",
			Long:  `List workspaces available in Codesphere`,
			Example: io.FormatExampleCommands("list workspaces", []io.Example{
				{Cmd: "--team-id <team-id>", Desc: "List all workspaces"},
			}),
		},
		Opts: opts,
	}
	l.cmd.RunE = l.RunE
	p.AddCommand(l.cmd)
}

func (l *ListWorkspacesCmd) RunE(_ *cobra.Command, args []string) (err error) {
	client, err := NewClient(l.Opts)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	workspaces, err := l.ListWorkspaces(client)
	if err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	t := io.GetTableWriter()
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

func (l *ListWorkspacesCmd) ListWorkspaces(client Client) ([]api.Workspace, error) {
	teams, err := l.getTeamIds(client)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams: %w", err)
	}
	workspaces := []api.Workspace{}
	for _, team := range teams {
		teamWorkspaces, err := client.ListWorkspaces(team)
		if err != nil {
			return nil, fmt.Errorf("failed to list workspaces: %w", err)
		}
		workspaces = append(workspaces, teamWorkspaces...)
	}
	return workspaces, nil
}

func (l *ListWorkspacesCmd) getTeamIds(client Client) (teams []int, err error) {
	if l.Opts.TeamId != nil && *l.Opts.TeamId >= 0 {
		teams = append(teams, *l.Opts.TeamId)
		return
	}
	teamIdEnv, err := l.Opts.Env.GetTeamId()
	if err != nil {
		err = fmt.Errorf("failed to get team ID from env: %w", err)
		return
	}
	if teamIdEnv >= 0 {
		teams = append(teams, teamIdEnv)
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
