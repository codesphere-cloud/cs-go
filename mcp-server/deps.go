// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"errors"
	"fmt"

	"github.com/codesphere-cloud/cs-go/cli/cmd/shared"
)

// Deps are the dependencies of the agent-facing tools.
type Deps struct {
	Client shared.Client
	// DefaultWorkspaceId returns the workspace of the current session (CS_WORKSPACE_ID).
	// It returns a negative ID when no default is configured.
	DefaultWorkspaceId func() (int, error)
}

// workspaceId returns the given ID, or the session's workspace when id is nil.
func (d Deps) workspaceId(id *int) (int, error) {
	if id != nil {
		if *id < 0 {
			return -1, errors.New("workspaceId must be non-negative")
		}
		return *id, nil
	}
	if d.DefaultWorkspaceId == nil {
		return -1, errors.New("workspaceId is required: no default workspace is configured")
	}
	wsId, err := d.DefaultWorkspaceId()
	if err != nil {
		return -1, fmt.Errorf("failed to read the default workspace from CS_WORKSPACE_ID: %w", err)
	}
	if wsId < 0 {
		return -1, errors.New("workspaceId is required: CS_WORKSPACE_ID is not set")
	}
	return wsId, nil
}
