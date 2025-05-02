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
func TeamIdByName(client Client, name string) (Team, error) {
	teams, err := client.ListTeams()
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
