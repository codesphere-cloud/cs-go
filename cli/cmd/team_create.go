// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"fmt"
	"github.com/codesphere-cloud/cs-go/api"
	"github.com/spf13/cobra"
)

type CreateTeamCmd struct {
	cmd           *cobra.Command
	Opts          CreateTeamOpts
	ClientFactory func(GlobalOptions) (Client, error)
}

type CreateTeamOpts struct {
	*GlobalOptions
	Name string
	DcId int
}

func AddCreateTeamCmd(team *cobra.Command, opts *GlobalOptions) {
	t := CreateTeamCmd{
		cmd: &cobra.Command{
			Use:   "create",
			Short: "Create team",
			Long:  `Create a team in Codesphere or an Organization`,
		},
		Opts: CreateTeamOpts{
			GlobalOptions: opts,
		},
		ClientFactory: NewClient,
	}
	t.cmd.RunE = t.RunE
	t.cmd.Flags().StringVarP(&t.Opts.Name, "name", "n", "", "Team name")
	t.cmd.Flags().IntVarP(&t.Opts.DcId, "dc-id", "d", 0, "Data center ID")
	AddCmd(team, t.cmd)
}

func (c *CreateTeamCmd) RunE(_ *cobra.Command, args []string) error {
	client, err := c.ClientFactory(*c.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	orgId, err := c.Opts.GetOrgId()
	if err != nil {
		return errors.Join(err, errors.New("failed to get organization ID"))
	}

	createdTeam, err := c.CreateTeam(client, orgId, c.Opts.Name, c.Opts.DcId)
	if err != nil {
		return err
	}

	fmt.Printf("Team created: %+v in Organization: %+v\n", createdTeam.Id, orgId)
	return nil
}

func (c *CreateTeamCmd) CreateTeam(client Client, orgId string, teamName string, dcId int) (*api.Team, error) {
	if teamName == "" {
		return nil, errors.New("team name cannot be empty")
	}

	createdTeam, err := client.CreateTeam(orgId, teamName, dcId)
	if err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}
	return createdTeam, nil
}
