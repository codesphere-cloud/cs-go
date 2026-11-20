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
	Id         int            `json:"id" jsonschema:"Plan ID, one of the provider's plans"`
	Parameters map[string]int `json:"parameters" jsonschema:"Plan parameters, e.g. storage or CPU sizes"`
}

type CreateManagedServiceArgs struct {
	TeamId                int                    `json:"teamId" jsonschema:"ID of the team to create the managed service in"`
	Name                  string                 `json:"name" jsonschema:"Name of the managed service"`
	ProviderName          string                 `json:"providerName" jsonschema:"Provider name, e.g. postgres (name from list_managed_service_providers)"`
	ProviderSchemaVersion string                 `json:"providerSchemaVersion" jsonschema:"Provider schema version, e.g. v1 (schemaVersion from list_managed_service_providers)"`
	Plan                  ManagedServicePlanArg  `json:"plan" jsonschema:"Plan of the managed service"`
	Config                map[string]interface{} `json:"config" jsonschema:"Provider specific configuration, must match the provider's configSchema"`
	Secrets               map[string]interface{} `json:"secrets" jsonschema:"Provider specific secrets, must match the provider's secretsSchema. Values pass through the model context and the client's tool-call history"`
	Version               *string                `json:"version,omitempty" jsonschema:"Optional release to deploy, a key of the provider's versions. Omit it if the provider defines no versions"`
}

type UpdateManagedServiceArgs struct {
	ServiceId string                 `json:"serviceId" jsonschema:"UUID of the managed service"`
	Name      *string                `json:"name,omitempty" jsonschema:"New name"`
	Pause     *bool                  `json:"pause,omitempty" jsonschema:"Pause (true) or resume (false) the service"`
	Plan      *ManagedServicePlanArg `json:"plan,omitempty" jsonschema:"New plan. Replaces the current plan including all of its parameters"`
	Config    map[string]interface{} `json:"config,omitempty" jsonschema:"Complete provider configuration. Replaces the stored config as a whole (no merge), so keys you leave out are removed. Start from the config returned by list_managed_services"`
	Secrets   map[string]interface{} `json:"secrets,omitempty" jsonschema:"Complete provider secrets. Replaces all stored secrets (no merge); stored secrets cannot be read back, so include every secret. Values pass through the model context and the client's tool-call history"`
	Version   *string                `json:"version,omitempty" jsonschema:"Release to deploy, a key of the provider's versions. Changing it upgrades the service"`
}

func (p ManagedServicePlanArg) toAPI() api.ManagedServicePlan {
	return api.ManagedServicePlan{Id: p.Id, Parameters: p.Parameters}
}

func RegisterManagedServiceTools(server *mcp.Server, client *api.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_managed_services",
		Description: "List managed services (e.g. databases, queues) of a team, including their provider, plan, config and status",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ListManagedServicesArgs) (*mcp.CallToolResult, any, error) {
		services, err := client.ListManagedServices(args.TeamId, args.IncludeDeleted)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, itemsResult(services), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_managed_service",
		Description: "Get provider-reported details of a managed service, e.g. hostname and service URLs. The fields are defined by the provider; use list_managed_services for name, plan, config and status",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ManagedServiceIdArgs) (*mcp.CallToolResult, any, error) {
		details, err := client.GetManagedService(args.ServiceId)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, details, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_managed_service_providers",
		Description: "List available managed service providers with their schemaVersion, plans, versions and config, secrets and details schemas",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
		providers, err := client.ListManagedServiceProviders()
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, itemsResult(providers), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "create_managed_service",
		Description: "Create a new managed service. May incur costs. Look up the provider with list_managed_service_providers first. " +
			"Not idempotent: if a call fails or times out, check list_managed_services before retrying to avoid creating a second service",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: new(false)},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args CreateManagedServiceArgs) (*mcp.CallToolResult, any, error) {
		create := api.CreateManagedServiceArgs{
			TeamId:   args.TeamId,
			Name:     args.Name,
			Provider: api.NewManagedServiceProviderRef(args.ProviderName, args.ProviderSchemaVersion),
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
		Description: "Update a managed service (rename, pause/resume, change plan, config, secrets or version). Plan, config and secrets replace the stored values as a whole",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: new(true)},
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
		Annotations: &mcp.ToolAnnotations{DestructiveHint: new(true)},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ManagedServiceIdArgs) (*mcp.CallToolResult, any, error) {
		if err := client.DeleteManagedService(args.ServiceId); err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, map[string]string{"status": "deleted"}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "schedule_managed_service_backup",
		Description: "Schedule a backup of a managed service",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: new(false)},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ManagedServiceIdArgs) (*mcp.CallToolResult, any, error) {
		backup, err := client.ScheduleManagedServiceBackup(args.ServiceId)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, backup, nil
	})
}
