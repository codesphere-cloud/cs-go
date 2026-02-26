// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"log"

	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

// GitPullCmd represents the pull command
type GitPullCmd struct {
	cmd  *cobra.Command
	Opts GitPullOpts
}

type GitPullOpts struct {
	*GlobalOptions
	Remote *string
	Branch *string
}

func (c *GitPullCmd) RunE(_ *cobra.Command, args []string) error {
	wsId, err := c.Opts.GetWorkspaceId()
	if err != nil {
		return fmt.Errorf("failed to get workspace ID: %w", err)
	}

	client, err := NewClient(*c.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	err = client.GitPull(wsId, *c.Opts.Remote, *c.Opts.Branch)
	if err != nil {
		return err
	}

	log.Printf("Git pull completed successfully for workspace %d\n", wsId)
	return nil
}

func AddGitPullCmd(git *cobra.Command, opts *GlobalOptions) {
	pull := GitPullCmd{
		cmd: &cobra.Command{
			Use:   "pull",
			Short: "Pull latest changes from git repository",
			Long: io.Long(`Pull latest changes from the remote git repository.

				if specified, pulls a specific branch.`),
			Example: io.FormatExampleCommands("git pull", []io.Example{
				{Cmd: "", Desc: "Pull latest HEAD from current branch"},
				{Cmd: "--remote origin --branch staging", Desc: "Pull branch staging from remote origin"},
			}),
		},
		Opts: GitPullOpts{GlobalOptions: opts},
	}

	git.AddCommand(pull.cmd)
	pull.Opts.Branch = pull.cmd.Flags().String("branch", "", "Branch to pull")
	pull.Opts.Remote = pull.cmd.Flags().String("remote", "", "Remote to pull from")
	pull.cmd.RunE = pull.RunE
}
