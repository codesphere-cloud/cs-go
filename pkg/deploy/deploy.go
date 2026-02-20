// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package deploy

import (
	"fmt"
	"strings"
	"time"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/pkg/pipeline"
)

// Client defines the API operations needed for preview deployments.
// This is a subset of the full Codesphere API client.
// Pipeline operations (StartPipelineStage, GetPipelineState, DeployLandscape)
// are handled via the pipeline.Client interface.
type Client interface {
	pipeline.Client
	ListWorkspaces(teamId int) ([]api.Workspace, error)
	DeployWorkspace(args api.DeployWorkspaceArgs) (*api.Workspace, error)
	DeleteWorkspace(wsId int) error
	WaitForWorkspaceRunning(workspace *api.Workspace, timeout time.Duration) error
	SetEnvVarOnWorkspace(workspaceId int, vars map[string]string) error
	GitPull(wsId int, remote string, branch string) error
}

// Config holds all parameters needed for a preview deployment.
// This is provider-agnostic — no references to GitHub, GitLab, etc.
type Config struct {
	TeamId    int
	PlanId    int
	Name      string
	EnvVars   map[string]string
	VpnConfig string
	Branch    string
	Stages    []string
	RepoUrl   string
	Timeout   time.Duration
	Profile   string
}

// Result holds the output of a successful deployment.
type Result struct {
	WorkspaceId  int
	WorkspaceURL string
}

// Deployer orchestrates preview environment lifecycle operations.
type Deployer struct {
	Client Client
}

// NewDeployer creates a new preview deployer with the given API client.
func NewDeployer(client Client) *Deployer {
	return &Deployer{Client: client}
}

// FindWorkspace looks for an existing workspace by name within a team.
// Returns nil if no workspace with the given name is found.
func (d *Deployer) FindWorkspace(teamId int, name string) (*api.Workspace, error) {
	fmt.Printf("🔍 Looking for workspace '%s'...\n", name)

	workspaces, err := d.Client.ListWorkspaces(teamId)
	if err != nil {
		return nil, fmt.Errorf("listing workspaces: %w", err)
	}

	for i := range workspaces {
		if workspaces[i].Name == name {
			fmt.Printf("  Found: id=%d\n", workspaces[i].Id)
			return &workspaces[i], nil
		}
	}
	return nil, nil
}

// CreateWorkspace creates a new preview workspace with the given configuration.
func (d *Deployer) CreateWorkspace(cfg Config) (*api.Workspace, error) {
	fmt.Printf("🚀 Creating workspace '%s'...\n", cfg.Name)

	ws, err := d.Client.DeployWorkspace(api.DeployWorkspaceArgs{
		TeamId:        cfg.TeamId,
		PlanId:        cfg.PlanId,
		Name:          cfg.Name,
		EnvVars:       cfg.EnvVars,
		VpnConfigName: strPtr(cfg.VpnConfig),
		IsPrivateRepo: true,
		GitUrl:        strPtr(cfg.RepoUrl),
		Branch:        strPtr(cfg.Branch),
		Timeout:       cfg.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("creating workspace: %w", err)
	}

	fmt.Printf("  Created: id=%d\n", ws.Id)
	return ws, nil
}

// UpdateWorkspace updates an existing preview workspace by pulling the latest
// branch and setting environment variables.
func (d *Deployer) UpdateWorkspace(ws *api.Workspace, cfg Config) error {
	fmt.Println("  ⏰ Waiting for workspace to be running...")
	if err := d.Client.WaitForWorkspaceRunning(ws, cfg.Timeout); err != nil {
		return err
	}
	fmt.Println("  ✅ Workspace is running.")

	fmt.Printf("  📥 Pulling branch '%s'...\n", cfg.Branch)
	if err := d.Client.GitPull(ws.Id, "origin", cfg.Branch); err != nil {
		return fmt.Errorf("git pull: %w", err)
	}

	if len(cfg.EnvVars) > 0 {
		fmt.Printf("  🔧 Setting %d environment variable(s)...\n", len(cfg.EnvVars))
		if err := d.Client.SetEnvVarOnWorkspace(ws.Id, cfg.EnvVars); err != nil {
			return fmt.Errorf("setting env vars: %w", err)
		}
	}

	return nil
}

// DeleteWorkspace deletes a workspace by ID.
func (d *Deployer) DeleteWorkspace(wsId int) error {
	fmt.Printf("🗑️  Deleting workspace %d...\n", wsId)
	return d.Client.DeleteWorkspace(wsId)
}

// RunPipeline runs pipeline stages using the shared pipeline runner.
// The flow is: prepare → test (if present) → sync landscape → run.
func (d *Deployer) RunPipeline(wsId int, cfg Config) error {
	if len(cfg.Stages) == 0 {
		return nil
	}

	fmt.Printf("🔧 Running pipeline: %s\n", strings.Join(cfg.Stages, " → "))

	runner := pipeline.NewRunner(d.Client, nil)
	return runner.RunStages(wsId, cfg.Stages, pipeline.Config{
		Profile: cfg.Profile,
		Timeout: cfg.Timeout,
	})
}

// Deploy orchestrates the full preview environment lifecycle:
//   - If isDelete is true, finds and deletes the workspace.
//   - Otherwise, creates a new workspace or updates an existing one,
//     then runs the configured pipeline stages.
//
// Returns a Result with the workspace ID and URL on success.
func (d *Deployer) Deploy(cfg Config, isDelete bool) (*Result, error) {
	fmt.Printf("🌿 Target branch: %s\n", cfg.Branch)

	if isDelete {
		ws, err := d.FindWorkspace(cfg.TeamId, cfg.Name)
		if err != nil {
			return nil, fmt.Errorf("finding workspace: %w", err)
		}
		if ws != nil {
			if err := d.DeleteWorkspace(ws.Id); err != nil {
				return nil, fmt.Errorf("deleting workspace: %w", err)
			}
			fmt.Println("✅ Workspace deleted.")
		} else {
			fmt.Println("ℹ️  No workspace found — nothing to delete.")
		}
		return nil, nil
	}

	// Create or update
	existing, err := d.FindWorkspace(cfg.TeamId, cfg.Name)
	if err != nil {
		return nil, fmt.Errorf("finding workspace: %w", err)
	}

	var wsId int
	if existing != nil {
		if err := d.UpdateWorkspace(existing, cfg); err != nil {
			return nil, fmt.Errorf("updating workspace: %w", err)
		}
		wsId = existing.Id
		fmt.Printf("✅ Workspace %d updated.\n", wsId)
	} else {
		ws, err := d.CreateWorkspace(cfg)
		if err != nil {
			return nil, fmt.Errorf("creating workspace: %w", err)
		}
		wsId = ws.Id
		fmt.Println("✅ New workspace created.")
	}

	if err := d.RunPipeline(wsId, cfg); err != nil {
		return nil, fmt.Errorf("running pipeline: %w", err)
	}

	url := fmt.Sprintf("https://%d-3000.2.codesphere.com/", wsId)
	fmt.Printf("🔗 Deployment URL: %s\n", url)

	return &Result{
		WorkspaceId:  wsId,
		WorkspaceURL: url,
	}, nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
