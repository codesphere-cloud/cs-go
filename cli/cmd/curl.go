// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type CurlCmd struct {
	cmd      *cobra.Command
	Opts     GlobalOptions
	Port     *int
	Timeout  *time.Duration
	Insecure bool
}

func (c *CurlCmd) RunE(_ *cobra.Command, args []string) error {
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

	if len(args) == 0 {
		return fmt.Errorf("path is required (e.g., / or /api/endpoint)")
	}

	path := args[0]
	curlArgs := args[1:]

	return c.CurlWorkspace(client, wsId, token, path, curlArgs)
}

func AddCurlCmd(rootCmd *cobra.Command, opts GlobalOptions) {
	curl := CurlCmd{
		cmd: &cobra.Command{
			Use:   "curl [path] [-- curl-args...]",
			Short: "Send authenticated HTTP requests to workspace dev domain",
			Long:  `Send authenticated HTTP requests to a workspace's development domain using curl-like syntax.`,
			Example: io.FormatExampleCommands("curl", []io.Example{
				{Cmd: "/ -w 1234", Desc: "GET request to workspace root"},
				{Cmd: "/api/health -w 1234 -p 3001", Desc: "GET request to port 3001"},
				{Cmd: "/api/data -w 1234 -- -XPOST -d '{\"key\":\"value\"}'", Desc: "POST request with data"},
				{Cmd: "/api/endpoint -w 1234 -- -v", Desc: "verbose output"},
				{Cmd: "/ -- -I", Desc: "HEAD request using workspace from env var"},
			}),
			Args: cobra.MinimumNArgs(1),
		},
		Opts: opts,
	}
	curl.Port = curl.cmd.Flags().IntP("port", "p", 3000, "Port to connect to")
	curl.Timeout = curl.cmd.Flags().DurationP("timeout", "", 30*time.Second, "Timeout for the request")
	curl.cmd.Flags().BoolVar(&curl.Insecure, "insecure", false, "skip TLS certificate verification (for testing only)")
	rootCmd.AddCommand(curl.cmd)
	curl.cmd.RunE = curl.RunE
}

func (c *CurlCmd) CurlWorkspace(client Client, wsId int, token string, path string, curlArgs []string) error {
	workspace, err := client.GetWorkspace(wsId)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Get team to obtain datacenter ID
	team, err := client.GetTeam(workspace.TeamId)
	if err != nil {
		return fmt.Errorf("failed to get team: %w", err)
	}

	port := 3000
	if c.Port != nil {
		port = *c.Port
	}

	// Construct URL using datacenter format: ${WORKSPACE_ID}-${PORT}.${DATACENTER_ID}.codesphere.com
	url := fmt.Sprintf("https://%d-%d.%d.codesphere.com%s", wsId, port, team.DefaultDataCenterId, path)

	log.Printf("Sending request to workspace %d (%s) at %s\n", wsId, workspace.Name, url)

	timeout := 30 * time.Second
	if c.Timeout != nil {
		timeout = *c.Timeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Build curl command
	cmdArgs := []string{"curl"}

	// Add authentication header
	cmdArgs = append(cmdArgs, "-H", fmt.Sprintf("x-forward-security: %s", token))

	// Add insecure flag if specified
	if c.Insecure {
		cmdArgs = append(cmdArgs, "-k")
	}

	// Add user's curl arguments
	cmdArgs = append(cmdArgs, curlArgs...)

	// Add URL as the last argument
	cmdArgs = append(cmdArgs, url)

	// Execute curl command
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("timeout exceeded while requesting workspace %d", wsId)
		}
		return fmt.Errorf("curl command failed: %w", err)
	}

	return nil
}
