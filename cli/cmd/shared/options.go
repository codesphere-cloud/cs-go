// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package shared

import (
	"github.com/codesphere-cloud/cs-go/api"
	"github.com/spf13/cobra"
)

type OutputFormat string

const (
	OutputFormatTable OutputFormat = "table"
	OutputFormatJSON  OutputFormat = "json"
	OutputFormatYAML  OutputFormat = "yaml"
)

type RootOptions interface {
	GetTeamId() (int, error)
	GetOrgId() (string, error)
	GetWorkspaceId() (int, error)
	GetApiUrl() string
	GetApiToken() (string, error)
	GetVerbose() bool
	NewClient() (*api.Client, error)
}

// AddCmd adds a command, inheriting the parent's Args validator if not explicitly set.
// Individual commands that need different argument rules can override this by setting their own Args validator.
func AddCmd(parent *cobra.Command, cmd *cobra.Command) {
	if cmd.Args == nil {
		cmd.Args = parent.Args
	}
	parent.AddCommand(cmd)
}
