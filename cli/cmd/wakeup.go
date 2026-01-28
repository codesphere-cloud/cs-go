// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"time"

	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type WakeUpCmd struct {
	cmd     *cobra.Command
	Opts    GlobalOptions
	Timeout *time.Duration
}

func (c *WakeUpCmd) RunE(_ *cobra.Command, args []string) error {
	client, err := NewClient(c.Opts)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	wsId, err := c.Opts.GetWorkspaceId()
	if err != nil {
		return fmt.Errorf("failed to get workspace ID: %w", err)
	}

	return c.WakeUpWorkspace(client, wsId)
}

func AddWakeUpCmd(rootCmd *cobra.Command, opts GlobalOptions) {
	wakeup := WakeUpCmd{
		cmd: &cobra.Command{
			Use:   "wake-up",
			Short: "Wake up an on-demand workspace",
			Long:  `Wake up an on-demand workspace by scaling it to 1 replica via the API.`,
			Example: io.FormatExampleCommands("wake-up", []io.Example{
				{Cmd: "-w 1234", Desc: "wake up workspace 1234"},
				{Cmd: "", Desc: "wake up workspace set by environment variable CS_WORKSPACE_ID"},
				{Cmd: "-w 1234 --timeout 60s", Desc: "wake up workspace with 60 second timeout"},
			}),
		},
		Opts: opts,
	}
	wakeup.Timeout = wakeup.cmd.Flags().DurationP("timeout", "", 120*time.Second, "Timeout for waking up the workspace")
	rootCmd.AddCommand(wakeup.cmd)
	wakeup.cmd.RunE = wakeup.RunE
}

func (c *WakeUpCmd) WakeUpWorkspace(client Client, wsId int) error {
	workspace, err := client.GetWorkspace(wsId)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Check if workspace is already running
	status, err := client.WorkspaceStatus(wsId)
	if err != nil {
		return fmt.Errorf("failed to get workspace status: %w", err)
	}

	if status.IsRunning {
		fmt.Printf("Workspace %d (%s) is already running\n", wsId, workspace.Name)
		return nil
	}

	fmt.Printf("Waking up workspace %d (%s)...\n", wsId, workspace.Name)

	// Scale workspace to at least 1 replica to wake it up
	// If workspace already has replicas configured (but not running), preserve that count
	targetReplicas := 1
	if workspace.Replicas > 1 {
		targetReplicas = workspace.Replicas
	}

	err = client.ScaleWorkspace(wsId, targetReplicas)
	if err != nil {
		return fmt.Errorf("failed to scale workspace: %w", err)
	}

	timeout := 120 * time.Second
	if c.Timeout != nil {
		timeout = *c.Timeout
	}

	fmt.Printf("Waiting for workspace %d to be running...\n", wsId)
	err = client.WaitForWorkspaceRunning(&workspace, timeout)
	if err != nil {
		return fmt.Errorf("workspace did not become running: %w", err)
	}

	fmt.Printf("Workspace %d is now running\n", wsId)
	return nil
}
