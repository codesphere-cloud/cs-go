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

type ListTeamMembersCmd struct {
	cmd           *cobra.Command
	Opts          *ListOptions
	ClientFactory func(shared.RootOptions) (shared.Client, error)
}

func AddListTeamMembersCmd(p *cobra.Command, opts *ListOptions) {
	l := ListTeamMembersCmd{
		cmd: &cobra.Command{
			Use:     "resource-group-members",
			Aliases: []string{"team-members"},
			Short:   "List resource group members",
			Long:    `List all members of a resource group`,
			Example: io.FormatExampleCommands("list resource-group-members", []io.Example{
				{Cmd: "-t <resourceGroupId>", Desc: "List all members of a resource group"},
				{Cmd: "-t <resourceGroupId> -o json", Desc: "List all members of a resource group in JSON format"},
			}),
		},
		Opts:          opts,
		ClientFactory: func(opts shared.RootOptions) (shared.Client, error) { return opts.NewClient() },
	}
	l.cmd.RunE = l.RunE
	shared.AddCmd(p, l.cmd)
}

func (l *ListTeamMembersCmd) RunE(_ *cobra.Command, args []string) error {
	client, err := l.ClientFactory(l.Opts)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	teamId, err := l.Opts.GetTeamId()
	if err != nil {
		return err
	}

	return l.ListTeamMembers(client, teamId)
}

func (l *ListTeamMembersCmd) ListTeamMembers(client shared.Client, teamId int) error {
	members, err := client.ListTeamMembers(teamId)
	if err != nil {
		return fmt.Errorf("failed to list team members: %w", err)
	}

	switch l.Opts.OutputFormat {
	case shared.OutputFormatJSON:
		return io.PrintJSON(members)
	case shared.OutputFormatYAML:
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
