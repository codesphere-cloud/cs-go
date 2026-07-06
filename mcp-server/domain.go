// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListDomainsArgs struct {
	TeamId int `json:"teamId" jsonschema:"ID of the team"`
}

type GetDomainArgs struct {
	TeamId     int    `json:"teamId" jsonschema:"ID of the team"`
	DomainName string `json:"domainName" jsonschema:"The domain name"`
}

type CreateDomainArgs struct {
	TeamId     int    `json:"teamId" jsonschema:"ID of the team"`
	DomainName string `json:"domainName" jsonschema:"The domain name to create"`
}

type DeleteDomainArgs struct {
	TeamId     int    `json:"teamId" jsonschema:"ID of the team"`
	DomainName string `json:"domainName" jsonschema:"The domain name to delete"`
}

type UpdateDomainArgs struct {
	TeamId     int                  `json:"teamId" jsonschema:"ID of the team"`
	DomainName string               `json:"domainName" jsonschema:"The domain name to update"`
	Args       api.UpdateDomainArgs `json:"args" jsonschema:"The update arguments array"`
}

type VerifyDomainArgs struct {
	TeamId     int    `json:"teamId" jsonschema:"ID of the team"`
	DomainName string `json:"domainName" jsonschema:"The domain name to verify"`
}

func RegisterDomainTools(server *mcp.Server, client *api.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_domains",
		Description: "List all domains for a team",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ListDomainsArgs) (*mcp.CallToolResult, any, error) {
		domains, err := client.ListDomains(args.TeamId)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, map[string]any{"items": domains}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_domain",
		Description: "Get a specific domain by name",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetDomainArgs) (*mcp.CallToolResult, any, error) {
		domain, err := client.GetDomain(args.TeamId, args.DomainName)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, domain, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_domain",
		Description: "Create a domain for a team",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args CreateDomainArgs) (*mcp.CallToolResult, any, error) {
		domain, err := client.CreateDomain(args.TeamId, args.DomainName)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, domain, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_domain",
		Description: "Delete a domain",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args DeleteDomainArgs) (*mcp.CallToolResult, any, error) {
		err := client.DeleteDomain(args.TeamId, args.DomainName)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, map[string]string{"status": "deleted"}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_domain",
		Description: "Update a domain",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args UpdateDomainArgs) (*mcp.CallToolResult, any, error) {
		domain, err := client.UpdateDomain(args.TeamId, args.DomainName, args.Args)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, domain, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "verify_domain",
		Description: "Trigger verification for a domain",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args VerifyDomainArgs) (*mcp.CallToolResult, any, error) {
		status, err := client.VerifyDomain(args.TeamId, args.DomainName)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, status, nil
	})
}
