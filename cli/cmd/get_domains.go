// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/spf13/cobra"
)

type GetDomainsCmd struct {
	Opts GlobalOptions
	cmd  *cobra.Command
}

func addGetDomainsCmd(p *cobra.Command, opts GlobalOptions) {
	g := GetDomainsCmd{
		cmd: &cobra.Command{
			Use:   "domains",
			Short: "Get domains for a workspace",
			Long:  `Get both the devDomain and any custom domains for a workspace`,
			Example: io.FormatExampleCommands("list domains", []io.Example{
				{Cmd: "--workspace-id <workspace-id>", Desc: "Get domains for a specific workspace"},
			}),
		},
		Opts: opts,
	}
	g.cmd.RunE = g.RunE
	p.AddCommand(g.cmd)
}

func (g *GetDomainsCmd) RunE(_ *cobra.Command, args []string) (err error) {
	client, err := NewClient(g.Opts)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	workspaceId, err := g.Opts.GetWorkspaceId()
	if err != nil {
		return fmt.Errorf("failed to get workspace ID: %w", err)
	}

	domains, err := client.GetWorkspaceDomains(workspaceId)
	if err != nil {
		return fmt.Errorf("failed to get workspace domains: %w", err)
	}

	t := io.GetTableWriter()
	t.AppendHeader(table.Row{"Type", "Domain"})

	t.AppendRow(table.Row{"Dev Domain", domains.DevDomain})

	for _, customDomain := range domains.CustomDomains {
		t.AppendRow(table.Row{"Custom Domain", customDomain})
	}

	t.Render()

	return nil
}
