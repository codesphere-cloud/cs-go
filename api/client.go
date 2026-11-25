// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"net/http"
	"net/url"

	"github.com/codesphere-cloud/cs-go/api/errors"
	"github.com/codesphere-cloud/cs-go/api/openapi_client"
)

type Client struct {
	token string
	api   *openapi_client.APIClient
	time  Time
}

type Configuration struct {
	// Url of the codesphere environment
	// Defaults to https://codesphere.com
	BaseUrl *url.URL
	// Codesphere api token
	Token string

	// Verbose output for debugging
	Verbose bool
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

// For use in tests
func NewClientWithCustomDeps(opts Configuration, api *openapi_client.APIClient, time Time) *Client {
	return &Client{
		token: opts.Token,
		api:   api,
		time:  time,
	}
}

func NewClient(opts Configuration) *Client {
	cfg := openapi_client.NewConfiguration()
	cfg.HTTPClient = NewHttpClient()
	cfg.Servers = []openapi_client.ServerConfiguration{{
		URL: opts.BaseUrl.String(),
	}}
	cfg.Debug = opts.Verbose
	return NewClientWithCustomDeps(opts, openapi_client.NewAPIClient(cfg), &RealTime{})
}

// withAuth returns a child of ctx that carries the API token for the generated
// OpenAPI client. Cancellation and deadlines of ctx propagate to the HTTP request.
func (c *Client) withAuth(ctx context.Context) context.Context {
	return context.WithValue(ctx, openapi_client.ContextAccessToken, c.token)
}

// NewHttpClient creates a http client to use for API calls.
// The default http.Client only copies a few "safe" headers
// This custom CheckRedirect ensures all headers are transferred,
// including authorization headers which are necessary for DC redirects
func NewHttpClient() *http.Client {
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Prevent infinite redirects, same as in the default client
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}

			for key, values := range via[0].Header {
				for _, value := range values {
					req.Header.Add(key, value)
				}
			}
			return nil
		},
	}
}

func (c *Client) ListDataCenters(ctx context.Context) ([]DataCenter, error) {
	datacenters, r, err := c.api.MetadataAPI.MetadataGetDatacenters(c.withAuth(ctx)).Execute()
	return datacenters, errors.FormatAPIError(r, err)
}

func (c *Client) ListDomains(ctx context.Context, teamId int) ([]Domain, error) {
	domains, r, err := c.api.DomainsAPI.DomainsListDomains(c.withAuth(ctx), teamId).Execute()
	return domains, errors.FormatAPIError(r, err)
}

func (c *Client) GetDomain(ctx context.Context, teamId int, domainName string) (*Domain, error) {
	domain, r, err := c.api.DomainsAPI.DomainsGetDomain(c.withAuth(ctx), teamId, domainName).Execute()
	return domain, errors.FormatAPIError(r, err)
}

func (c *Client) CreateDomain(ctx context.Context, teamId int, domainName string) (*Domain, error) {
	domain, r, err := c.api.DomainsAPI.DomainsCreateDomain(c.withAuth(ctx), teamId, domainName).Execute()
	return domain, errors.FormatAPIError(r, err)
}

func (c *Client) DeleteDomain(ctx context.Context, teamId int, domainName string) error {
	r, err := c.api.DomainsAPI.DomainsDeleteDomain(c.withAuth(ctx), teamId, domainName).Execute()
	return errors.FormatAPIError(r, err)
}

func (c *Client) UpdateDomain(
	ctx context.Context, teamId int, domainName string, args UpdateDomainArgs,
) (*Domain, error) {
	domain, r, err := c.api.DomainsAPI.
		DomainsUpdateDomain(c.withAuth(ctx), teamId, domainName).
		DomainsUpdateDomainRequest(args).
		Execute()
	return domain, errors.FormatAPIError(r, err)
}

func (c *Client) VerifyDomain(
	ctx context.Context, teamId int, domainName string,
) (*DomainVerificationStatus, error) {
	status, r, err := c.api.DomainsAPI.
		DomainsVerifyDomain(c.withAuth(ctx), teamId, domainName).Execute()
	return status, errors.FormatAPIError(r, err)
}

func (c *Client) UpdateWorkspaceConnections(
	ctx context.Context, teamId int, domainName string, connections PathToWorkspaces,
) (*Domain, error) {
	req := make(map[string][]int)
	for path, workspaces := range connections {
		ids := make([]int, len(workspaces))
		for i, w := range workspaces {
			ids[i] = w.Id
		}
		req[path] = ids
	}
	domain, r, err := c.api.DomainsAPI.
		DomainsUpdateWorkspaceConnections(c.withAuth(ctx), teamId, domainName).
		RequestBody(req).Execute()
	return domain, errors.FormatAPIError(r, err)
}

func (c *Client) ListBaseimages(ctx context.Context) ([]Baseimage, error) {
	baseimages, r, err := c.api.MetadataAPI.MetadataGetWorkspaceBaseImages(c.withAuth(ctx)).Execute()
	return baseimages, errors.FormatAPIError(r, err)
}
