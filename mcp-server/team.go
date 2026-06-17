// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListTeamsArgs struct {
	OrgId string `json:"orgId,omitempty" jsonschema:"Optional organization ID to list teams for"`
}

type GetTeamArgs struct {
	TeamId int `json:"teamId" jsonschema:"ID of the team to get"`
}

type CreateTeamArgs struct {
	OrgId string `json:"orgId,omitempty" jsonschema:"Optional organization ID to create the team in"`
	Name  string `json:"name" jsonschema:"Name of the new team"`
	Dc    int    `json:"dc" jsonschema:"Default datacenter ID"`
}

type DeleteTeamArgs struct {
	OrgId  string `json:"orgId,omitempty" jsonschema:"Optional organization ID the team belongs to"`
	TeamId int    `json:"teamId" jsonschema:"ID of the team to delete"`
}

func RegisterTeamTools(server *mcp.Server, client *api.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_teams",
		Description: "List all teams the authenticated user belongs to",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ListTeamsArgs) (*mcp.CallToolResult, any, error) {
		teams, err := client.ListTeams(args.OrgId)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, teams, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_team",
		Description: "Get details of a single team",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetTeamArgs) (*mcp.CallToolResult, any, error) {
		team, err := client.GetTeam(args.TeamId)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, team, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_team",
		Description: "Create a new team",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args CreateTeamArgs) (*mcp.CallToolResult, any, error) {
		team, err := client.CreateTeam(args.OrgId, args.Name, args.Dc)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, team, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_team",
		Description: "Delete a team by ID",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args DeleteTeamArgs) (*mcp.CallToolResult, any, error) {
		err := client.DeleteTeam(args.OrgId, args.TeamId)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, map[string]string{"status": "deleted"}, nil
	})
}
