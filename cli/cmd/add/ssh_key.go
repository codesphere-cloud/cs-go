// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"bytes"
	"fmt"
	"os"

	"github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

type AddSSHKeyCmd struct {
	ClientProvider shared.ClientProvider
}

func AddAddSSHKeyCmd(parent *cobra.Command, opts shared.RootOptions) {
	addKey := AddSSHKeyCmd{ClientProvider: opts}
	command := &cobra.Command{
		Use:   "ssh-key <public-key-file>",
		Short: "Add a public SSH key to your account",
		Long:  "Upload an OpenSSH public key for workspace SSH access using the account associated with your API token.",
		Args:  cobra.ExactArgs(1),
		Example: io.FormatExampleCommands("add ssh-key", []io.Example{
			{Cmd: "~/.ssh/id_ed25519.pub", Desc: "Upload your Ed25519 public key"},
		}),
		RunE: addKey.RunE,
	}
	shared.AddCmd(parent, command)
}

func (c *AddSSHKeyCmd) RunE(_ *cobra.Command, args []string) error {
	client, err := c.ClientProvider.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}
	return c.AddSSHKey(client, args[0])
}

func (c *AddSSHKeyCmd) AddSSHKey(client shared.Client, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read SSH public key: %w", err)
	}
	data = bytes.TrimSpace(data)
	_, _, options, _, err := ssh.ParseAuthorizedKey(data)
	if err != nil || len(options) != 0 || bytes.ContainsAny(data, "\r\n") {
		return fmt.Errorf("%q must contain a single OpenSSH public key", path)
	}
	if err := client.UploadSSHPublicKey(string(data)); err != nil {
		return fmt.Errorf("failed to add SSH public key: %w", err)
	}
	return nil
}
