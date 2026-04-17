// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/codesphere-cloud/cs-go/pkg/pipeline"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// PsCmd represents the ps command
type PsCmd struct {
	cmd  *cobra.Command
	Opts *GlobalOptions
}

type ServerStatus struct {
	Server         string `json:"server"`
	ReplicaCount   int    `json:"state"`
	ReplicaRunning int    `json:"replica"`
}

func (c *PsCmd) RunE(_ *cobra.Command, args []string) error {
	client, err := NewClient(*c.Opts)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	wsId, err := c.Opts.GetWorkspaceId()
	if err != nil {
		return fmt.Errorf("failed to get workspace ID: %w", err)
	}
	status, err := client.GetPipelineState(wsId, "run")
	if err != nil {
		return fmt.Errorf("failed to get pipeline state: %w", err)
	}

	serverStatus := map[string]*ServerStatus{}
	for _, replica := range status {
		stat, ok := serverStatus[replica.Server]
		if !ok {
			stat = &ServerStatus{
				Server:         replica.Server,
				ReplicaCount:   0,
				ReplicaRunning: 0,
			}
			serverStatus[replica.Server] = stat
		}
		if replica.State == "running" || replica.Server == pipeline.IdeServer {
			stat.ReplicaRunning++
		}
		stat.ReplicaCount++
		continue
	}

	t := io.GetTableWriter()
	t.AppendHeader(table.Row{"Server", "Replica (running/desired)"})
	for _, stat := range serverStatus {
		t.AppendRow(table.Row{stat.Server, fmt.Sprintf("%d/%d", stat.ReplicaRunning, stat.ReplicaCount)})
	}
	t.Render()

	return nil
}

func AddPsCmd(rootCmd *cobra.Command, opts *GlobalOptions) {
	ps := PsCmd{
		cmd: &cobra.Command{
			Use:   "ps",
			Short: "List services of a workspace",
			Long:  `Lists all services of a workspace with their current state.`,
		},
		Opts: opts,
	}
	rootCmd.AddCommand(ps.cmd)
	ps.cmd.RunE = ps.RunE
}
