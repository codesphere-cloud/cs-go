// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"time"

	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type WakeUpOptions struct {
	*GlobalOptions
	Timeout time.Duration
	Profile string
}

type WakeUpCmd struct {
	cmd  *cobra.Command
	Opts WakeUpOptions
}

func (c *WakeUpCmd) RunE(_ *cobra.Command, args []string) error {
	client, err := NewClient(*c.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	wsId, err := c.Opts.GetWorkspaceId()
	if err != nil {
		return fmt.Errorf("failed to get workspace ID: %w", err)
	}

	token, err := c.Opts.Env().GetApiToken()
	if err != nil {
		return fmt.Errorf("failed to get API token: %w", err)
	}

	return client.WakeUpWorkspace(wsId, token, c.Opts.Profile, c.Opts.Timeout)
}

func AddWakeUpCmd(rootCmd *cobra.Command, opts *GlobalOptions) {
	wakeup := WakeUpCmd{
		cmd: &cobra.Command{
			Use:   "wake-up",
			Short: "Wake up an on-demand workspace",
			Long:  `Wake up an on-demand workspace by scaling it to 1 replica via the API. Optionally syncs the landscape to start services.`,
			Example: io.FormatExampleCommands("wake-up", []io.Example{
				{Cmd: "-w 1234", Desc: "wake up workspace 1234"},
				{Cmd: "", Desc: "wake up workspace set by environment variable CS_WORKSPACE_ID"},
				{Cmd: "-w 1234 --timeout 60s", Desc: "wake up workspace with 60 second timeout"},
				{Cmd: "-w 1234", Desc: "wake up workspace and deploy landscape from CI profile"},
				{Cmd: "-w 1234 --profile prod", Desc: "wake up workspace and deploy landscape with prod profile"},
			}),
		},
		Opts: WakeUpOptions{
			GlobalOptions: opts,
		},
	}
	wakeup.cmd.Flags().DurationVar(&wakeup.Opts.Timeout, "timeout", 120*time.Second, "Timeout for waking up the workspace")
	wakeup.cmd.Flags().StringVarP(&wakeup.Opts.Profile, "profile", "p", "", "CI profile to use for landscape deploy (e.g. 'prod' for ci.prod.yml)")
	rootCmd.AddCommand(wakeup.cmd)
	wakeup.cmd.RunE = wakeup.RunE
}
