// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/codesphere-cloud/cs-go/pkg/io"
)

type SyncLandscapeCmd struct {
	cmd  *cobra.Command
	Opts SyncLandscapeOpts
}

type SyncLandscapeOpts struct {
	*GlobalOptions

	Profile string
}

func (c *SyncLandscapeCmd) RunE(_ *cobra.Command, args []string) error {
	workspaceId, err := c.Opts.GetWorkspaceId()
	if err != nil {
		return fmt.Errorf("failed to get workspace ID: %w", err)
	}

	client, err := NewClient(*c.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	return c.SyncLandscape(client, workspaceId)
}

func AddSyncLandscapeCmd(sync *cobra.Command, opts *GlobalOptions) {
	workspace := SyncLandscapeCmd{
		cmd: &cobra.Command{
			Use:   "landscape",
			Short: "Sync landscape",
			Long:  io.Long(`Sync landscape according to CI profile, i.e. allocate resources for defined services.`),
		},
		Opts: SyncLandscapeOpts{GlobalOptions: opts},
	}

	workspace.cmd.Flags().StringVarP(&workspace.Opts.Profile, "profile", "p", "", "CI profile to use (e.g. 'prod' for the profile defined in 'ci.prod.yml'), defaults to the ci.yml profile")

	workspace.cmd.RunE = workspace.RunE

	sync.AddCommand(workspace.cmd)
}

func (c *SyncLandscapeCmd) SyncLandscape(client Client, wsId int) error {
	return client.DeployLandscape(wsId, c.Opts.Profile)
	//TODO: Wait for deployment to be synced if possible
}
