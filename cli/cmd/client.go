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
)

type Client = api.Client

// CommandExecutor abstracts command execution for testing
type CommandExecutor interface {
	Execute(ctx context.Context, name string, args []string, stdout, stderr io.Writer) error
}

func NewClient(opts GlobalOptions) (Client, error) {
	token, err := opts.Env().GetApiToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get API token: %w", err)
	}
	apiUrl, err := url.Parse(opts.GetApiUrl())
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL '%s': %w", opts.GetApiUrl(), err)
	}
	client := api.NewClient(context.Background(), api.Configuration{
		BaseUrl: apiUrl,
		Token:   token,
		Verbose: opts.Verbose,
	})
	return client, nil
}
