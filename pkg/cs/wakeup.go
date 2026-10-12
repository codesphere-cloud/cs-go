// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cs

import (
	"fmt"
	"log"
	"time"

	"github.com/codesphere-cloud/cs-go/api"
)

// Client is the API contract needed to wake up a workspace.
type Client interface {
	GetWorkspace(workspaceId int) (api.Workspace, error)
	WorkspaceStatus(workspaceId int) (*api.WorkspaceStatus, error)
	ScaleWorkspace(wsId int, replicas int) error
	WaitForWorkspaceRunning(workspace *api.Workspace, timeout time.Duration) error
}

// WakeUpWorkspace returns the workspace after ensuring it is running. If it is
// stopped, it scales the workspace to at least one replica and waits up to
// timeout for it to start.
func WakeUpWorkspace(client Client, wsId int, timeout time.Duration) (*api.Workspace, error) {
	workspace, err := client.GetWorkspace(wsId)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	// Check if workspace is already running
	status, err := client.WorkspaceStatus(wsId)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace status: %w", err)
	}
	if status.IsRunning {
		log.Printf("Workspace %d (%s) is already running\n", wsId, workspace.Name)
		return &workspace, nil
	}

	log.Printf("Waking up workspace %d (%s)...\n", wsId, workspace.Name)

	// Scale workspace to at least 1 replica to wake it up
	// If workspace already has replicas configured (but not running), preserve that count
	targetReplicas := 1
	if workspace.Replicas > 1 {
		targetReplicas = workspace.Replicas
	}

	err = client.ScaleWorkspace(wsId, targetReplicas)
	if err != nil {
		return nil, fmt.Errorf("failed to scale workspace: %w", err)
	}

	log.Printf("Waiting for workspace %d to be running...\n", wsId)
	err = client.WaitForWorkspaceRunning(&workspace, timeout)
	if err != nil {
		return nil, fmt.Errorf("workspace did not become running: %w", err)
	}
	return &workspace, nil
}
