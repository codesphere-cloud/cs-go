// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"fmt"

	"github.com/codesphere-cloud/cs-go/api/errors"
	"github.com/codesphere-cloud/cs-go/api/openapi_client"

	"time"
)

func (c *Client) ListWorkspaces(teamId int) ([]Workspace, error) {
	workspaces, _, err := c.api.WorkspacesAPI.WorkspacesListWorkspaces(c.ctx, float32(teamId)).Execute()
	return workspaces, errors.FormatAPIError(err)
}

func (c *Client) GetWorkspace(workspaceId int) (Workspace, error) {
	workspace, _, err := c.api.WorkspacesAPI.WorkspacesGetWorkspace(c.ctx, float32(workspaceId)).Execute()

	if workspace != nil {
		return *workspace, errors.FormatAPIError(err)
	}
	return Workspace{}, errors.FormatAPIError(err)
}

func (c *Client) DeleteWorkspace(workspaceId int) error {
	_, err := c.api.WorkspacesAPI.WorkspacesDeleteWorkspace(c.ctx, float32(workspaceId)).Execute()
	return errors.FormatAPIError(err)
}

func (c *Client) WorkspaceStatus(workspaceId int) (*WorkspaceStatus, error) {
	status, _, err := c.api.WorkspacesAPI.WorkspacesGetWorkspaceStatus(c.ctx, float32(workspaceId)).Execute()
	return status, errors.FormatAPIError(err)
}

func (c *Client) CreateWorkspace(args CreateWorkspaceArgs) (*Workspace, error) {
	workspace, _, err := c.api.WorkspacesAPI.WorkspacesCreateWorkspace(c.ctx).WorkspacesCreateWorkspaceRequest(args).Execute()
	return workspace, errors.FormatAPIError(err)
}

func (c *Client) SetEnvVarOnWorkspace(workspaceId int, envVars map[string]string) error {
	vars := []openapi_client.WorkspacesListEnvVars200ResponseInner{}
	for k, v := range envVars {
		vars = append(vars, openapi_client.WorkspacesListEnvVars200ResponseInner{
			Name:  k,
			Value: v,
		})
	}

	req := c.api.WorkspacesAPI.WorkspacesSetEnvVar(c.ctx, float32(workspaceId)).WorkspacesListEnvVars200ResponseInner(vars)
	_, err := c.api.WorkspacesAPI.WorkspacesSetEnvVarExecute(req)
	return errors.FormatAPIError(err)
}

func (c *Client) ExecCommand(workspaceId int, command string, workdir string, env map[string]string) (string, string, error) {

	workdirP := &workdir
	if workdir == "" {
		workdirP = nil
	}
	cmd := openapi_client.WorkspacesExecuteCommandRequest{
		Command:    command,
		WorkingDir: workdirP,
		Env:        &env,
	}

	req := c.api.WorkspacesAPI.WorkspacesExecuteCommand(c.ctx, float32(workspaceId)).WorkspacesExecuteCommandRequest(cmd)
	res, _, err := req.Execute()

	if err != nil {
		return "", "", errors.FormatAPIError(err)
	}
	if res == nil {
		return "", "", errors.FormatAPIError(err)
	}
	return res.Output, res.Error, errors.FormatAPIError(err)
}

func (c *Client) DeployLandscape(wsId int, profile string) error {
	if profile == "ci.yml" || profile == "" {
		req := c.api.WorkspacesAPI.WorkspacesDeployLandscape(c.ctx, float32(wsId))
		_, err := req.Execute()
		return errors.FormatAPIError(err)
	}
	req := c.api.WorkspacesAPI.WorkspacesDeployLandscape1(c.ctx, float32(wsId), profile)
	_, err := req.Execute()
	return errors.FormatAPIError(err)
}

func (c *Client) StartPipelineStage(wsId int, profile string, stage string) error {
	if profile == "ci.yml" || profile == "" {
		req := c.api.WorkspacesAPI.WorkspacesStartPipelineStage(c.ctx, float32(wsId), stage)
		_, err := req.Execute()
		return errors.FormatAPIError(err)
	}
	req := c.api.WorkspacesAPI.WorkspacesStartPipelineStage1(c.ctx, float32(wsId), stage, profile)
	_, err := req.Execute()
	return errors.FormatAPIError(err)
}

func (c *Client) GetPipelineState(wsId int, stage string) ([]PipelineStatus, error) {
	req := c.api.WorkspacesAPI.WorkspacesPipelineStatus(c.ctx, float32(wsId), stage)
	res, _, err := req.Execute()
	return res, errors.FormatAPIError(err)
}

// Waits for a given workspace to be running.
//
// Returns [TimedOut] error if the workspace does not become running in time.
func (client *Client) WaitForWorkspaceRunning(workspace *Workspace, timeout time.Duration) error {
	delay := 5 * time.Second

	maxWaitTime := client.time.Now().Add(timeout)
	for {
		status, err := client.WorkspaceStatus(workspace.Id)

		if err != nil {
			// TODO: log error and retry until timeout is reached.
			return errors.FormatAPIError(err)
		}
		if status.IsRunning {
			return nil
		}
		if client.time.Now().After(maxWaitTime) {
			break
		}
		client.time.Sleep(delay)
	}

	return errors.TimedOut(
		fmt.Sprintf("waiting for workspace %s(%d) to be ready", workspace.Name, workspace.Id),
		timeout)
}

type DeployWorkspaceArgs struct {
	TeamId        int
	PlanId        int
	Name          string
	EnvVars       map[string]string
	VpnConfigName *string //must be nil to use default

	IsPrivateRepo bool
	GitUrl        *string //must be nil to use default
	Branch        *string //must be nil to use default

	Timeout time.Duration
}

// Deploys a workspace with the given configuration.
//
// Returns [TimedOut] error if the timeout is reached
func (client Client) DeployWorkspace(args DeployWorkspaceArgs) (*Workspace, error) {
	workspace, err := client.CreateWorkspace(CreateWorkspaceArgs{
		TeamId:            args.TeamId,
		Name:              args.Name,
		PlanId:            args.PlanId,
		IsPrivateRepo:     args.IsPrivateRepo,
		GitUrl:            args.GitUrl,
		InitialBranch:     args.Branch,
		SourceWorkspaceId: nil,
		WelcomeMessage:    nil,
		Replicas:          1,
		VpnConfig:         args.VpnConfigName,
	})
	if err != nil {
		return nil, err
	}
	if err := client.WaitForWorkspaceRunning(workspace, args.Timeout); err != nil {
		return workspace, err
	}

	if len(args.EnvVars) != 0 {
		if err := client.SetEnvVarOnWorkspace(workspace.Id, args.EnvVars); err != nil {
			return workspace, err
		}
	}
	return workspace, nil
}

func (c Client) GitPull(workspaceId int, remote string, branch string) error {
	if remote == "" {
		req := c.api.WorkspacesAPI.WorkspacesGitPull(c.ctx, float32(workspaceId))
		_, err := req.Execute()
		return errors.FormatAPIError(err)
	}

	req := c.api.WorkspacesAPI.WorkspacesGitPull2(c.ctx, float32(workspaceId), remote, branch)
	_, err := req.Execute()
	return errors.FormatAPIError(err)
}
