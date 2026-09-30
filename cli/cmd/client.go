// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

//go:generate go tool mockery

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd/shared"
)

// Client aliases the common contract so root and child commands use the same interface.
type Client = shared.Client

// CommandExecutor abstracts command execution for testing
type CommandExecutor interface {
	// Execute runs the command, streaming to stdout/stderr as usual, and additionally
	// returns the combined stdout+stderr output that was written.
	Execute(ctx context.Context, name string, args []string, stdout, stderr io.Writer) (string, error)
}

func NewClient(opts GlobalOptions) (Client, error) {
	return opts.NewClient()
}

func (o GlobalOptions) NewClient() (*api.Client, error) {
	token, err := o.Env.GetApiToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get API token: %w", err)
	}
	apiUrl, err := url.Parse(o.GetApiUrl())
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL '%s': %w", o.GetApiUrl(), err)
	}
	client := api.NewClient(context.Background(), api.Configuration{
		BaseUrl: apiUrl,
		Token:   token,
		Verbose: o.Verbose,
	})
	return client, nil
}
