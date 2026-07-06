// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	cserrors "github.com/codesphere-cloud/cs-go/api/errors"
	"github.com/codesphere-cloud/cs-go/api/openapi_client"
)

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

	apiTeams, r, err := c.api.TeamsAPI.TeamsListTeams(c.ctx).Execute()
	if err != nil {
		return nil, cserrors.FormatAPIError(r, err)
	}
	res := make([]Team, len(apiTeams))
	for i, t := range apiTeams {
		res[i] = *ConvertFromListTeams(t)
	}
	return res, nil
}

func (c *Client) GetTeam(teamId int) (*Team, error) {
	team, r, err := c.api.TeamsAPI.TeamsGetTeam(c.ctx, teamId).Execute()
	if err != nil {
		return nil, cserrors.FormatAPIError(r, err)
	}
	if team == nil {
		return nil, cserrors.NotFound("team not found")
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

func (c *Client) DeleteTeam(teamId int) error {
	r, err := c.api.TeamsAPI.TeamsDeleteTeam(c.ctx, teamId).Execute()
	return cserrors.FormatAPIError(r, err)
}

func (c *Client) AddTeamMember(teamId int, email string, role int) error {
	r, err := c.api.TeamsAPI.TeamsInviteMember(c.ctx, teamId).
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
	r, err := c.api.TeamsAPI.TeamsRemoveMember(c.ctx, teamId, userId).Execute()
	if err != nil {
		return cserrors.FormatAPIError(r, err)
	}
	return nil
}

func (c *Client) ListTeamMembers(teamId int) ([]TeamMember, error) {
	members, r, err := c.api.TeamsAPI.TeamsListMembers(c.ctx, teamId).Execute()
	if err != nil {
		return nil, cserrors.FormatAPIError(r, err)
	}
	res := make([]TeamMember, len(members))
	for i, m := range members {
		res[i] = ConvertToTeamMember(m)
	}
	return res, nil
}
