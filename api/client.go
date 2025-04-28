// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"net/url"

	"github.com/codesphere-cloud/cs-go/api/openapi_client"
)

type Client struct {
	ctx context.Context
	api *openapi_client.APIClient
}

type Configuration struct {
	// Url of the codesphere environment
	// Defaults to https://codesphere.com
	BaseUrl *url.URL
	// Codesphere api token
	Token string
}

func (c Configuration) GetApiUrl() *url.URL {
	if c.BaseUrl != nil {
		return c.BaseUrl
	}

	// url.Parse() won't return an error on this static string,
	// hence it's safe to ignore it.
	defaultUrl, _ := url.Parse("https://codesphere.com/api")
	return defaultUrl
}

func NewClientWithCustomApi(ctx context.Context, opts Configuration, api *openapi_client.APIClient) *Client {
	return &Client{
		ctx: context.WithValue(ctx, openapi_client.ContextAccessToken, opts.Token),
		api: api,
	}
}

func NewClient(ctx context.Context, opts Configuration) *Client {
	cfg := openapi_client.NewConfiguration()
	cfg.Servers = []openapi_client.ServerConfiguration{{
		URL: opts.BaseUrl.String(),
	}}
	return NewClientWithCustomApi(ctx, opts, openapi_client.NewAPIClient(cfg))
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

func (c *Client) GetTeam(teamId int) (*Team, error) {
	team, _, err := c.api.TeamsAPI.TeamsGetTeam(c.ctx, float32(teamId)).Execute()
	return ConvertToTeam(team), err
}

func (c *Client) CreateTeam(name string, dc int) (*Team, error) {
	team, _, err := c.api.TeamsAPI.TeamsCreateTeam(c.ctx).
		TeamsCreateTeamRequest(openapi_client.TeamsCreateTeamRequest{
			Name: name,
			Dc:   dc,
		}).
		Execute()
	return ConvertToTeam(team), err
}

func (c *Client) DeleteTeam(teamId int) error {
	_, err := c.api.TeamsAPI.TeamsDeleteTeam(c.ctx, float32(teamId)).Execute()
	return err
}

func (c *Client) ListDomains(teamId int) ([]Domain, error) {
	domains, _, err := c.api.DomainsAPI.DomainsListDomains(c.ctx, float32(teamId)).Execute()
	return domains, err
}

func (c *Client) GetDomain(teamId int, domainName string) (*Domain, error) {
	domain, _, err := c.api.DomainsAPI.DomainsGetDomain(c.ctx, float32(teamId), domainName).Execute()
	return domain, err
}

func (c *Client) CreateDomain(teamId int, domainName string) (*Domain, error) {
	domain, _, err := c.api.DomainsAPI.DomainsCreateDomain(c.ctx, float32(teamId), domainName).Execute()
	return domain, err
}

func (c *Client) DeleteDomain(teamId int, domainName string) error {
	_, err := c.api.DomainsAPI.DomainsDeleteDomain(c.ctx, float32(teamId), domainName).Execute()
	return err
}

func (c *Client) UpdateDomain(
	teamId int, domainName string, args UpdateDomainArgs,
) (*Domain, error) {
	domain, _, err := c.api.DomainsAPI.
		DomainsUpdateDomain(c.ctx, float32(teamId), domainName).
		DomainsGetDomain200ResponseCustomConfig(args).
		Execute()
	return domain, err
}

func (c *Client) VerifyDomain(
	teamId int, domainName string,
) (*DomainVerificationStatus, error) {
	status, _, err := c.api.DomainsAPI.
		DomainsVerifyDomain(c.ctx, float32(teamId), domainName).Execute()
	return status, err
}

func (c *Client) UpdateWorkspaceConnections(
	teamId int, domainName string, connections PathToWorkspaces,
) (*Domain, error) {
	req := make(map[string][]int)
	for path, workspaces := range connections {
		ids := make([]int, len(workspaces))
		for i, w := range workspaces {
			ids[i] = w.Id
		}
		req[path] = ids
	}
	domain, _, err := c.api.DomainsAPI.
		DomainsUpdateWorkspaceConnections(c.ctx, float32(teamId), domainName).
		RequestBody(req).Execute()
	return domain, err
}
