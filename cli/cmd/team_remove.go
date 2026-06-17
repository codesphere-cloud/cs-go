// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"fmt"

	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type RemoveTeamCmd struct {
	cmd           *cobra.Command
	Opts          RemoveTeamOpts
	ClientFactory func(GlobalOptions) (Client, error)
}

type RemoveTeamOpts struct {
	*GlobalOptions
}

func AddRemoveTeamCmd(team *cobra.Command, opts *GlobalOptions) {
	t := RemoveTeamCmd{
		cmd: &cobra.Command{
			Use:   "remove",
			Short: "Remove team",
			Long:  `Remove a team from Codesphere or an Organization`,
			Example: io.FormatExampleCommands("team remove", []io.Example{
				{Cmd: "-t <teamId>", Desc: "Remove a team that does not belong to an Organization"},
				{Cmd: "-O <orgId> -t <teamId>", Desc: "Remove a team that does belong to an Organization"},
			}),
		},
		Opts: RemoveTeamOpts{
			GlobalOptions: opts,
		},
		ClientFactory: NewClient,
	}
	t.cmd.RunE = t.RunE
	AddCmd(team, t.cmd)
}

func (c *RemoveTeamCmd) RunE(_ *cobra.Command, args []string) error {
	client, err := c.ClientFactory(*c.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codespehre client: %w", err)
	}

	orgId, err := c.Opts.GetOrgId()
	if err != nil {
		return err
	}

	teamId, err := c.Opts.GetTeamId()
	if err != nil {
		return errors.New("team ID not set, use -t or CS_TEAM_ID to set it")
	}

	err = client.DeleteTeam(orgId, teamId)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}
	return nil
}
