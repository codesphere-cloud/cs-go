// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"github.com/spf13/cobra"
)

type RemoveTeamCmd struct {
	cmd  *cobra.Command
	Opts RemoveTeamOpts
}

type RemoveTeamOpts struct {
	*GlobalOptions
	name string
}

func AddRemoveTeamCmd(team *cobra.Command, opts *GlobalOptions) {
	t := RemoveTeamCmd{
		cmd: &cobra.Command{
			Use:   "remove",
			Short: "Remove team",
			Long:  `Remove a team from Codesphere or an Organization`,
		},
		Opts: RemoveTeamOpts{
			GlobalOptions: opts,
		},
	}
	t.cmd.RunE = t.RunE
	t.cmd.Flags().StringVarP(&t.Opts.name, "name", "n", "", "Team name")

	AddCmd(team, t.cmd)
}

func (c *RemoveTeamCmd) RunE(_ *cobra.Command, args []string) error {
	// TODO: Implement team removal logic
	client, err := NewClient(*c.Opts.GlobalOptions)
	if err != nil {
		return err
	}

	orgId, err := c.Opts.GetOrgId()
	if err != nil {
		return err
	}

	teamId, err := c.Opts.GetTeamId()
	if err != nil {
		return errors.New("team ID not set, use -T or CS_TEAM_ID to set it")
	}

	//

	client.DeleteTeam(orgId, teamId)
	return nil

}
