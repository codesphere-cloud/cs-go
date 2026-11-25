// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"fmt"

	cserrors "github.com/codesphere-cloud/cs-go/api/errors"
	"github.com/codesphere-cloud/cs-go/api/openapi_client"
)

// Fetches the team for a given team name.
// If the name is ambigous an error is thrown.
//
// Returns [NotFound] if no team with the given Id could be found
// Returns [Duplicated] if no team with the given Id could be found
func (client *Client) TeamIdByName(ctx context.Context, name string, orgId string) (Team, error) {
	teams, err := client.ListTeams(ctx, orgId)
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

func (c *Client) ListTeams(ctx context.Context, orgId string) ([]Team, error) {
	if orgId != "" {
		teams, r, err := c.api.OrganizationsAPI.OrganizationsListOrgTeams(c.withAuth(ctx), orgId).Execute()
		if err != nil {
			return nil, cserrors.FormatAPIError(r, err)
		}

		res := make([]Team, len(teams))
		for i, t := range teams {
			res[i] = *ConvertOrgTeamToTeam(t, orgId)
		}
		return res, nil
	}

	apiTeams, r, err := c.api.TeamsAPI.TeamsListTeams(c.withAuth(ctx)).Execute()
	if err != nil {
		return nil, cserrors.FormatAPIError(r, err)
	}
	res := make([]Team, len(apiTeams))
	for i, t := range apiTeams {
		res[i] = *ConvertFromListTeams(t)
	}
	return res, nil
}

func (c *Client) GetTeam(ctx context.Context, teamId int) (*Team, error) {
	team, r, err := c.api.TeamsAPI.TeamsGetTeam(c.withAuth(ctx), teamId).Execute()
	if err != nil {
		return nil, cserrors.FormatAPIError(r, err)
	}
	if team == nil {
		return nil, cserrors.NotFound("team not found")
	}
	return ConvertToTeam(team), nil
}

func (c *Client) CreateTeam(ctx context.Context, orgId string, name string, dc int) (*Team, error) {
	if orgId == "" {
		return c.createTeam(ctx, name, dc)
	}
	return c.createOrgTeam(ctx, orgId, name, dc)

}

func (c *Client) createTeam(ctx context.Context, name string, dc int) (*Team, error) {
	team, r, err := c.api.TeamsAPI.TeamsCreateTeam(c.withAuth(ctx)).
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

func (c *Client) createOrgTeam(ctx context.Context, orgId string, name string, dc int) (*Team, error) {
	team, r, err := c.api.TeamsAPI.TeamsCreateTeam(c.withAuth(ctx)).
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

func (c *Client) DeleteTeam(ctx context.Context, teamId int) error {
	r, err := c.api.TeamsAPI.TeamsDeleteTeam(c.withAuth(ctx), teamId).Execute()
	return cserrors.FormatAPIError(r, err)
}

func (c *Client) AddTeamMember(ctx context.Context, teamId int, email string, role int) error {
	r, err := c.api.TeamsAPI.TeamsInviteMember(c.withAuth(ctx), teamId).
		TeamsInviteMemberRequest(openapi_client.TeamsInviteMemberRequest{
			UserEmail: email,
			Role:      role,
		}).Execute()
	if err != nil {
		return cserrors.FormatAPIError(r, err)
	}
	return nil
}

func (c *Client) RemoveTeamMember(ctx context.Context, teamId int, userId int) error {
	r, err := c.api.TeamsAPI.TeamsRemoveMember(c.withAuth(ctx), teamId, userId).Execute()
	if err != nil {
		return cserrors.FormatAPIError(r, err)
	}
	return nil
}

func (c *Client) ListTeamMembers(ctx context.Context, teamId int) ([]TeamMember, error) {
	members, r, err := c.api.TeamsAPI.TeamsListMembers(c.withAuth(ctx), teamId).Execute()
	if err != nil {
		return nil, cserrors.FormatAPIError(r, err)
	}
	res := make([]TeamMember, len(members))
	for i, m := range members {
		res[i] = ConvertToTeamMember(m)
	}
	return res, nil
}
