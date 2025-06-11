// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"fmt"

	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type Prompt interface {
	InputPrompt(prompt string) string
}

type DeleteWorkspaceCmd struct {
	cmd    *cobra.Command
	Opts   DeleteWorkspaceOpts
	Prompt Prompt
}

type DeleteWorkspaceOpts struct {
	GlobalOptions
	Confirmed *bool
}

func (c *DeleteWorkspaceCmd) RunE(_ *cobra.Command, args []string) error {
	wsId, err := c.Opts.GetWorkspaceId()
	if err != nil {
		return fmt.Errorf("failed to get workspace ID: %w", err)
	}

	client, err := NewClient(c.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	return c.DeleteWorkspace(client, wsId)
}

func AddDeleteWorkspaceCmd(delete *cobra.Command, opts GlobalOptions) {
	workspace := DeleteWorkspaceCmd{
		cmd: &cobra.Command{
			Use:   "workspace",
			Short: "Delete workspace",
			Long: io.Long(`Delete workspace after confirmation.

			Confirmation can be given interactively or with the --yes flag`),
		},
		Opts:   DeleteWorkspaceOpts{GlobalOptions: opts},
		Prompt: &io.Prompt{},
	}
	workspace.Opts.Confirmed = workspace.cmd.Flags().Bool("yes", false, "Confirm deletion of workspace")
	delete.AddCommand(workspace.cmd)
	workspace.cmd.RunE = workspace.RunE
}

func (c *DeleteWorkspaceCmd) DeleteWorkspace(client Client, wsId int) error {

	workspace, err := client.GetWorkspace(wsId)
	if err != nil {
		return fmt.Errorf("failed to get workspace %d: %w", wsId, err)
	}

	if !*c.Opts.Confirmed {
		fmt.Printf("Please confirm deletion of workspace '%s', ID %d, in team %d by entering its name:\n", workspace.Name, workspace.Id, workspace.TeamId)
		confirmation := c.Prompt.InputPrompt("Confirmation delete")

		if confirmation != workspace.Name {
			return errors.New("confirmation failed")
		}
	}

	return client.DeleteWorkspace(wsId)
}
