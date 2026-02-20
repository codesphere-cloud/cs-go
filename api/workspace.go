// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

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
	vars := []openapi_client.WorkspacesCreateWorkspaceRequestEnvInner{}
	for k, v := range envVars {
		vars = append(vars, openapi_client.WorkspacesCreateWorkspaceRequestEnvInner{
			Name:  k,
			Value: v,
		})
	}

	req := c.api.WorkspacesAPI.WorkspacesSetEnvVar(c.ctx, float32(workspaceId)).
		WorkspacesCreateWorkspaceRequestEnvInner(vars)
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

// ScaleWorkspace sets the number of replicas for a workspace.
// For on-demand workspaces, setting replicas to 1 wakes up the workspace.
func (c *Client) ScaleWorkspace(wsId int, replicas int) error {
	req := c.api.WorkspacesAPI.WorkspacesUpdateWorkspace(c.ctx, float32(wsId)).
		WorkspacesUpdateWorkspaceRequest(openapi_client.WorkspacesUpdateWorkspaceRequest{
			Replicas: &replicas,
		})
	_, err := req.Execute()
	return errors.FormatAPIError(err)
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
			if client.time.Now().After(maxWaitTime) {
				return errors.FormatAPIError(err)
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
func (client Client) DeployWorkspace(args DeployWorkspaceArgs) (*Workspace, error) {
	workspace, err := client.CreateWorkspace(CreateWorkspaceArgs{
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

// logEntry represents a single log line from the SSE stream.
type logEntry struct {
	Timestamp string `json:"timestamp"`
	Kind      string `json:"kind"`
	Data      string `json:"data"`
}

// StreamLogs connects to the Codesphere SSE log endpoint and writes parsed
// log entries to the provided writer until the context is cancelled or the
// stream ends. This is used during pipeline execution to provide real-time
// log output.
func (c *Client) StreamLogs(ctx context.Context, apiUrl string, wsId int, stage string, step int, w io.Writer) error {
	endpoint := fmt.Sprintf("%s/workspaces/%d/logs/%s/%d", apiUrl, wsId, stage, step)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to construct log stream request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")

	// Set auth from the client's context token
	if token, ok := ctx.Value(openapi_client.ContextAccessToken).(string); ok && token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	} else if token, ok := c.ctx.Value(openapi_client.ContextAccessToken).(string); ok && token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// Context cancellation is expected when the stage finishes
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("failed to connect to log stream: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("log stream responded with status %d", resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)

	for {
		// Check if context is done
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		// Parse one SSE event
		var eventData string
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if ctx.Err() != nil || err == io.EOF {
					return nil
				}
				return fmt.Errorf("failed to read log stream: %w", err)
			}

			line = strings.TrimSpace(line)

			if strings.HasPrefix(line, "data:") {
				data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
				if eventData != "" {
					eventData += "\n" + data
				} else {
					eventData = data
				}
			} else if line == "" && eventData != "" {
				// Empty line marks end of SSE event
				break
			}
		}

		// Parse and print log entries
		var entries []logEntry
		if err := json.Unmarshal([]byte(eventData), &entries); err != nil {
			// Skip unparseable events (e.g. error responses)
			log.Printf("⚠ log stream: %s", eventData)
			eventData = ""
			continue
		}

		for _, entry := range entries {
			_, _ = fmt.Fprintf(w, "%s | %s\n", entry.Timestamp, entry.Data)
		}
		eventData = ""
	}
}
