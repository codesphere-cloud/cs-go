// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

//go:generate mockery

import (
	"context"
	"fmt"
	"net/url"

	"github.com/codesphere-cloud/cs-go/api"
)

type Client interface {
	ListTeams() ([]api.Team, error)
	ListWorkspaces(teamId int) ([]api.Workspace, error)
	GetWorkspace(workspaceId int) (api.Workspace, error)
	SetEnvVarOnWorkspace(workspaceId int, vars map[string]string) error
	ExecCommand(workspaceId int, command string, workdir string, env map[string]string) (string, string, error)
	ListWorkspacePlans() ([]api.WorkspacePlan, error)
	DeployWorkspace(args api.DeployWorkspaceArgs) (*api.Workspace, error)
	DeleteWorkspace(wsId int) error
	StartPipelineStage(wsId int, profile string, stage string) error
	GetPipelineState(wsId int, stage string) ([]api.PipelineStatus, error)
}

func NewClient(opts GlobalOptions) (Client, error) {
	token, err := opts.Env.GetApiToken()
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
	})
	return client, nil
}
