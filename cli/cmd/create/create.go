// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	shared "github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/spf13/cobra"
)

type CreateCmd struct {
	cmd *cobra.Command
}

func AddCreateCmd(rootCmd *cobra.Command, opts shared.RootOptions) {
	create := CreateCmd{
		cmd: &cobra.Command{
			Use:   "create",
			Short: "Create codesphere resource",
			Long:  `Create Codesphere resources like workspaces, environment variables, and secrets.`,
		},
	}
	shared.AddCmd(rootCmd, create.cmd)

	AddCreateWorkspaceCmd(create.cmd, opts)
	AddCreateOrganizationCmd(create.cmd, opts)
	AddCreateEnvCmd(create.cmd, opts)
	AddCreateTeamCmd(create.cmd, opts)
}
