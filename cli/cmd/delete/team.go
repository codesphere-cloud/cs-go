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
	ClientFactory func(shared.RootOptions) (shared.Client, error)
}

type DeleteTeamOpts struct {
	shared.RootOptions
}

func AddDeleteTeamCmd(delete *cobra.Command, opts shared.RootOptions) {
	t := DeleteTeamCmd{
		cmd: &cobra.Command{
			Use:     "resource-group",
			Aliases: []string{"team"},
			Short:   "Delete resource group",
			Long:    `Delete a resource group from Codesphere or an Organization`,
			Example: io.FormatExampleCommands("delete resource-group", []io.Example{
				{Cmd: "-t <resourceGroupId>", Desc: "Delete a resource group"},
			}),
		},
		Opts: DeleteTeamOpts{
			RootOptions: opts,
		},
		ClientFactory: func(opts shared.RootOptions) (shared.Client, error) { return opts.NewClient() },
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
		return errors.New("resource group ID not set, use --resource-group or CS_RESOURCE_GROUP_ID to set it")
	}

	err = client.DeleteTeam(teamId)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}
	return nil
}
