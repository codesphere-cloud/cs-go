package mcpserver

import (
	"context"
	"log"
	"net/url"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func Run() error {
	opts := cs.NewEnv()

	token, err := opts.GetApiToken()
	if err != nil {
		log.Fatalf("failed to get API token: %v", err)
	}
	apiUrl, err := url.Parse(opts.GetApiUrl())
	if err != nil {
		log.Fatalf("failed to parse URL '%s': %v", opts.GetApiUrl(), err)
	}

	client := api.NewClient(context.Background(), api.Configuration{
		BaseUrl: apiUrl,
		Token:   token,
		Verbose: false,
	})

	server := mcp.NewServer(&mcp.Implementation{Name: "codesphere-mcp", Version: "v1.0.0"}, nil)

	RegisterMetadataTools(server, client)
	RegisterPlanTools(server, client)
	RegisterTeamTools(server, client)
	RegisterDomainTools(server, client)
	RegisterWorkspaceTools(server, client)
	RegisterUsageTools(server, client)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		return err
	}
	return nil
}

// adding usage tools to main
