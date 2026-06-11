package mcpserver

import (
	"context"
	"time"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetLandscapeServiceEventsArgs struct {
	TeamId     int    `json:"teamId" jsonschema:"ID of the team"`
	ResourceId string `json:"resourceId" jsonschema:"Resource ID"`
	BeginDate  string `json:"beginDate" jsonschema:"Begin date in RFC3339 format"`
	EndDate    string `json:"endDate" jsonschema:"End date in RFC3339 format"`
	Limit      int    `json:"limit,omitempty" jsonschema:"Limit"`
	Offset     int    `json:"offset,omitempty" jsonschema:"Offset"`
}

type GetUsageSummaryLandscapeArgs struct {
	TeamId    int    `json:"teamId" jsonschema:"ID of the team"`
	BeginDate string `json:"beginDate" jsonschema:"Begin date in RFC3339 format"`
	EndDate   string `json:"endDate" jsonschema:"End date in RFC3339 format"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Limit"`
	Offset    int    `json:"offset,omitempty" jsonschema:"Offset"`
}

func RegisterUsageTools(server *mcp.Server, client *api.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_landscape_service_events",
		Description: "Get landscape service events usage",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetLandscapeServiceEventsArgs) (*mcp.CallToolResult, any, error) {
		begin, _ := time.Parse(time.RFC3339, args.BeginDate)
		end, _ := time.Parse(time.RFC3339, args.EndDate)
		events, err := client.GetLandscapeServiceEvents(args.TeamId, args.ResourceId, begin, end, args.Limit, args.Offset)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, events, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_usage_summary_landscape",
		Description: "Get overall usage summary for a landscape",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetUsageSummaryLandscapeArgs) (*mcp.CallToolResult, any, error) {
		begin, _ := time.Parse(time.RFC3339, args.BeginDate)
		end, _ := time.Parse(time.RFC3339, args.EndDate)
		summary, err := client.GetUsageSummaryLandscape(args.TeamId, begin, end, args.Limit, args.Offset)
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, err
		}
		return nil, summary, nil
	})
}
