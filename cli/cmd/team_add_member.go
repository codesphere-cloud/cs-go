// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"fmt"
	"net/mail"

	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/codesphere-cloud/cs-go/pkg/io"
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
			Long: io.Long(`Add team member to a team.
			
				To add a member to a team within an organization, the CS_ORG_ID environment variable or the -O/--org flag must be set.`),
			Example: io.FormatExampleCommands("team member add", []io.Example{
				{Cmd: "-t <teamId> -e user@example.com -r 1", Desc: "Add a user to a team as a member"},
				{Cmd: "-t <teamId> -e admin@example.com -r 0", Desc: "Add a user to a team as an admin"},
				{Cmd: "-O <org-id> -t  <teamId> -e user@example.com -r 1", Desc: "Add a user to a team within an organization"},
			}),
		},
		Opts: AddTeamMemberOpts{
			GlobalOptions: opts,
		},
		ClientFactory: NewClient,
	}
	t.cmd.RunE = t.RunE
	t.cmd.Flags().StringVarP(&t.Opts.Email, "email", "e", "", "Team member email")
	_ = t.cmd.MarkFlagRequired("email")
	t.cmd.Flags().IntVarP(&t.Opts.Role, "role", "r", int(cs.RoleAdmin), "Team member role 0=admin, 1=member")
	AddCmd(team, t.cmd)
}

func (c *AddTeamMemberCmd) RunE(_ *cobra.Command, args []string) error {
	client, err := c.ClientFactory(*c.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
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

	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email address: %w", err)
	}

	if !cs.TeamRole(role).IsValid() {
		return errors.New("invalid role: must be 0 for admin or 1 for member")
	}

	err := client.AddTeamMember(teamId, email, role)
	if err != nil {
		return fmt.Errorf("failed to add member to team: %w", err)
	}

	return nil
}
