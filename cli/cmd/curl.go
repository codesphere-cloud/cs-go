// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
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

type CurlOptions struct {
	Timeout  time.Duration
	Insecure bool
}

type CurlCmd struct {
	cmd      *cobra.Command
	Opts     GlobalOptions
	CurlOpts CurlOptions
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
				{Cmd: "/api/health -w 1234", Desc: "GET request to health endpoint"},
				{Cmd: "/api/data -w 1234 -- -XPOST -d '{\"key\":\"value\"}'", Desc: "POST request with data"},
				{Cmd: "/api/endpoint -w 1234 -- -v", Desc: "verbose output"},
				{Cmd: "/ -- -I", Desc: "HEAD request using workspace from env var"},
			}),
			Args: cobra.MinimumNArgs(1),
		},
		Opts:     opts,
		Executor: &DefaultCommandExecutor{},
	}
	curl.cmd.Flags().DurationVar(&curl.CurlOpts.Timeout, "timeout", 30*time.Second, "Timeout for the request")
	curl.cmd.Flags().BoolVar(&curl.CurlOpts.Insecure, "insecure", false, "skip TLS certificate verification (for testing only)")
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

	// Use the workspace's dev domain
	devDomain := *workspace.DevDomain
	url := fmt.Sprintf("https://%s%s", devDomain, path)

	log.Printf("Sending request to workspace %d (%s) at %s\n", wsId, workspace.Name, url)

	ctx, cancel := context.WithTimeout(context.Background(), c.CurlOpts.Timeout)
	defer cancel()

	// Build curl command with authentication header
	cmdArgs := []string{"curl", "-H", fmt.Sprintf("x-forward-security: %s", token)}

	// Add insecure flag if specified
	if c.CurlOpts.Insecure {
		cmdArgs = append(cmdArgs, "-k")
	}

	cmdArgs = append(cmdArgs, curlArgs...)
	cmdArgs = append(cmdArgs, url)

	err = c.Executor.Execute(ctx, cmdArgs[0], cmdArgs[1:], os.Stdout, os.Stderr)
	if err != nil && err == context.DeadlineExceeded {
		return fmt.Errorf("timeout exceeded while requesting workspace %d", wsId)
	}

	return err
}
