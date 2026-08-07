// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package delete

import (
	"errors"
	"fmt"

	shared "github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type DeleteTeamMemberCmd struct {
	cmd           *cobra.Command
	Opts          DeleteTeamMemberOpts
	ClientFactory func(shared.RootOptions) (Client, error)
}

type DeleteTeamMemberOpts struct {
	shared.RootOptions
	UserId int
}

func AddDeleteTeamMemberCmd(delete *cobra.Command, opts shared.RootOptions) {
	res := DeleteTeamMemberCmd{
		cmd: &cobra.Command{
			Use:   "team-member",
			Short: "Delete team member",
			Long: io.Long(`Delete a member from a team.

				To delete a member from a team within an organization, the CS_ORG_ID environment variable or the -O/--org flag must be set.`),
			Example: io.FormatExampleCommands("delete team-member", []io.Example{
				{Cmd: "-t <teamId> -u <userId>", Desc: "Delete a user from a team"},
				{Cmd: "-O <org-id> -t <teamId> -u <userId>", Desc: "Delete a user from a team within an organization"},
			}),
		},
		Opts: DeleteTeamMemberOpts{
			RootOptions: opts,
		},
		ClientFactory: func(opts shared.RootOptions) (Client, error) { return opts.NewClient() },
	}
	res.cmd.Flags().IntVarP(&res.Opts.UserId, "user", "u", 0, "Team member user ID")
	_ = res.cmd.MarkFlagRequired("user")
	res.cmd.RunE = res.RunE
	shared.AddCmd(delete, res.cmd)
}

func (c *DeleteTeamMemberCmd) RunE(_ *cobra.Command, args []string) error {
	client, err := c.ClientFactory(c.Opts.RootOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	teamId, err := c.Opts.GetTeamId()
	if err != nil {
		return err
	}

	return c.DeleteTeamMember(client, teamId, c.Opts.UserId)
}

func (c *DeleteTeamMemberCmd) DeleteTeamMember(client Client, teamId int, userId int) error {
	if userId <= 0 {
		return errors.New("user ID has to be set")
	}

	err := client.RemoveTeamMember(teamId, userId)
	if err != nil {
		return fmt.Errorf("failed to remove member from team: %w", err)
	}

	return nil
}
