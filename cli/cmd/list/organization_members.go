// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"

	shared "github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

type ListOrganizationMembersCmd struct {
	cmd           *cobra.Command
	Opts          *ListOptions
	ClientFactory func(shared.RootOptions) (shared.Client, error)
}

func AddListOrganizationMembersCmd(p *cobra.Command, opts *ListOptions) {
	l := ListOrganizationMembersCmd{
		cmd: &cobra.Command{
			Use:     "organization-members",
			Aliases: []string{"org-members"},
			Short:   "List organization members",
			Long:    `List all members of an organization`,
			Example: io.FormatExampleCommands("list organization-members", []io.Example{
				{Cmd: "--org <orgId>", Desc: "List all members of an organization"},
				{Cmd: "--org <orgId> --output json", Desc: "List all members of an organization in JSON format"},
			}),
		},
		Opts:          opts,
		ClientFactory: func(opts shared.RootOptions) (shared.Client, error) { return opts.NewClient() },
	}
	l.cmd.RunE = l.RunE
	shared.AddCmd(p, l.cmd)
}

func (l *ListOrganizationMembersCmd) RunE(_ *cobra.Command, args []string) error {
	client, err := l.ClientFactory(l.Opts)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	orgId, err := shared.RequireOrganizationID(l.Opts.RootOptions)
	if err != nil {
		return err
	}

	return l.ListOrganizationMembers(client, orgId)
}

func (l *ListOrganizationMembersCmd) ListOrganizationMembers(client shared.Client, orgId string) error {
	members, err := client.ListOrganizationMembers(orgId)
	if err != nil {
		return fmt.Errorf("failed to list organization members: %w", err)
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
		t.AppendRow(table.Row{m.UserId, name, email, m.Role, m.Pending})
	}
	t.Render()

	return nil
}
