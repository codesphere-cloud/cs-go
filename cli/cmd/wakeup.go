// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type WakeUpCmd struct {
	cmd      *cobra.Command
	Opts     GlobalOptions
	Timeout  *time.Duration
	Insecure bool
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

	token, err := c.Opts.Env.GetApiToken()
	if err != nil {
		return fmt.Errorf("failed to get API token: %w", err)
	}

	return c.WakeUpWorkspace(client, wsId, token)
}

func AddWakeUpCmd(rootCmd *cobra.Command, opts GlobalOptions) {
	wakeup := WakeUpCmd{
		cmd: &cobra.Command{
			Use:   "wake-up",
			Short: "Wake up an on-demand workspace",
			Long:  `Wake up an on-demand workspace by making an authenticated request to its services domain.`,
			Example: io.FormatExampleCommands("wake-up", []io.Example{
				{Cmd: "-w 1234", Desc: "wake up workspace 1234"},
				{Cmd: "", Desc: "wake up workspace set by environment variable CS_WORKSPACE_ID"},
				{Cmd: "-w 1234 --timeout 60s", Desc: "wake up workspace with 60 second timeout"},
			}),
		},
		Opts: opts,
	}
	wakeup.Timeout = wakeup.cmd.Flags().DurationP("timeout", "", 120*time.Second, "Timeout for waking up the workspace")
	wakeup.cmd.Flags().BoolVar(&wakeup.Insecure, "insecure", false, "skip TLS certificate verification (for testing only)")
	rootCmd.AddCommand(wakeup.cmd)
	wakeup.cmd.RunE = wakeup.RunE
}

func (c *WakeUpCmd) WakeUpWorkspace(client Client, wsId int, token string) error {
	workspace, err := client.GetWorkspace(wsId)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Get team to obtain datacenter ID
	team, err := client.GetTeam(workspace.TeamId)
	if err != nil {
		return fmt.Errorf("failed to get team: %w", err)
	}

	// Construct the services domain using datacenter format: ${WORKSPACE_ID}-3000.${DATACENTER_ID}.codesphere.com
	servicesDomain := fmt.Sprintf("https://%d-3000.%d.codesphere.com", wsId, team.DefaultDataCenterId)

	log.Printf("Waking up workspace %d (%s) at URL: %s\n", wsId, workspace.Name, servicesDomain)
	timeout := 120 * time.Second
	if c.Timeout != nil {
		timeout = *c.Timeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err = makeWakeUpRequest(ctx, servicesDomain, token, c.Insecure)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("timeout exceeded while waking up workspace %d", wsId)
		}
		return fmt.Errorf("failed to wake up workspace: %w", err)
	}

	log.Printf("Waiting for workspace %d to be running...\n", wsId)
	err = client.WaitForWorkspaceRunning(&workspace, timeout)
	if err != nil {
		return fmt.Errorf("workspace did not become running: %w", err)
	}

	log.Printf("Successfully woke up workspace %d\n", wsId)
	return nil
}

func makeWakeUpRequest(ctx context.Context, servicesDomain string, token string, insecure bool) error {
	req, err := http.NewRequestWithContext(ctx, "GET", servicesDomain, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-forward-security", token)

	transport := &http.Transport{}
	if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	log.Printf("Wake-up request received status: %d %s\n", resp.StatusCode, resp.Status)

	// Accept 2xx, 3xx, and 5xx responses (5xx is expected when workspace is starting)
	// 4xx errors indicate authentication issues
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return fmt.Errorf("authentication failed: %s", resp.Status)
	}

	return nil
}
