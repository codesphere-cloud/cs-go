// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

type TeamManageCmd struct {
	cmd *cobra.Command
}

func AddTeamManageCmd(rootCmd *cobra.Command, opts *GlobalOptions) {
	t := TeamManageCmd{
		cmd: &cobra.Command{
			Use:   "team",
			Short: "Manage Team",
			Long:  `Manage Team Resources like Members or Roles in Teams`,
		},
	}

	AddCmd(rootCmd, t.cmd)
	AddMemberCmd(t.cmd, opts)
	AddCreateTeamCmd(t.cmd, opts)
	AddRemoveTeamCmd(t.cmd, opts)
}

func AddMemberCmd(t *cobra.Command, opts *GlobalOptions) {

	memberCmd := &cobra.Command{
		Use:   "member",
		Short: "Manage team members",
	}

	AddAddTeamMemberCmd(memberCmd, opts)
	AddRemoveTeamMemberCmd(memberCmd, opts)
	AddListTeamMembersCmd(memberCmd, opts)

	t.AddCommand(memberCmd)
}
