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

	if workspace.DevDomain == nil {
		return fmt.Errorf("workspace %d does not have a development domain configured", wsId)
	}

	// Construct the services domain: ${WORKSPACE_ID}-3000.${DEV_DOMAIN}
	servicesDomain := fmt.Sprintf("https://%d-3000.%s", wsId, *workspace.DevDomain)

	log.Printf("Waking up workspace %d (%s)...\n", wsId, workspace.Name)
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
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			req.Header.Set("x-forward-security", token)
			return nil
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// 4xx errors are considered failures
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return fmt.Errorf("received error response: %s", resp.Status)
	}

	return nil
}
