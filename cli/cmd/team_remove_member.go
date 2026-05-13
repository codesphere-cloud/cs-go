package cmd

import (
	"github.com/spf13/cobra"
)

type RemoveTeamMemberCmd struct {
	cmd  *cobra.Command
	Opts CreateWorkspaceOpts
}

func AddRemoveTeamMemberCmd(team *cobra.Command, opts *GlobalOptions) {
	res := RemoveTeamMemberCmd{
		cmd: &cobra.Command{
			Use:   "remove",
			Short: "Remove team member",
			Long:  `Remove team member from a team`,
		},
		Opts: CreateWorkspaceOpts{
			GlobalOptions: opts,
		},
	}
	res.cmd.RunE = res.RunE
	AddCmd(team, res.cmd)
}

func (c *RemoveTeamMemberCmd) RunE(_ *cobra.Command, args []string) error {
	// TODO: Implement team member removal logic
	return nil
}
