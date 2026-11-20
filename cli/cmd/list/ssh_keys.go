// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"
	"time"

	"github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

type ListSSHKeysCmd struct {
	Opts *ListOptions
}

func AddListSSHKeysCmd(parent *cobra.Command, opts *ListOptions) {
	listKeys := ListSSHKeysCmd{Opts: opts}
	command := &cobra.Command{
		Use:   "ssh-keys",
		Short: "List the public SSH keys on your account",
		Example: io.FormatExampleCommands("list ssh-keys", []io.Example{
			{Cmd: "--output json", Desc: "List public SSH keys in JSON format"},
		}),
		RunE: listKeys.RunE,
	}
	shared.AddCmd(parent, command)
}

func (c *ListSSHKeysCmd) RunE(_ *cobra.Command, _ []string) error {
	client, err := c.Opts.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}
	return c.ListSSHKeys(client)
}

func (c *ListSSHKeysCmd) ListSSHKeys(client shared.Client) error {
	keys, err := client.ListSSHPublicKeys()
	if err != nil {
		return fmt.Errorf("failed to list SSH public keys: %w", err)
	}

	switch c.Opts.OutputFormat {
	case shared.OutputFormatJSON:
		return io.PrintJSON(keys)
	case shared.OutputFormatYAML:
		return io.PrintYAML(keys)
	}

	t := io.GetTableWriter()
	t.AppendHeader(table.Row{"Title", "Fingerprint", "Created at"})
	for _, key := range keys {
		t.AppendRow(table.Row{key.Title, key.PublicKeyFingerprint, key.CreatedAt.Format(time.RFC3339)})
	}
	t.Render()
	return nil
}
