package api

import (
	"fmt"

	"github.com/codesphere-cloud/cs-go/api/openapi_client"
	"github.com/codesphere-cloud/cs-go/pkg/errors"

	"time"
)

type WaitForWorkspaceRunningOptions struct {
	Timeout time.Duration
	Delay   time.Duration
}

func (c *Client) ListWorkspaces(teamId int) ([]Workspace, error) {
	req := c.api.WorkspacesAPI.WorkspacesListWorkspaces(c.ctx, float32(teamId))
	workspaces, _, err := c.api.WorkspacesAPI.WorkspacesListWorkspacesExecute(req)
	return workspaces, err
}

func (c *Client) WorkspaceStatus(workspaceId int) (*WorkspaceStatus, error) {
	req := c.api.WorkspacesAPI.WorkspacesGetWorkspaceStatus(c.ctx, float32(workspaceId))
	status, _, err := c.api.WorkspacesAPI.WorkspacesGetWorkspaceStatusExecute(req)
	return status, err
}

func (c *Client) CreateWorkspace(args CreateWorkspaceArgs) (*Workspace, error) {
	req := c.api.WorkspacesAPI.WorkspacesCreateWorkspace(c.ctx).WorkspacesCreateWorkspaceRequest(args)
	workspace, _, err := c.api.WorkspacesAPI.
		WorkspacesCreateWorkspaceExecute(req)
	return workspace, err
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
	return err
}

// Waits for a given workspace to be running.
//
// Returns [TimedOut] error if the workspace does not become running in time.
func WaitForWorkspaceRunning(
	client *Client,
	workspace *Workspace,
	opts WaitForWorkspaceRunningOptions,
) error {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 20 * time.Minute
	}
	delay := opts.Delay
	if delay == 0 {
		delay = 5 * time.Second
	}

	maxWaitTime := time.Now().Add(timeout)
	for time.Now().Before(maxWaitTime) {
		status, err := client.WorkspaceStatus(workspace.Id)

		if err != nil {
			// TODO: log error and retry until timeout is reached.
			return err
		}
		if status.IsRunning {
			return nil
		}
		time.Sleep(delay)
	}

	return errors.NewTimedOut(
		fmt.Sprintf("Waiting for workspace %s(%d) to be ready", workspace.Name, workspace.Id),
		timeout)
}

type DeployWorkspaceArgs struct {
	TeamId        int
	PlanId        int
	Name          string
	EnvVars       map[string]string
	VpnConfigName *string

	Timeout time.Duration
}

// Deploys a workspace with the given configuration.
//
// Returns [TimedOut] error if the timeout is reached
func DeployWorkspace(
	client Client,
	args DeployWorkspaceArgs,
) error {
	workspace, err := client.CreateWorkspace(CreateWorkspaceArgs{
		TeamId:            args.TeamId,
		Name:              args.Name,
		PlanId:            args.PlanId,
		IsPrivateRepo:     true,
		GitUrl:            nil,
		InitialBranch:     nil,
		SourceWorkspaceId: nil,
		WelcomeMessage:    nil,
		Replicas:          1,
		VpnConfig:         args.VpnConfigName,
	})
	if err != nil {
		return err
	}
	if err := WaitForWorkspaceRunning(&client, workspace, WaitForWorkspaceRunningOptions{Timeout: args.Timeout}); err != nil {
		return err
	}

	if len(args.EnvVars) != 0 {
		if err := client.SetEnvVarOnWorkspace(workspace.Id, args.EnvVars); err != nil {
			return err
		}
	}
	return nil
}
