// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/codesphere-cloud/cs-go/api/errors"
	"github.com/codesphere-cloud/cs-go/api/openapi_client"
)

type Client interface {
	ListTeams() ([]Team, error)
	ListWorkspaces(teamId int) ([]Workspace, error)
	ListBaseimages() ([]Baseimage, error)
	GetWorkspace(workspaceId int) (Workspace, error)
	WorkspaceStatus(workspaceId int) (*WorkspaceStatus, error)
	WaitForWorkspaceRunning(workspace *Workspace, timeout time.Duration) error
	ScaleWorkspace(wsId int, replicas int) error
	ScaleLandscapeServices(wsId int, services map[string]int) error
	SetEnvVarOnWorkspace(workspaceId int, vars map[string]string) error
	ExecCommand(workspaceId int, command string, workdir string, env map[string]string) (string, string, error)
	ListWorkspacePlans() ([]WorkspacePlan, error)
	DeployWorkspace(args DeployWorkspaceArgs) (*Workspace, error)
	DeleteWorkspace(wsId int) error
	StartPipelineStage(wsId int, profile string, stage string) error
	GetPipelineState(wsId int, stage string) ([]PipelineStatus, error)
	GitPull(wsId int, remote string, branch string) error
	DeployLandscape(wsId int, profile string) error
	WakeUpWorkspace(wsId int, token string, profile string, timeout time.Duration) error
}

type RealClient struct {
	ctx  context.Context
	api  *openapi_client.APIClient
	time Time
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
func NewClientWithCustomDeps(ctx context.Context, opts Configuration, api *openapi_client.APIClient, time Time) *RealClient {
	return &RealClient{
		ctx:  context.WithValue(ctx, openapi_client.ContextAccessToken, opts.Token),
		api:  api,
		time: time,
	}
}

func NewClient(ctx context.Context, opts Configuration) *RealClient {
	cfg := openapi_client.NewConfiguration()
	cfg.HTTPClient = NewHttpClient()
	cfg.Servers = []openapi_client.ServerConfiguration{{
		URL: opts.BaseUrl.String(),
	}}
	cfg.Debug = opts.Verbose
	return NewClientWithCustomDeps(ctx, opts, openapi_client.NewAPIClient(cfg), &RealTime{})
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

func (c *RealClient) ListDataCenters() ([]DataCenter, error) {
	datacenters, r, err := c.api.MetadataAPI.MetadataGetDatacenters(c.ctx).Execute()
	return datacenters, errors.FormatAPIError(r, err)
}

func (c *RealClient) ListDomains(teamId int) ([]Domain, error) {
	domains, r, err := c.api.DomainsAPI.DomainsListDomains(c.ctx, float32(teamId)).Execute()
	return domains, errors.FormatAPIError(r, err)
}

func (c *RealClient) GetDomain(teamId int, domainName string) (*Domain, error) {
	domain, r, err := c.api.DomainsAPI.DomainsGetDomain(c.ctx, float32(teamId), domainName).Execute()
	return domain, errors.FormatAPIError(r, err)
}

func (c *RealClient) CreateDomain(teamId int, domainName string) (*Domain, error) {
	domain, r, err := c.api.DomainsAPI.DomainsCreateDomain(c.ctx, float32(teamId), domainName).Execute()
	return domain, errors.FormatAPIError(r, err)
}

func (c *RealClient) DeleteDomain(teamId int, domainName string) error {
	r, err := c.api.DomainsAPI.DomainsDeleteDomain(c.ctx, float32(teamId), domainName).Execute()
	return errors.FormatAPIError(r, err)
}

func (c *RealClient) UpdateDomain(
	teamId int, domainName string, args UpdateDomainArgs,
) (*Domain, error) {
	domain, r, err := c.api.DomainsAPI.
		DomainsUpdateDomain(c.ctx, float32(teamId), domainName).
		DomainsUpdateDomainRequest(args).
		Execute()
	return domain, errors.FormatAPIError(r, err)
}

func (c *RealClient) VerifyDomain(
	teamId int, domainName string,
) (*DomainVerificationStatus, error) {
	status, r, err := c.api.DomainsAPI.
		DomainsVerifyDomain(c.ctx, float32(teamId), domainName).Execute()
	return status, errors.FormatAPIError(r, err)
}

func (c *RealClient) UpdateWorkspaceConnections(
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
	domain, r, err := c.api.DomainsAPI.
		DomainsUpdateWorkspaceConnections(c.ctx, float32(teamId), domainName).
		RequestBody(req).Execute()
	return domain, errors.FormatAPIError(r, err)
}

func (c *RealClient) ListBaseimages() ([]Baseimage, error) {
	baseimages, r, err := c.api.MetadataAPI.MetadataGetWorkspaceBaseImages(c.ctx).Execute()
	return baseimages, errors.FormatAPIError(r, err)
}
