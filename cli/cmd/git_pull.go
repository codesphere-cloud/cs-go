package cmd

import (
	"fmt"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

// GitPullCmd represents the pull command
type GitPullCmd struct {
	cmd  *cobra.Command
	Opts GitPullOpts
}

type GitPullOpts struct {
	GlobalOptions
	Remote *string
	Branch *string
}

func (c *GitPullCmd) RunE(_ *cobra.Command, args []string) error {
	wsId, err := c.Opts.GetWorkspaceId()
	if err != nil {
		return fmt.Errorf("failed to get workspace ID: %w", err)
	}

	client, err := NewClient(c.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}
	return client.GitPull(wsId, *c.Opts.Remote, *c.Opts.Branch)
}

func AddGitPullCmd(git *cobra.Command, opts GlobalOptions) {
	pull := GitPullCmd{
		cmd: &cobra.Command{
			Use:   "pull",
			Short: "Pull latest changes from git repository",
			Long: io.Long(`Pull latest changes from the remote git repository.

				if specified, pulls a specific branch.`),
			Example: io.FormatExampleCommands("pull", []io.Example{
				{Cmd: "", Desc: "Pull latest HEAD from current branch"},
				{Cmd: "--remote origin --branch staging", Desc: "Pull branch staging from remote origin"},
			}),
		},
		Opts: GitPullOpts{GlobalOptions: opts},
	}

	git.AddCommand(pull.cmd)
	pull.Opts.Branch = git.Flags().String("branch", "", "Branch to pull")
	pull.Opts.Remote = git.Flags().String("remote", "", "Remote to pull from")
	pull.cmd.RunE = pull.RunE
}
