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

type ListTeamMembersCmd struct {
	cmd           *cobra.Command
	Opts          ListTeamMembersOpts
	ClientFactory func(GlobalOptions) (Client, error)
}

type ListTeamMembersOpts struct {
	*GlobalOptions
	OutputFormat OutputFormat
}

func AddListTeamMembersCmd(team *cobra.Command, opts *GlobalOptions) {
	l := ListTeamMembersCmd{
		cmd: &cobra.Command{
			Use:   "list",
			Short: "List team members",
			Long:  `List all members of a team`,
			Example: io.FormatExampleCommands("team member list", []io.Example{
				{Cmd: "-t <teamId>", Desc: "List all members of a team"},
				{Cmd: "-t <teamId> -o json", Desc: "List all members of a team in JSON format"},
			}),
		},
		Opts: ListTeamMembersOpts{
			GlobalOptions: opts,
		},
		ClientFactory: NewClient,
	}
	l.cmd.RunE = l.RunE
	l.cmd.Flags().StringVarP((*string)(&l.Opts.OutputFormat), "output", "o", "table", "Output format (table, json, yaml)")
	AddCmd(team, l.cmd)
}

func (l *ListTeamMembersCmd) RunE(_ *cobra.Command, args []string) error {
	client, err := l.ClientFactory(*l.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	teamId, err := l.Opts.GetTeamId()
	if err != nil {
		return err
	}

	return l.ListTeamMembers(client, teamId)
}

func (l *ListTeamMembersCmd) ListTeamMembers(client Client, teamId int) error {
	members, err := client.ListTeamMembers(teamId)
	if err != nil {
		return fmt.Errorf("failed to list team members: %w", err)
	}

	switch l.Opts.OutputFormat {
	case OutputFormatJSON:
		return io.PrintJSON(members)
	case OutputFormatYAML:
		return io.PrintYAML(members)
	}

	t := io.GetTableWriter()
	t.AppendHeader(table.Row{"User ID", "Name", "Email", "Role", "Pending"})
	for _, m := range members {
		name := ""
		if m.Name != nil {
			name = *m.Name
		}
		email := ""
		if m.Email != nil {
			email = *m.Email
		}
		t.AppendRow(table.Row{m.UserId, name, email, cs.GetRoleName(m.Role), m.Pending})
	}
	t.Render()

	return nil
}
