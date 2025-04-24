package cs

import (
	"fmt"
	"time"

	"github.com/codesphere-cloud/cs-go/api"
)

type WaitForWorkspaceRunningOptions struct {
	Timeout time.Duration
	Delay   time.Duration
}

// Waits for a given workspace to be running.
//
// Returns [TimedOut] error if the workspace does not become running in time.
func WaitForWorkspaceRunning(
	client *api.Client,
	workspace *api.Workspace,
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

	return NewTimedOut(
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
	client api.Client,
	args DeployWorkspaceArgs,
) error {
	workspace, err := client.CreateWorkspace(api.CreateWorkspaceArgs{
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
