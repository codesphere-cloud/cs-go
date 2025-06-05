// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/codesphere-cloud/cs-go/pkg/out"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

// ListPlansCmd represents the plans command
type ListPlansCmd struct {
	cmd  *cobra.Command
	Opts GlobalOptions
}

func (c *ListPlansCmd) RunE(_ *cobra.Command, args []string) error {

	client, err := NewClient(c.Opts)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	plans, err := client.ListWorkspacePlans()
	if err != nil {
		return fmt.Errorf("failed to list plans: %s", err)
	}

	t := out.GetTableWriter()
	t.AppendHeader(table.Row{"ID", "Name", "On Demand", "CPU", "RAM(GiB)", "SSD(GiB)", "Price(USD)", "Max Replicas"})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Name: "Price(USD)", Align: text.AlignRight},
		{Name: "RAM(GiB)", Align: text.AlignRight},
		{Name: "SSD(GiB)", Align: text.AlignRight},
	})
	for _, plan := range plans {
		if plan.Deprecated {
			continue
		}
		onDemand := ""
		if plan.Characteristics.OnDemand {
			onDemand = "*"
		}
		t.AppendRow(table.Row{
			plan.Id,
			plan.Title,
			onDemand,
			plan.Characteristics.CPU,
			formatBytesAsGib(plan.Characteristics.RAM),
			formatBytesAsGib(plan.Characteristics.SSD),
			fmt.Sprintf("%.2f", plan.PriceUsd),
			plan.MaxReplicas,
		})
	}
	t.Render()

	return nil
}

func formatBytesAsGib(in int) string {
	return fmt.Sprintf("%.2f", float32(in)/1024/1024/1024)
}

func AddListPlansCmd(list *cobra.Command, opts GlobalOptions) {
	plans := ListPlansCmd{
		cmd: &cobra.Command{
			Use:   "plans",
			Short: "List available plans",
			Long: out.Long(`List available workpace plans.

				When creating new workspaces you need to select a specific plan.`),
		},
		Opts: opts,
	}
	list.AddCommand(plans.cmd)
	plans.cmd.RunE = plans.RunE
}
