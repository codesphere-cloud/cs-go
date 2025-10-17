package cmd

import (
	"github.com/spf13/cobra"

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
	rootCmd.AddCommand(sync.cmd)

	AddSyncLandscapeCmd(sync.cmd, opts)
}
