// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package delete

import (
	shared "github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/spf13/cobra"
)

type DeleteCmd struct {
	cmd *cobra.Command
}

func AddDeleteCmd(rootCmd *cobra.Command, opts shared.RootOptions) {
	delete := DeleteCmd{
		cmd: &cobra.Command{
			Use:   "delete",
			Short: "Delete Codesphere resources",
			Long:  `Delete Codesphere resources, e.g. workspaces or teams.`,
		},
	}
	shared.AddCmd(rootCmd, delete.cmd)

	AddDeleteWorkspaceCmd(delete.cmd, opts)
	AddDeleteTeamCmd(delete.cmd, opts)
	AddDeleteTeamMemberCmd(delete.cmd, opts)
}
