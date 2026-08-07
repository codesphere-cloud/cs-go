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

type DeleteTeamCmd struct {
	cmd           *cobra.Command
	Opts          DeleteTeamOpts
	ClientFactory func(shared.RootOptions) (Client, error)
}

type DeleteTeamOpts struct {
	shared.RootOptions
}

func AddDeleteTeamCmd(delete *cobra.Command, opts shared.RootOptions) {
	t := DeleteTeamCmd{
		cmd: &cobra.Command{
			Use:   "team",
			Short: "Delete team",
			Long:  `Delete a team from Codesphere or an Organization`,
			Example: io.FormatExampleCommands("delete team", []io.Example{
				{Cmd: "-t <teamId>", Desc: "Delete a team"},
			}),
		},
		Opts: DeleteTeamOpts{
			RootOptions: opts,
		},
		ClientFactory: func(opts shared.RootOptions) (Client, error) { return opts.NewClient() },
	}
	t.cmd.RunE = t.RunE
	shared.AddCmd(delete, t.cmd)
}

func (c *DeleteTeamCmd) RunE(_ *cobra.Command, args []string) error {
	client, err := c.ClientFactory(c.Opts.RootOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	teamId, err := c.Opts.GetTeamId()
	if err != nil {
		return errors.New("team ID not set, use -t or CS_TEAM_ID to set it")
	}

	err = client.DeleteTeam(teamId)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}
	return nil
}
