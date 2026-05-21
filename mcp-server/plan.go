package main

import (
	"context"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListWorkspacePlansArgs struct{}

func RegisterPlanTools(server *mcp.Server, client *api.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_workspace_plans",
		Description: "List all standard Codesphere workspace plans",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ListWorkspacePlansArgs) (*mcp.CallToolResult, any, error) {
		plans, err := client.ListWorkspacePlans()
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, plans, nil
	})
}
