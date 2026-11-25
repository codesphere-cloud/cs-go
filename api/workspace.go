// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"fmt"
	"time"

	"github.com/codesphere-cloud/cs-go/api/errors"
	"github.com/codesphere-cloud/cs-go/api/openapi_client"
)

func (c *Client) ListWorkspaces(ctx context.Context, teamId int) ([]Workspace, error) {
	workspaces, r, err := c.api.WorkspacesAPI.WorkspacesListWorkspaces(c.withAuth(ctx), teamId).Execute()
	return workspaces, errors.FormatAPIError(r, err)
}

func (c *Client) GetWorkspace(ctx context.Context, workspaceId int) (Workspace, error) {
	workspace, r, err := c.api.WorkspacesAPI.WorkspacesGetWorkspace(c.withAuth(ctx), workspaceId).Execute()

	if workspace != nil {
		return *workspace, errors.FormatAPIError(r, err)
	}
	return Workspace{}, errors.FormatAPIError(r, err)
}

func (c *Client) DeleteWorkspace(ctx context.Context, workspaceId int) error {
	r, err := c.api.WorkspacesAPI.WorkspacesDeleteWorkspace(c.withAuth(ctx), workspaceId).Execute()
	return errors.FormatAPIError(r, err)
}

func (c *Client) WorkspaceStatus(ctx context.Context, workspaceId int) (*WorkspaceStatus, error) {
	status, r, err := c.api.WorkspacesAPI.WorkspacesGetWorkspaceStatus(c.withAuth(ctx), workspaceId).Execute()
	return status, errors.FormatAPIError(r, err)
}

func (c *Client) CreateWorkspace(ctx context.Context, args CreateWorkspaceArgs) (*Workspace, error) {
	workspace, r, err := c.api.WorkspacesAPI.WorkspacesCreateWorkspace(c.withAuth(ctx)).WorkspacesCreateWorkspaceRequest(args).Execute()
	return workspace, errors.FormatAPIError(r, err)
}

func (c *Client) SetEnvVarOnWorkspace(ctx context.Context, workspaceId int, envVars map[string]string) error {
	vars := []openapi_client.WorkspacesCreateWorkspaceRequestEnvInner{}
	for k, v := range envVars {
		vars = append(vars, openapi_client.WorkspacesCreateWorkspaceRequestEnvInner{
			Name:  k,
			Value: v,
		})
	}

	req := c.api.WorkspacesAPI.WorkspacesSetEnvVar(c.withAuth(ctx), workspaceId).
		WorkspacesCreateWorkspaceRequestEnvInner(vars)
	r, err := c.api.WorkspacesAPI.WorkspacesSetEnvVarExecute(req)
	return errors.FormatAPIError(r, err)
}

func (c *Client) ExecCommand(ctx context.Context, workspaceId int, command string, workdir string, env map[string]string) (string, string, error) {

	workdirP := &workdir
	if workdir == "" {
		workdirP = nil
	}
	cmd := openapi_client.WorkspacesExecuteCommandRequest{
		Command:    command,
		WorkingDir: workdirP,
		Env:        &env,
	}

	req := c.api.WorkspacesAPI.WorkspacesExecuteCommand(c.withAuth(ctx), workspaceId).WorkspacesExecuteCommandRequest(cmd)
	res, r, err := req.Execute()

	if err != nil {
		return "", "", errors.FormatAPIError(r, err)
	}
	if res == nil {
		return "", "", errors.FormatAPIError(r, err)
	}
	return res.Output, res.Error, errors.FormatAPIError(r, err)
}

func (c *Client) DeployLandscape(ctx context.Context, wsId int, profile string) error {
	if profile == "ci.yml" || profile == "" {
		req := c.api.WorkspacesAPI.WorkspacesDeployLandscape(c.withAuth(ctx), wsId)
		r, err := req.Execute()
		return errors.FormatAPIError(r, err)
	}
	req := c.api.WorkspacesAPI.WorkspacesDeployLandscape1(c.withAuth(ctx), wsId, profile)
	r, err := req.Execute()
	return errors.FormatAPIError(r, err)
}

func (c *Client) StartPipelineStage(ctx context.Context, wsId int, profile string, stage string) error {
	if profile == "ci.yml" || profile == "" {
		req := c.api.WorkspacesAPI.WorkspacesStartPipelineStage(c.withAuth(ctx), wsId, stage)
		r, err := req.Execute()
		return errors.FormatAPIError(r, err)
	}
	req := c.api.WorkspacesAPI.WorkspacesStartPipelineStage1(c.withAuth(ctx), wsId, stage, profile)
	r, err := req.Execute()
	return errors.FormatAPIError(r, err)
}

// StopPipelineStage stops the currently running pipeline, so selection of a pipeline is not supported.
func (c *Client) StopPipelineStage(ctx context.Context, wsId int, stage string) error {
	req := c.api.WorkspacesAPI.WorkspacesStopPipelineStage(c.withAuth(ctx), wsId, stage)
	r, err := req.Execute()
	return errors.FormatAPIError(r, err)
}

func (c *Client) GetPipelineState(ctx context.Context, wsId int, stage string) ([]PipelineStatus, error) {
	req := c.api.WorkspacesAPI.WorkspacesPipelineStatus(c.withAuth(ctx), wsId, stage)
	res, r, err := req.Execute()
	return res, errors.FormatAPIError(r, err)
}

// ScaleWorkspace sets the number of replicas for a workspace.
// For on-demand workspaces, setting replicas to 1 wakes up the workspace.
func (c *Client) ScaleWorkspace(ctx context.Context, wsId int, replicas int) error {
	req := c.api.WorkspacesAPI.WorkspacesUpdateWorkspace(c.withAuth(ctx), wsId).
		WorkspacesUpdateWorkspaceRequest(openapi_client.WorkspacesUpdateWorkspaceRequest{
			Replicas: &replicas,
		})
	r, err := req.Execute()
	return errors.FormatAPIError(r, err)
}

// ScaleLandscapeServices scales landscape services by name.
// The services map contains service name -> replica count.
func (c *Client) ScaleLandscapeServices(ctx context.Context, wsId int, services map[string]int) error {
	req := c.api.WorkspacesAPI.WorkspacesScaleLandscapeServices(c.withAuth(ctx), wsId).
		RequestBody(services)
	resp, err := req.Execute()
	return errors.FormatAPIError(resp, err)
}

// Waits for a given workspace to be running.
//
// Returns [TimedOut] error if the workspace does not become running in time.
func (client *Client) WaitForWorkspaceRunning(ctx context.Context, workspace *Workspace, timeout time.Duration) error {
	delay := 5 * time.Second

	maxWaitTime := client.time.Now().Add(timeout)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		status, err := client.WorkspaceStatus(ctx, workspace.Id)

		if err != nil {
			if client.time.Now().After(maxWaitTime) {
				return err
			}
			client.time.Sleep(delay)
			continue
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
	BaseImage     *string //must be nil to use default
	Restricted    *bool   //must be nil to use default

	Timeout time.Duration
}

// Deploys a workspace with the given configuration.
//
// Returns [TimedOut] error if the timeout is reached
func (client Client) DeployWorkspace(ctx context.Context, args DeployWorkspaceArgs) (*Workspace, error) {
	createArgs := CreateWorkspaceArgs{
		TeamId:            args.TeamId,
		Name:              args.Name,
		PlanId:            args.PlanId,
		IsPrivateRepo:     args.IsPrivateRepo,
		GitUrl:            args.GitUrl,
		InitialBranch:     args.Branch,
		BaseImage:         args.BaseImage,
		Restricted:        args.Restricted,
		SourceWorkspaceId: nil,
		WelcomeMessage:    nil,
		Replicas:          1,
		VpnConfig:         args.VpnConfigName,
	}

	var workspace *Workspace
	var err error
	for attempt := 0; attempt < 5; attempt++ {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		workspace, err = client.CreateWorkspace(ctx, createArgs)
		if err == nil {
			break
		}
		if !errors.IsRetryable(err) {
			return nil, err
		}
		client.time.Sleep(time.Duration(attempt+1) * 5 * time.Second)
	}
	if err != nil {
		return nil, err
	}
	if err := client.WaitForWorkspaceRunning(ctx, workspace, args.Timeout); err != nil {
		return workspace, err
	}

	if len(args.EnvVars) != 0 {
		if err := client.SetEnvVarOnWorkspace(ctx, workspace.Id, args.EnvVars); err != nil {
			return workspace, err
		}
	}
	return workspace, nil
}

func (c Client) GitPull(ctx context.Context, workspaceId int, remote string, branch string) error {
	if remote == "" {
		req := c.api.WorkspacesAPI.WorkspacesGitPull(c.withAuth(ctx), workspaceId)
		r, err := req.Execute()
		return errors.FormatAPIError(r, err)
	}

	req := c.api.WorkspacesAPI.WorkspacesGitPull2(c.withAuth(ctx), workspaceId, remote, branch)
	r, err := req.Execute()
	return errors.FormatAPIError(r, err)
}

func (c *Client) GetLogsOfStage(ctx context.Context, workspaceId int, stage string, step int) (*openapi_client.WorkspacesLogs200Response, error) {
	resp, r, err := c.api.WorkspacesAPI.WorkspacesLogs(c.withAuth(ctx), workspaceId, stage, step).Execute()
	return resp, errors.FormatAPIError(r, err)
}

func (c *Client) GetLogsOfReplica(ctx context.Context, workspaceId int, step int, replica string) (*openapi_client.WorkspacesReplicaLogs200Response, error) {
	resp, r, err := c.api.WorkspacesAPI.WorkspacesReplicaLogs(c.withAuth(ctx), workspaceId, step, replica).Execute()
	return resp, errors.FormatAPIError(r, err)
}

func (c *Client) GetLogsOfServer(ctx context.Context, workspaceId int, step int, server string) (*openapi_client.WorkspacesServerLogs200Response, error) {
	resp, r, err := c.api.WorkspacesAPI.WorkspacesServerLogs(c.withAuth(ctx), workspaceId, step, server).Execute()
	return resp, errors.FormatAPIError(r, err)
}
