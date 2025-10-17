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

type ListBaseimagesCmd struct {
	Opts GlobalOptions
	cmd  *cobra.Command
}

func AddListBaseimagesCmd(p *cobra.Command, opts GlobalOptions) {
	l := ListBaseimagesCmd{
		cmd: &cobra.Command{
			Use:   "baseimages",
			Short: "List baseimages",
			Long:  `List baseimages available in Codesphere for workspace creation`,
			Example: io.FormatExampleCommands("list baseimages", []io.Example{
				{Cmd: "", Desc: "List all baseimages"},
			}),
		},
		Opts: opts,
	}
	l.cmd.RunE = l.RunE
	p.AddCommand(l.cmd)
}

func (l *ListBaseimagesCmd) RunE(_ *cobra.Command, args []string) (err error) {
	client, err := NewClient(l.Opts)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	baseimages, err := l.ListBaseimages(client)
	if err != nil {
		return fmt.Errorf("failed to list baseimages: %w", err)
	}

	t := io.GetTableWriter()
	t.AppendHeader(table.Row{"ID", "Name", "SupportedUntil"})
	for _, b := range baseimages {
		t.AppendRow(table.Row{b.Id, b.Name, b.SupportedUntil.Format("2006-01-02")})
	}
	t.Render()

	return nil
}

func (l *ListBaseimagesCmd) ListBaseimages(client Client) ([]api.Baseimage, error) {
	baseimages, err := client.ListBaseimages()
	if err != nil {
		return nil, fmt.Errorf("failed to list baseimages: %w", err)
	}
	return baseimages, nil
}
