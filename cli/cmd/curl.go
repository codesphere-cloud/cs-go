// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	io_pkg "github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

// DefaultCommandExecutor uses os/exec to run commands
type DefaultCommandExecutor struct{}

func (e *DefaultCommandExecutor) Execute(ctx context.Context, name string, args []string, stdout, stderr io.Writer) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

type CurlCmd struct {
	cmd      *cobra.Command
	Opts     GlobalOptions
	Port     *int
	Timeout  *time.Duration
	Insecure bool
	Executor CommandExecutor // Injectable for testing
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
			Example: io_pkg.FormatExampleCommands("curl", []io_pkg.Example{
				{Cmd: "/ -w 1234", Desc: "GET request to workspace root"},
				{Cmd: "/api/health -w 1234 -p 3001", Desc: "GET request to port 3001"},
				{Cmd: "/api/data -w 1234 -- -XPOST -d '{\"key\":\"value\"}'", Desc: "POST request with data"},
				{Cmd: "/api/endpoint -w 1234 -- -v", Desc: "verbose output"},
				{Cmd: "/ -- -I", Desc: "HEAD request using workspace from env var"},
			}),
			Args: cobra.MinimumNArgs(1),
		},
		Opts:     opts,
		Executor: &DefaultCommandExecutor{},
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

	// Get the dev domain from the workspace
	if workspace.DevDomain == nil || *workspace.DevDomain == "" {
		return fmt.Errorf("workspace %d does not have a dev domain configured", wsId)
	}

	port := 3000
	if c.Port != nil {
		port = *c.Port
	}

	// Use the workspace's dev domain and replace the port if needed
	// DevDomain format is: {workspace_id}-{port}.{domain}
	devDomain := *workspace.DevDomain
	var url string
	if port != 3000 {
		// Replace the default port (3000) with the custom port in the dev domain
		url = fmt.Sprintf("https://%d-%d.%s%s", wsId, port, devDomain[strings.Index(devDomain, ".")+1:], path)
	} else {
		url = fmt.Sprintf("https://%s%s", devDomain, path)
	}

	fmt.Fprintf(os.Stderr, "Sending request to workspace %d (%s) at %s\n", wsId, workspace.Name, url)

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
	err = c.Executor.Execute(ctx, cmdArgs[0], cmdArgs[1:], os.Stdout, os.Stderr)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("timeout exceeded while requesting workspace %d", wsId)
		}
		return fmt.Errorf("curl command failed: %w", err)
	}

	return nil
}
