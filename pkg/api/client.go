package api

import (
	"context"
	"net/url"

	"github.com/codesphere-cloud/cs-go/pkg/api/openapi_client"
)

type Client struct {
	ctx context.Context
	api *openapi_client.APIClient
}

type Configuration struct {
	BaseUrl *url.URL
	Token   string
}

func (c Configuration) GetApiUrl() *url.URL {
	if c.BaseUrl != nil {
		return c.BaseUrl
	}
	defaultUrl, _ := url.Parse("https://codesphere.com/api")
	return defaultUrl
}

func NewClient(ctx context.Context, opts Configuration) *Client {
	cfg := openapi_client.NewConfiguration()
	cfg.Servers = []openapi_client.ServerConfiguration{{
		URL: opts.GetApiUrl().String(),
	}}

	return &Client{
		ctx: context.WithValue(ctx, openapi_client.ContextAccessToken, opts.Token),
		api: openapi_client.NewAPIClient(cfg),
	}
}

func (c *Client) ListDataCenters() ([]DataCenter, error) {
	datacenters, _, err := c.api.MetadataAPI.MetadataGetDatacenters(c.ctx).Execute()
	return datacenters, err
}

func (c *Client) ListWorkspacePlans() ([]WorkspacePlan, error) {
	plans, _, err := c.api.MetadataAPI.MetadataGetWorkspacePlans(c.ctx).Execute()
	return plans, err
}

func (c *Client) ListTeams() ([]Team, error) {
	teams, _, err := c.api.TeamsAPI.TeamsListTeams(c.ctx).Execute()
	return teams, err
}

func (c *Client) ListWorkspaces(teamId int) ([]Workspace, error) {
	workspaces, _, err := c.api.WorkspacesAPI.WorkspacesListWorkspaces(c.ctx, float32(teamId)).Execute()
	return workspaces, err
}
