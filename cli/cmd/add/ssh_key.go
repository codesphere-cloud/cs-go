// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

type AddSSHKeyCmd struct {
	Opts shared.RootOptions
}

func AddAddSSHKeyCmd(parent *cobra.Command, opts shared.RootOptions) {
	addKey := AddSSHKeyCmd{Opts: opts}
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

func (c *AddSSHKeyCmd) RunE(cmd *cobra.Command, args []string) error {
	client, err := c.Opts.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}
	return c.AddSSHKey(cmd.Context(), client, args[0])
}

func (c *AddSSHKeyCmd) AddSSHKey(ctx context.Context, client shared.Client, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read SSH public key: %w", err)
	}
	data = bytes.TrimSpace(data)
	// ParseAuthorizedKey reads one entry at a time and skips invalid lines.
	if bytes.ContainsAny(data, "\r\n") {
		return fmt.Errorf("%q must contain a single OpenSSH public key", path)
	}
	// Ignore authorized_keys directives such as no-port-forwarding. They
	// configure sshd on the host that owns the authorized_keys file.
	key, comment, _, _, err := ssh.ParseAuthorizedKey(data)
	if err != nil {
		return fmt.Errorf("failed to parse SSH public key in %q: %w", path, err)
	}
	publicKey := strings.TrimSuffix(string(ssh.MarshalAuthorizedKey(key)), "\n")
	if comment != "" {
		publicKey += " " + comment
	}
	if err := client.UploadSSHPublicKey(ctx, publicKey); err != nil {
		return fmt.Errorf("failed to add SSH public key: %w", err)
	}
	return nil
}
