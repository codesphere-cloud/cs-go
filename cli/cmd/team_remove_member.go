// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"fmt"

	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type RemoveTeamMemberCmd struct {
	cmd           *cobra.Command
	Opts          RemoveTeamMemberOpts
	ClientFactory func(GlobalOptions) (Client, error)
}

type RemoveTeamMemberOpts struct {
	*GlobalOptions
	UserId int
}

func AddRemoveTeamMemberCmd(team *cobra.Command, opts *GlobalOptions) {
	res := RemoveTeamMemberCmd{
		cmd: &cobra.Command{
			Use:   "remove",
			Short: "Remove team member",
			Long: io.Long(`Remove team member from a team.

				To add a member to a team within an organization, the CS_ORG_ID environment variable or the -O/--org flag must be set.`),
			Example: io.FormatExampleCommands("team member remove", []io.Example{
				{Cmd: "-t <teamId> -u <userId>", Desc: "Remove a user from a team"},
				{Cmd: "-O <org-id> -t <teamId> -u <userId>", Desc: "Remove a user from a team within an organization"},
			}),
		},
		Opts: RemoveTeamMemberOpts{
			GlobalOptions: opts,
		},
		ClientFactory: NewClient,
	}
	res.cmd.Flags().IntVarP(&res.Opts.UserId, "user", "u", 0, "Team member user ID")
	res.cmd.MarkFlagRequired("user")
	res.cmd.RunE = res.RunE
	AddCmd(team, res.cmd)
}

func (c *RemoveTeamMemberCmd) RunE(_ *cobra.Command, args []string) error {
	client, err := c.ClientFactory(*c.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codespehre client: %w", err)
	}

	teamId, err := c.Opts.GetTeamId()
	if err != nil {
		return err
	}

	return c.RemoveTeamMember(client, teamId, c.Opts.UserId)
}

func (c *RemoveTeamMemberCmd) RemoveTeamMember(client Client, teamId int, userId int) error {
	if userId <= 0 {
		return errors.New("user ID has to be set")
	}

	err := client.RemoveTeamMember(teamId, userId)
	if err != nil {
		return fmt.Errorf("failed to remove member from team: %w", err)
	}

	return nil
}
