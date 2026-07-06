// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListDataCentersArgs struct{}
type ListBaseimagesArgs struct{}

func RegisterMetadataTools(server *mcp.Server, client *api.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_data_centers",
		Description: "List all available data centers in Codesphere",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ListDataCentersArgs) (*mcp.CallToolResult, any, error) {
		dcs, err := client.ListDataCenters()
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, map[string]any{"items": dcs}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_base_images",
		Description: "List all base images available for workspaces",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ListBaseimagesArgs) (*mcp.CallToolResult, any, error) {
		images, err := client.ListBaseimages()
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, map[string]any{"items": images}, nil
	})
}
