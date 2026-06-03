// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"fmt"

	cserrors "github.com/codesphere-cloud/cs-go/api/errors"
	"github.com/codesphere-cloud/cs-go/api/openapi_client"
)

// Fetches the team for a given team name.
// If the name is ambigous an error is thrown.
//
// Returns [NotFound] if no plan with the given Id could be found
// Returns [Duplicated] if no plan with the given Id could be found
func (client *Client) TeamIdByName(name string) (Team, error) {
	teams, err := client.ListTeams("")
	if err != nil {
		return Team{}, err
	}

	matchingTeams := []Team{}
	for _, t := range teams {
		if t.Name == name {
			matchingTeams = append(matchingTeams, t)
		}
	}

	if len(matchingTeams) == 0 {
		return Team{}, cserrors.NotFound(fmt.Sprintf("no team with name %s found", name))
	}

	if len(matchingTeams) > 1 {
		return Team{}, cserrors.Duplicated(fmt.Sprintf("multiple teams (%v) with the name %s found.", matchingTeams, name))
	}

	return matchingTeams[0], nil
}

func (c *Client) ListTeams(orgId string) ([]Team, error) {
	if orgId != "" {
		teams, r, err := c.api.OrganizationsAPI.OrganizationsListOrgTeams(c.ctx, orgId).Execute()
		if err != nil {
			return nil, cserrors.FormatAPIError(r, err)
		}

		res := make([]Team, len(teams))
		for i, t := range teams {
			res[i] = *ConvertOrgTeamToTeam(t, orgId)
		}
		return res, nil
	}

	teams, r, err := c.api.TeamsAPI.TeamsListTeams(c.ctx).Execute()
	return teams, cserrors.FormatAPIError(r, err)
}

func (c *Client) GetTeam(teamId int) (*Team, error) {
	team, r, err := c.api.TeamsAPI.TeamsGetTeam(c.ctx, float32(teamId)).Execute()
	if err != nil {
		return nil, cserrors.FormatAPIError(r, err)
	}
	return ConvertToTeam(team), nil
}

func (c *Client) CreateTeam(orgId string, name string, dc int) (*Team, error) {
	if orgId == "" {
		return c.createTeam(name, dc)
	}
	return c.createOrgTeam(orgId, name, dc)

}

func (c *Client) createTeam(name string, dc int) (*Team, error) {
	team, r, err := c.api.TeamsAPI.TeamsCreateTeam(c.ctx).
		TeamsCreateTeamRequest(openapi_client.TeamsCreateTeamRequest{
			Name: name,
			Dc:   dc,
		}).
		Execute()
	if err != nil {
		return nil, cserrors.FormatAPIError(r, err)
	}
	return ConvertToTeam(team), nil
}

func (c *Client) createOrgTeam(orgId string, name string, dc int) (*Team, error) {
	team, r, err := c.api.TeamsAPI.TeamsCreateTeam(c.ctx).
		TeamsCreateTeamRequest(openapi_client.TeamsCreateTeamRequest{
			Name:           name,
			Dc:             dc,
			OrganizationId: &orgId,
		}).
		Execute()
	if err != nil {
		return nil, cserrors.FormatAPIError(r, err)
	}
	return ConvertToTeam(team), nil
}

func (c *Client) DeleteTeam(orgId string, teamId int) error {
	r, err := c.api.TeamsAPI.TeamsDeleteTeam(c.ctx, float32(teamId)).Execute()
	return cserrors.FormatAPIError(r, err)
}

func (c *Client) AddTeamMember(teamId int, email string, role int) error {
	r, err := c.api.TeamsAPI.TeamsInviteMember(c.ctx, float32(teamId)).
		TeamsInviteMemberRequest(openapi_client.TeamsInviteMemberRequest{
			UserEmail: email,
			Role:      role,
		}).Execute()
	if err != nil {
		return cserrors.FormatAPIError(r, err)
	}
	return nil
}

func (c *Client) RemoveTeamMember(teamId int, userId int) error {
	r, err := c.api.TeamsAPI.TeamsRemoveMember(c.ctx, float32(teamId), float32(userId)).Execute()
	if err != nil {
		return cserrors.FormatAPIError(r, err)
	}
	return nil
}
