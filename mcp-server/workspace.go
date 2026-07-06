// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListWorkspacesArgs struct {
	TeamId int `json:"teamId" jsonschema:"ID of the team to list workspaces for"`
}

type GetWorkspaceArgs struct {
	WorkspaceId int `json:"workspaceId" jsonschema:"ID of the workspace to get"`
}

type DeleteWorkspaceArgs struct {
	WorkspaceId int `json:"workspaceId" jsonschema:"ID of the workspace to delete"`
}

type WorkspaceStatusArgs struct {
	WorkspaceId int `json:"workspaceId" jsonschema:"ID of the workspace"`
}

type CreateWorkspaceInput struct {
	TeamId        int     `json:"teamId" jsonschema:"ID of the team to create the workspace in"`
	Name          string  `json:"name" jsonschema:"Name of the workspace"`
	PlanId        int     `json:"planId" jsonschema:"Plan ID for the workspace (e.g. 1 for basic)"`
	IsPrivateRepo bool    `json:"isPrivateRepo" jsonschema:"Whether the repository is private"`
	Replicas      int     `json:"replicas" jsonschema:"Number of replicas, usually 1"`
	BaseImage     *string `json:"baseImage,omitempty" jsonschema:"Optional base image name"`
	GitUrl        *string `json:"gitUrl,omitempty" jsonschema:"Optional git repository URL"`
	InitialBranch *string `json:"initialBranch,omitempty" jsonschema:"Optional initial branch"`
}

type SetEnvVarArgs struct {
	WorkspaceId int               `json:"workspaceId" jsonschema:"ID of the workspace"`
	EnvVars     map[string]string `json:"envVars" jsonschema:"Key-value map of environment variables"`
}

type ExecCommandArgs struct {
	WorkspaceId int               `json:"workspaceId" jsonschema:"ID of the workspace"`
	Command     string            `json:"command" jsonschema:"Command to execute"`
	Workdir     string            `json:"workdir" jsonschema:"Working directory"`
	Env         map[string]string `json:"env" jsonschema:"Environment variables to pass"`
}

type DeployLandscapeArgs struct {
	WorkspaceId int    `json:"workspaceId" jsonschema:"ID of the workspace"`
	Profile     string `json:"profile" jsonschema:"The landscape profile"`
}

type StartPipelineStageArgs struct {
	WorkspaceId int    `json:"workspaceId" jsonschema:"ID of the workspace"`
	Profile     string `json:"profile" jsonschema:"Profile name"`
	Stage       string `json:"stage" jsonschema:"Stage name"`
}

type StopPipelineStageArgs struct {
	WorkspaceId int    `json:"workspaceId" jsonschema:"ID of the workspace"`
	Stage       string `json:"stage" jsonschema:"Stage name"`
}

type GetPipelineStateArgs struct {
	WorkspaceId int    `json:"workspaceId" jsonschema:"ID of the workspace"`
	Stage       string `json:"stage" jsonschema:"Stage name"`
}

type ScaleWorkspaceArgs struct {
	WorkspaceId int `json:"workspaceId" jsonschema:"ID of the workspace"`
	Replicas    int `json:"replicas" jsonschema:"Number of replicas"`
}

type ScaleLandscapeServicesArgs struct {
	WorkspaceId int            `json:"workspaceId" jsonschema:"ID of the workspace"`
	Services    map[string]int `json:"services" jsonschema:"Map of service name to replica count"`
}

type GetLogsOfStageArgs struct {
	WorkspaceId int    `json:"workspaceId" jsonschema:"ID of the workspace"`
	Stage       string `json:"stage" jsonschema:"Stage name, like 'run'"`
	Step        int    `json:"step" jsonschema:"Index of execution step (default 0)"`
}

type GetLogsOfReplicaArgs struct {
	WorkspaceId int    `json:"workspaceId" jsonschema:"ID of the workspace"`
	Step        int    `json:"step" jsonschema:"Index of execution step (default 0)"`
	Replica     string `json:"replica" jsonschema:"ID of server replica"`
}

type GetLogsOfServerArgs struct {
	WorkspaceId int    `json:"workspaceId" jsonschema:"ID of the workspace"`
	Step        int    `json:"step" jsonschema:"Index of execution step (default 0)"`
	Server      string `json:"server" jsonschema:"Name of the landscape server"`
}

func RegisterWorkspaceTools(server *mcp.Server, client *api.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_workspaces",
		Description: "List all workspaces for a given team",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ListWorkspacesArgs) (*mcp.CallToolResult, any, error) {
		workspaces, err := client.ListWorkspaces(args.TeamId)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, map[string]any{"items": workspaces}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_workspace",
		Description: "Get details for a specific workspace",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetWorkspaceArgs) (*mcp.CallToolResult, any, error) {
		workspace, err := client.GetWorkspace(args.WorkspaceId)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, workspace, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_workspace",
		Description: "Delete a workspace by ID",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args DeleteWorkspaceArgs) (*mcp.CallToolResult, any, error) {
		err := client.DeleteWorkspace(args.WorkspaceId)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, map[string]string{"status": "deleted"}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "workspace_status",
		Description: "Get the runtime status of a workspace",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args WorkspaceStatusArgs) (*mcp.CallToolResult, any, error) {
		status, err := client.WorkspaceStatus(args.WorkspaceId)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, status, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_workspace",
		Description: "Create a new workspace",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args CreateWorkspaceInput) (*mcp.CallToolResult, any, error) {
		apiArgs := api.CreateWorkspaceArgs{
			TeamId:        args.TeamId,
			Name:          args.Name,
			PlanId:        args.PlanId,
			IsPrivateRepo: args.IsPrivateRepo,
			Replicas:      args.Replicas,
			BaseImage:     args.BaseImage,
			GitUrl:        args.GitUrl,
			InitialBranch: args.InitialBranch,
		}
		workspace, err := client.CreateWorkspace(apiArgs)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, workspace, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_workspace_env_var",
		Description: "Set environment variables on a workspace",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args SetEnvVarArgs) (*mcp.CallToolResult, any, error) {
		err := client.SetEnvVarOnWorkspace(args.WorkspaceId, args.EnvVars)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, map[string]string{"status": "success"}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "exec_command",
		Description: "Execute a command inside a workspace",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ExecCommandArgs) (*mcp.CallToolResult, any, error) {
		stdout, stderr, err := client.ExecCommand(args.WorkspaceId, args.Command, args.Workdir, args.Env)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, map[string]string{"stdout": stdout, "stderr": stderr}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "deploy_landscape",
		Description: "Deploy a workspace landscape",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args DeployLandscapeArgs) (*mcp.CallToolResult, any, error) {
		err := client.DeployLandscape(args.WorkspaceId, args.Profile)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, map[string]string{"status": "started"}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "start_pipeline_stage",
		Description: "Start a CI pipeline stage",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args StartPipelineStageArgs) (*mcp.CallToolResult, any, error) {
		err := client.StartPipelineStage(args.WorkspaceId, args.Profile, args.Stage)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, map[string]string{"status": "started"}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "stop_pipeline_stage",
		Description: "Stop a CI pipeline stage",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args StopPipelineStageArgs) (*mcp.CallToolResult, any, error) {
		err := client.StopPipelineStage(args.WorkspaceId, args.Stage)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, map[string]string{"status": "stopped"}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_pipeline_state",
		Description: "Get the status of a pipeline stage",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetPipelineStateArgs) (*mcp.CallToolResult, any, error) {
		states, err := client.GetPipelineState(args.WorkspaceId, args.Stage)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, map[string]any{"items": states}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "scale_workspace",
		Description: "Scale the number of replicas for a workspace",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ScaleWorkspaceArgs) (*mcp.CallToolResult, any, error) {
		err := client.ScaleWorkspace(args.WorkspaceId, args.Replicas)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, map[string]string{"status": "scaled"}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "scale_landscape_services",
		Description: "Scale specific services in a landscape",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ScaleLandscapeServicesArgs) (*mcp.CallToolResult, any, error) {
		err := client.ScaleLandscapeServices(args.WorkspaceId, args.Services)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, map[string]string{"status": "scaled"}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_logs_of_stage",
		Description: "Retrieve logs of a workspace by stage",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetLogsOfStageArgs) (*mcp.CallToolResult, any, error) {
		logs, err := client.GetLogsOfStage(args.WorkspaceId, args.Stage, args.Step)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, logs, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_logs_of_replica",
		Description: "Retrieve logs of a workspace by replica",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetLogsOfReplicaArgs) (*mcp.CallToolResult, any, error) {
		logs, err := client.GetLogsOfReplica(args.WorkspaceId, args.Step, args.Replica)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, logs, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_logs_of_server",
		Description: "Retrieve logs of a workspace by server",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetLogsOfServerArgs) (*mcp.CallToolResult, any, error) {
		logs, err := client.GetLogsOfServer(args.WorkspaceId, args.Step, args.Server)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, logs, nil
	})
}
