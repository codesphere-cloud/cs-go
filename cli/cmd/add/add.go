// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	shared "github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/spf13/cobra"
)

type AddCmd struct {
	cmd *cobra.Command
}

func AddAddCmd(rootCmd *cobra.Command, opts shared.RootOptions) {
	add := AddCmd{
		cmd: &cobra.Command{
			Use:   "add",
			Short: "Add Codesphere resources",
			Long:  `Add resources to existing Codesphere resources, e.g. team members.`,
		},
	}
	shared.AddCmd(rootCmd, add.cmd)

	AddAddTeamMemberCmd(add.cmd, opts)
}
