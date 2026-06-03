// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

type AddTeamMemberCmd struct {
	cmd           *cobra.Command
	Opts          AddTeamMemberOpts
	ClientFactory func(GlobalOptions) (Client, error)
}

type AddTeamMemberOpts struct {
	*GlobalOptions
	Email  string
	Role   int
	TeamId int
}

func AddAddTeamMemberCmd(team *cobra.Command, opts *GlobalOptions) {
	t := AddTeamMemberCmd{
		cmd: &cobra.Command{
			Use:   "add",
			Short: "Add team member",
			Long:  `Add team member to a team`,
		},
		Opts: AddTeamMemberOpts{
			GlobalOptions: opts,
		},
		ClientFactory: NewClient,
	}
	t.cmd.RunE = t.RunE
	t.cmd.Flags().StringVarP(&t.Opts.Email, "email", "e", "", "Team member email")
	t.cmd.Flags().IntVarP(&t.Opts.Role, "role", "r", 0, "Team member role 0=admin, 1=member")
	AddCmd(team, t.cmd)
}

func (c *AddTeamMemberCmd) RunE(_ *cobra.Command, args []string) error {
	client, err := c.ClientFactory(*c.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codespehre client: %w", err)
	}

	teamId, err := c.Opts.GetTeamId()
	if err != nil {
		return err
	}

	err = c.AddTeamMember(client, teamId, c.Opts.Email, c.Opts.Role)
	return err

}

func (c *AddTeamMemberCmd) AddTeamMember(client Client, teamId int, email string, role int) error {
	if email == "" {
		return errors.New("email cannot be empty")
	}

	fmt.Printf("add member: %s to team %d with role: %d", email, teamId, role)

	err := client.AddTeamMember(teamId, email, role)
	if err != nil {
		return fmt.Errorf("failed to add member to team: %w", err)
	}

	return nil
}
