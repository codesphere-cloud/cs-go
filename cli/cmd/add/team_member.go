// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"errors"
	"fmt"
	"net/mail"

	shared "github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type AddTeamMemberCmd struct {
	cmd           *cobra.Command
	Opts          AddTeamMemberOpts
	ClientFactory func(shared.RootOptions) (shared.Client, error)
}

type AddTeamMemberOpts struct {
	shared.RootOptions
	Email  string
	Role   string
	TeamId int
}

func AddAddTeamMemberCmd(add *cobra.Command, opts shared.RootOptions) {
	t := AddTeamMemberCmd{
		cmd: &cobra.Command{
			Use:     "resource-group-member",
			Aliases: []string{"team-member"},
			Short:   "Add resource group member",
			Long: io.Long(`Add a member to a resource group.

				To add a member to a resource group within an organization or a standalone resource group`),
			Example: io.FormatExampleCommands("add resource-group-member", []io.Example{
				{Cmd: "-t <resourceGroupId> -e user@example.com -r member", Desc: "Add a user to a resource group as a member"},
				{Cmd: "-t <resourceGroupId> -e admin@example.com -r admin", Desc: "Add a user to a resource group as an admin"},
			}),
		},
		Opts: AddTeamMemberOpts{
			RootOptions: opts,
		},
		ClientFactory: func(opts shared.RootOptions) (shared.Client, error) { return opts.NewClient() },
	}
	t.cmd.RunE = t.RunE
	t.cmd.Flags().StringVarP(&t.Opts.Email, "email", "e", "", "Resource group member email")
	_ = t.cmd.MarkFlagRequired("email")
	t.cmd.Flags().StringVarP(&t.Opts.Role, "role", "r", "member", "Resource group member role (member, admin)")
	shared.AddCmd(add, t.cmd)
}

func (c *AddTeamMemberCmd) RunE(_ *cobra.Command, args []string) error {
	client, err := c.ClientFactory(c.Opts.RootOptions)
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

func (c *AddTeamMemberCmd) AddTeamMember(client shared.Client, teamId int, email string, role string) error {
	if email == "" {
		return errors.New("email cannot be empty")
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email address: %w", err)
	}

	var apiRole cs.TeamRole
	switch role {
	case "member":
		apiRole = cs.RoleMember
	case "admin":
		apiRole = cs.RoleAdmin
	default:
		return errors.New("invalid role: must be member or admin")
	}

	err := client.AddTeamMember(teamId, email, int(apiRole))
	if err != nil {
		return fmt.Errorf("failed to add member to team: %w", err)
	}

	return nil
}
