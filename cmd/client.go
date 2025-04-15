package cmd

//go:generate mockery

import (
	"context"
	"fmt"
	"net/url"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
)

type Client interface {
	ListTeams() ([]api.Team, error)
	ListWorkspaces(teamId int) ([]api.Workspace, error)
}

func NewClient(opts GlobalOptions) (Client, error) {
	token, err := cs.GetApiToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get API token: %e", err)
	}
	apiUrl, err := url.Parse(opts.GetApiUrl())
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL '%s': %e", opts.GetApiUrl(), err)
	}
	client := api.NewClient(context.Background(), api.Configuration{
		BaseUrl: apiUrl,
		Token:   token,
	})
	return client, nil
}
