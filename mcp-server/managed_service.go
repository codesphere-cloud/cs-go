// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListManagedServicesArgs struct {
	TeamId         int  `json:"teamId" jsonschema:"ID of the team to list managed services for"`
	IncludeDeleted bool `json:"includeDeleted,omitempty" jsonschema:"Whether to include deleted managed services"`
}

type ManagedServiceIdArgs struct {
	ServiceId string `json:"serviceId" jsonschema:"UUID of the managed service"`
}

type ManagedServicePlanArg struct {
	Id         int            `json:"id" jsonschema:"Plan ID"`
	Parameters map[string]int `json:"parameters" jsonschema:"Plan parameters, e.g. storage or CPU sizes"`
}

type CreateManagedServiceArgs struct {
	TeamId          int                    `json:"teamId" jsonschema:"ID of the team to create the managed service in"`
	Name            string                 `json:"name" jsonschema:"Name of the managed service"`
	ProviderName    string                 `json:"providerName" jsonschema:"Provider name, e.g. postgres (see list_managed_service_providers)"`
	ProviderVersion *string                `json:"providerVersion,omitempty" jsonschema:"Optional provider version, e.g. v1"`
	Plan            ManagedServicePlanArg  `json:"plan" jsonschema:"Plan of the managed service"`
	Config          map[string]interface{} `json:"config" jsonschema:"Provider specific configuration object"`
	Secrets         map[string]interface{} `json:"secrets" jsonschema:"Provider specific secrets object"`
	Version         *string                `json:"version,omitempty" jsonschema:"Optional semantic version of the service"`
}

type UpdateManagedServiceArgs struct {
	ServiceId string                 `json:"serviceId" jsonschema:"UUID of the managed service"`
	Name      *string                `json:"name,omitempty" jsonschema:"New name"`
	Pause     *bool                  `json:"pause,omitempty" jsonschema:"Pause (true) or resume (false) the service"`
	Plan      *ManagedServicePlanArg `json:"plan,omitempty" jsonschema:"New plan"`
	Config    map[string]interface{} `json:"config,omitempty" jsonschema:"Updated provider specific configuration"`
	Secrets   map[string]interface{} `json:"secrets,omitempty" jsonschema:"Updated provider specific secrets"`
	Version   *string                `json:"version,omitempty" jsonschema:"New semantic version of the service"`
}

func (p ManagedServicePlanArg) toAPI() api.ManagedServicePlan {
	return api.ManagedServicePlan{Id: p.Id, Parameters: p.Parameters}
}

func RegisterManagedServiceTools(server *mcp.Server, client *api.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_managed_services",
		Description: "List managed services (e.g. databases, queues) of a team",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ListManagedServicesArgs) (*mcp.CallToolResult, any, error) {
		services, err := client.ListManagedServices(args.TeamId, args.IncludeDeleted)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, itemsResult(services), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_managed_service",
		Description: "Get details of a single managed service",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ManagedServiceIdArgs) (*mcp.CallToolResult, any, error) {
		details, err := client.GetManagedService(args.ServiceId)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, details, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_managed_service_providers",
		Description: "List available managed service providers and their versions",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
		providers, err := client.ListManagedServiceProviders()
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, itemsResult(providers), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_managed_service",
		Description: "Create a new managed service. May incur costs.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args CreateManagedServiceArgs) (*mcp.CallToolResult, any, error) {
		provider := api.NewManagedServiceProviderRef(args.ProviderName, args.ProviderVersion)
		create := api.CreateManagedServiceArgs{
			TeamId:   args.TeamId,
			Name:     args.Name,
			Provider: provider,
			Plan:     args.Plan.toAPI(),
			Config:   args.Config,
			Secrets:  args.Secrets,
			Version:  args.Version,
		}
		svc, err := client.CreateManagedService(create)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, svc, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_managed_service",
		Description: "Update a managed service (rename, pause/resume, change plan, config, secrets or version)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args UpdateManagedServiceArgs) (*mcp.CallToolResult, any, error) {
		update := api.UpdateManagedServiceArgs{
			Name:    args.Name,
			Pause:   args.Pause,
			Config:  args.Config,
			Secrets: args.Secrets,
			Version: args.Version,
		}
		if args.Plan != nil {
			plan := args.Plan.toAPI()
			update.Plan = &plan
		}
		svc, err := client.UpdateManagedService(args.ServiceId, update)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, svc, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_managed_service",
		Description: "Delete a managed service by ID. This is destructive.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ManagedServiceIdArgs) (*mcp.CallToolResult, any, error) {
		if err := client.DeleteManagedService(args.ServiceId); err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, map[string]string{"status": "deleted"}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "schedule_managed_service_backup",
		Description: "Schedule a backup of a managed service",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ManagedServiceIdArgs) (*mcp.CallToolResult, any, error) {
		backup, err := client.ScheduleManagedServiceBackup(args.ServiceId)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, backup, nil
	})
}
