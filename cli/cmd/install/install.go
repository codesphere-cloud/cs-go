// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package install

import (
	shared "github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/spf13/cobra"
)

type InstallCmd struct {
	cmd *cobra.Command
}

func AddInstallCmd(rootCmd *cobra.Command) {
	install := InstallCmd{
		cmd: &cobra.Command{
			Use:   "install",
			Short: "Install optional Codesphere CLI extensions",
			Long:  `Install optional extensions for the Codesphere CLI, such as the bundled AI agent skills.`,
		},
	}
	shared.AddCmd(rootCmd, install.cmd)

	AddInstallSkillsCmd(install.cmd)
}
