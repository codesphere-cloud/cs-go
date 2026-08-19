// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"

	shared "github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/codesphere-cloud/cs-go/pkg/io"
)

type SyncCmd struct {
	cmd *cobra.Command
}

func AddSyncCmd(rootCmd *cobra.Command, opts *GlobalOptions) {
	sync := SyncCmd{
		cmd: &cobra.Command{
			Use:   "sync",
			Short: "Sync Codesphere resources",
			Long:  io.Long(`Synchronize Codesphere resources, like infrastructure required to run services.`),
		},
	}
	shared.AddCmd(rootCmd, sync.cmd)

	AddSyncLandscapeCmd(sync.cmd, opts)
}
