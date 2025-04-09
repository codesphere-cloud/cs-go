package api

import (
	"context"
	"net/url"

	"github.com/codesphere-cloud/cs-go/pkg/api/openapi_client"
)

type Client interface {
	ListDataCenters() ([]DataCenter, error)
	ListWorkspacePlans() ([]WorkspacePlan, error)

	ListTeams() ([]Team, error)
}

type Configuration struct {
	BaseUrl *url.URL
	Token   string
}

func NewClient(ctx context.Context, opts Configuration) Client {
	cfg := openapi_client.NewConfiguration()
	if opts.BaseUrl != nil {
		cfg.Servers = []openapi_client.ServerConfiguration{{
			URL: opts.BaseUrl.String(),
		}}
	}

	return &client{
		ctx: context.WithValue(ctx, openapi_client.ContextAccessToken, opts.Token),
		api: openapi_client.NewAPIClient(cfg),
	}
}

type client struct {
	ctx context.Context
	api *openapi_client.APIClient
}

func (c *client) ListDataCenters() ([]DataCenter, error) {
	datacenters, _, err := c.api.MetadataAPI.MetadataGetDatacenters(c.ctx).Execute()
	return datacenters, err
}

func (c *client) ListWorkspacePlans() ([]WorkspacePlan, error) {
	plans, _, err := c.api.MetadataAPI.MetadataGetWorkspacePlans(c.ctx).Execute()
	return plans, err
}

func (c *client) ListTeams() ([]Team, error) {
	teams, _, err := c.api.TeamsAPI.TeamsListTeams(c.ctx).Execute()
	return teams, err
}
