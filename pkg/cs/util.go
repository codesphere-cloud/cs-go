// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/codesphere-cloud/cs-go/api"
	cserrors "github.com/codesphere-cloud/cs-go/pkg/errors"
)

type Step struct {
	State string
}

type ReplicaStatus struct {
	State   string `json:"state"`
	Steps   []Step `json:"steps"`
	Replica string `json:"replica"`
	Server  string `json:"server"`
}

func GetApiUrl() string {
	url := os.Getenv("CS_API")
	if url != "" {
		return url
	}
	return "https://codesphere.com/api"
}

func GetPipelineStatus(ws int, stage string) (res []ReplicaStatus, err error) {

	status, err := Get(fmt.Sprintf("workspaces/%d/pipeline/%s", ws, stage))
	if err != nil {
		err = fmt.Errorf("failed to get pipeline status: %w", err)
		return
	}

	err = json.Unmarshal(status, &res)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal pipeline status: %w", err)
		return
	}
	return
}

func Get(path string) (body []byte, err error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", GetApiUrl(), strings.TrimPrefix(path, "/")), http.NoBody)
	if err != nil {
		err = fmt.Errorf("failed to create request: %w", err)
		return
	}
	err = SetAuthoriziationHeader(req)
	if err != nil {
		err = fmt.Errorf("failed to set header: %w", err)
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf("GET failed: %w", err)
		return
	}
	defer func() { _ = res.Body.Close() }()
	body, err = io.ReadAll(res.Body)
	return
}

func GetApiToken() (string, error) {
	apiToken := os.Getenv("CS_TOKEN")
	if apiToken == "" {
		return "", errors.New("CS_TOKEN env var required, but not set")
	}
	return apiToken, nil
}

func SetAuthoriziationHeader(req *http.Request) error {
	token, err := GetApiToken()
	if err != nil {
		return fmt.Errorf("failed to get API token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func GetRoleName(role int) string {
	if role == 1 {
		return "Member"
	}
	return "Admin"
}

// Fetches the team for a given team name.
// If the name is ambigous an error is thrown.
//
// Returns [NotFound] if no plan with the given Id could be found
// Returns [Duplicated] if no plan with the given Id could be found
func TeamIdByName(
	client *api.Client,
	name string,
) (api.Team, error) {
	teams, err := client.ListTeams()
	if err != nil {
		return api.Team{}, err
	}

	matchingTeams := []api.Team{}
	for _, t := range teams {
		if t.Name == name {
			matchingTeams = append(matchingTeams, t)
		}
	}

	if len(matchingTeams) == 0 {
		return api.Team{}, cserrors.NewNotFound(fmt.Sprintf("No team with name %s found", name))
	}

	if len(matchingTeams) > 1 {
		return api.Team{}, cserrors.NewDuplicated(fmt.Sprintf("Multiple teams (%v) with the name %s found.", matchingTeams, name))
	}

	return matchingTeams[0], nil
}

// Fetches the workspace plan for a given name.
//
// Returns [NotFound] if no plan with the given Id could be found
func PlanByName(
	client *api.Client,
	name string,
) (api.WorkspacePlan, error) {
	plans, err := client.ListWorkspacePlans()
	if err != nil {
		return api.WorkspacePlan{}, err
	}

	for _, p := range plans {
		if p.Title == name {
			return p, nil
		}
	}
	return api.WorkspacePlan{}, cserrors.NewNotFound(fmt.Sprintf("No team with name %s found", name))
}
