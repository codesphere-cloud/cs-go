// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/spf13/cobra"

	"github.com/codesphere-cloud/cs-go/pkg/cs"
)

type UpdateCmd struct {
	cmd *cobra.Command
}

func (c *UpdateCmd) RunE(_ *cobra.Command, args []string) error {

	return SelfUpdate()
}

func AddUpdateCmd(rootCmd *cobra.Command) {
	update := UpdateCmd{
		cmd: &cobra.Command{
			Use:   "update",
			Short: "Update Codesphere CLI",
			Long:  `Updates the Codesphere CLI to the latest release from GitHub.`,
		},
	}
	rootCmd.AddCommand(update.cmd)
	update.cmd.RunE = update.RunE
}

func SelfUpdate() error {
	v := semver.MustParse(cs.Version())
	latest, err := selfupdate.UpdateSelf(v, "codesphere-cloud/cs-go")
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}
	if latest.Version.Equals(v) {
		fmt.Println("Current cs CLI is the latest version", cs.Version())
		return nil
	}
	fmt.Println("Successfully updated to version", latest.Version)
	fmt.Println("Release notes:\n", latest.ReleaseNotes)
	return nil
}
