// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package delete

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	csio "github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type DeleteSSHKeyCmd struct {
	ClientProvider shared.ClientProvider
}

func AddDeleteSSHKeyCmd(parent *cobra.Command, opts shared.RootOptions) {
	deleteKey := DeleteSSHKeyCmd{ClientProvider: opts}
	command := &cobra.Command{
		Use:   "ssh-key [fingerprint]",
		Short: "Delete a public SSH key from your account",
		Long:  "Delete a public SSH key by fingerprint, or select a key interactively to delete it.",
		Args:  cobra.MaximumNArgs(1),
		Example: csio.FormatExampleCommands("delete ssh-key", []csio.Example{
			{Cmd: "", Desc: "Select a public SSH key to delete"},
			{Cmd: "<fingerprint>", Desc: "Delete a public SSH key by fingerprint"},
		}),
		RunE: deleteKey.RunE,
	}
	shared.AddCmd(parent, command)
}

func (c *DeleteSSHKeyCmd) RunE(cmd *cobra.Command, args []string) error {
	client, err := c.ClientProvider.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}
	if len(args) == 1 {
		return c.DeleteSSHKey(client, args[0])
	}

	fingerprint, err := selectSSHKey(client, cmd.InOrStdin(), cmd.ErrOrStderr())
	if err != nil || fingerprint == "" {
		return err
	}
	return c.DeleteSSHKey(client, fingerprint)
}

func (c *DeleteSSHKeyCmd) DeleteSSHKey(client shared.Client, fingerprint string) error {
	if strings.TrimSpace(fingerprint) == "" {
		return errors.New("SSH key fingerprint cannot be empty")
	}
	if err := client.DeleteSSHPublicKey(fingerprint); err != nil {
		return fmt.Errorf("failed to delete SSH public key: %w", err)
	}
	return nil
}

func selectSSHKey(client shared.Client, input io.Reader, output io.Writer) (string, error) {
	keys, err := client.ListSSHPublicKeys()
	if err != nil {
		return "", fmt.Errorf("failed to list SSH public keys: %w", err)
	}
	if len(keys) == 0 {
		_, err := fmt.Fprintln(output, "No public SSH keys are registered.")
		return "", err
	}

	options := make([]string, len(keys))
	for i, key := range keys {
		options[i] = fmt.Sprintf("%q (%s)", key.Title, key.PublicKeyFingerprint)
	}
	index, selected, err := csio.SelectOption(input, output, options, "Enter the number of the SSH key to delete")
	if err != nil || !selected {
		return "", err
	}
	return keys[index].PublicKeyFingerprint, nil
}
