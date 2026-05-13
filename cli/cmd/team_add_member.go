package cmd

import (
	"github.com/spf13/cobra"
)

type AddTeamMemberCmd struct {
	cmd  *cobra.Command
	Opts CreateWorkspaceOpts
}

func AddAddTeamMemberCmd(team *cobra.Command, opts *GlobalOptions) {
	t := AddTeamMemberCmd{
		cmd: &cobra.Command{
			Use:   "add",
			Short: "Add team member",
			Long:  `Add team member to a team`,
		},
		Opts: CreateWorkspaceOpts{
			GlobalOptions: opts,
		},
	}
	t.cmd.RunE = t.RunE
	AddCmd(team, t.cmd)
}

func (c *AddTeamMemberCmd) RunE(_ *cobra.Command, args []string) error {
	return nil
}
