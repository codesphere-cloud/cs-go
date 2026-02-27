// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type WakeUpOptions struct {
	*GlobalOptions
	Timeout       time.Duration
	SyncLandscape bool
	Profile       string
	ScaleServices string // format: "service=replicas,service2=replicas"
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

	return c.WakeUpWorkspace(client, wsId)
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
				{Cmd: "-w 1234 --sync-landscape", Desc: "wake up workspace and deploy landscape from CI profile"},
				{Cmd: "-w 1234 --sync-landscape --profile prod", Desc: "wake up workspace and deploy landscape with prod profile"},
				{Cmd: "-w 1234 --scale-services web=1,api=2", Desc: "wake up workspace and scale specific services"},
			}),
		},
		Opts: WakeUpOptions{
			GlobalOptions: opts,
		},
	}
	wakeup.cmd.Flags().DurationVar(&wakeup.Opts.Timeout, "timeout", 120*time.Second, "Timeout for waking up the workspace")
	wakeup.cmd.Flags().BoolVar(&wakeup.Opts.SyncLandscape, "sync-landscape", false, "Deploy landscape from CI profile after waking up")
	wakeup.cmd.Flags().StringVarP(&wakeup.Opts.Profile, "profile", "p", "", "CI profile to use for landscape deploy (e.g. 'prod' for ci.prod.yml)")
	wakeup.cmd.Flags().StringVar(&wakeup.Opts.ScaleServices, "scale-services", "", "Scale specific landscape services (format: 'service=replicas,service2=replicas')")
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

	if !status.IsRunning {
		log.Printf("Waking up workspace %d (%s)...\n", wsId, workspace.Name)

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

		log.Printf("Waiting for workspace %d to be running...\n", wsId)
		err = client.WaitForWorkspaceRunning(&workspace, c.Opts.Timeout)
		if err != nil {
			return fmt.Errorf("workspace did not become running: %w", err)
		}
	}

	if c.Opts.SyncLandscape {
		log.Printf("Deploying landscape for workspace %d...\n", wsId)
		err = client.DeployLandscape(wsId, c.Opts.Profile)
		if err != nil {
			return fmt.Errorf("failed to deploy landscape: %w", err)
		}
		log.Printf("Landscape deployment initiated for workspace %d\n", wsId)
	}

	if c.Opts.ScaleServices != "" {
		services, err := parseScaleServices(c.Opts.ScaleServices)
		if err != nil {
			return fmt.Errorf("failed to parse scale-services: %w", err)
		}
		log.Printf("Scaling landscape services for workspace %d: %v\n", wsId, services)
		err = client.ScaleLandscapeServices(wsId, services)
		if err != nil {
			return fmt.Errorf("failed to scale landscape services: %w", err)
		}
		log.Printf("Landscape services scaled for workspace %d\n", wsId)
	}

	if workspace.DevDomain == nil || *workspace.DevDomain == "" {
		log.Printf("Workspace %d does not have a dev domain, skipping health check\n", wsId)
		return nil
	}

	log.Printf("Checking health of workspace %d (%s)...\n", wsId, workspace.Name)

	token, err := c.Opts.Env.GetApiToken()
	if err != nil {
		return fmt.Errorf("failed to get API token: %w", err)
	}

	err = c.waitForWorkspaceHealthy(*workspace.DevDomain, token, c.Opts.Timeout)
	if err != nil {
		return fmt.Errorf("workspace did not become healthy: %w", err)
	}

	log.Printf("Workspace %d is healthy and ready\n", wsId)

	return nil
}

func (c *WakeUpCmd) waitForWorkspaceHealthy(devDomain string, token string, timeout time.Duration) error {
	url := fmt.Sprintf("https://%s", devDomain)
	delay := 5 * time.Second
	maxWaitTime := time.Now().Add(timeout)

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("X-CS-Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := httpClient.Do(req)
		if err == nil {
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
			log.Printf("Workspace %s responded with status code %d, retrying...\n", devDomain, resp.StatusCode)
		}

		if time.Now().After(maxWaitTime) {
			return fmt.Errorf("timeout waiting for workspace to be healthy at %s", url)
		}

		time.Sleep(delay)
	}
}

// parseScaleServices parses a string like "web=1,api=2" into a map[string]int
func parseScaleServices(s string) (map[string]int, error) {
	result := make(map[string]int)
	if s == "" {
		return result, nil
	}

	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid format '%s', expected 'service=replicas'", pair)
		}
		service := strings.TrimSpace(parts[0])
		if service == "" {
			return nil, fmt.Errorf("empty service name in '%s'", pair)
		}
		replicas, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid replica count '%s' for service '%s': %w", parts[1], service, err)
		}
		if replicas < 1 {
			return nil, fmt.Errorf("replica count must be at least 1 for service '%s'", service)
		}
		result[service] = replicas
	}
	return result, nil
}
