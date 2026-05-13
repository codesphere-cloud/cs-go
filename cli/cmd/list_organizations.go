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

type ListOrgCmd struct {
	cmd           *cobra.Command
	Opts          *ListOptions
	ClientFactory func(GlobalOptions) (Client, error)
}

func AddListOrgCmd(p *cobra.Command, opts *ListOptions,
) {
	l := ListOrgCmd{
		cmd: &cobra.Command{
			Use:   "org",
			Short: "List organizations",
			Long:  `List organizations available in Codesphere`,
			Example: io.FormatExampleCommands("list org", []io.Example{
				{Desc: "List all organizations"},
			}),
		},
		Opts:          opts,
		ClientFactory: NewClient, // Default to the real client in production
	}
	l.cmd.RunE = l.RunE
	AddCmd(p, l.cmd)
}

func (l *ListOrgCmd) RunE(_ *cobra.Command, args []string) (err error) {
	client, err := l.ClientFactory(*l.Opts.GlobalOptions) // Use the injected factory
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
	case OutputFormatJSON:
		return orgs, io.PrintJSON(orgs)
	case OutputFormatYAML:
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
