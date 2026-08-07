// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"

	"github.com/codesphere-cloud/cs-go/api"
	shared "github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

type ListOrgCmd struct {
	cmd           *cobra.Command
	Opts          *ListOptions
	ClientFactory func(shared.RootOptions) (Client, error)
}

func AddListOrgCmd(p *cobra.Command, opts *ListOptions,
) {
	l := ListOrgCmd{
		cmd: &cobra.Command{
			Use:     "organization",
			Aliases: []string{"organization", "org", "orgs"},
			Short:   "List organizations",
			Long:    `List organizations available in Codesphere`,
			Example: io.FormatExampleCommands("list org", []io.Example{
				{Desc: "List all organizations"},
			}),
		},
		Opts:          opts,
		ClientFactory: func(opts shared.RootOptions) (Client, error) { return opts.NewClient() },
	}
	l.cmd.RunE = l.RunE
	shared.AddCmd(p, l.cmd)
}

func (l *ListOrgCmd) RunE(_ *cobra.Command, args []string) (err error) {
	client, err := l.ClientFactory(l.Opts)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	_, err = l.ListOrganizations(client)
	if err != nil {
		return fmt.Errorf("failed to list organizations: %w", err)
	}

	return nil
}

func (l *ListOrgCmd) ListOrganizations(client Client) ([]api.Organization, error) {
	orgs, err := client.ListOrganizations()
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	switch l.Opts.OutputFormat {
	case shared.OutputFormatJSON:
		return orgs, io.PrintJSON(orgs)
	case shared.OutputFormatYAML:
		return orgs, io.PrintYAML(orgs)
	}

	t := io.GetTableWriter()
	t.AppendHeader(table.Row{"ID", "Name"})
	for _, org := range orgs {
		t.AppendRow(table.Row{org.Id, org.Name})
	}
	t.Render()

	return orgs, nil
}
