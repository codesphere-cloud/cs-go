// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

type DeployCmd struct {
	cmd *cobra.Command
}

func AddDeployCmd(rootCmd *cobra.Command, opts GlobalOptions) {
	deploy := DeployCmd{
		cmd: &cobra.Command{
			Use:   "deploy",
			Short: "Deploy workspaces from CI/CD",
			Long:  `Deploy workspaces from CI/CD pipelines. Supports creating, updating, and deleting workspaces tied to git provider events like pull requests.`,
		},
	}
	rootCmd.AddCommand(deploy.cmd)

	// Add provider-specific subcommands
	AddDeployGitHubCmd(deploy.cmd, opts)
}
