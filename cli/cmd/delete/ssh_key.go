// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package delete

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	csio "github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type DeleteSSHKeyCmd struct {
	Opts shared.RootOptions
}

func AddDeleteSSHKeyCmd(parent *cobra.Command, opts shared.RootOptions) {
	deleteKey := DeleteSSHKeyCmd{Opts: opts}
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
	client, err := c.Opts.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}
	return c.DeleteSSHKey(cmd.Context(), client, cmd.InOrStdin(), cmd.ErrOrStderr(), args)
}

func (c *DeleteSSHKeyCmd) DeleteSSHKey(ctx context.Context, client shared.Client, input io.Reader, output io.Writer, args []string) error {
	var fingerprint string
	if len(args) == 0 {
		selected, err := selectSSHKey(ctx, client, input, output)
		if err != nil || selected == "" {
			return err
		}
		fingerprint = selected
	} else {
		fingerprint = args[0]
	}

	if strings.TrimSpace(fingerprint) == "" {
		return errors.New("SSH key fingerprint cannot be empty")
	}
	if err := client.DeleteSSHPublicKey(ctx, fingerprint); err != nil {
		return fmt.Errorf("failed to delete SSH public key: %w", err)
	}
	return nil
}

func selectSSHKey(ctx context.Context, client shared.Client, input io.Reader, output io.Writer) (string, error) {
	keys, err := client.ListSSHPublicKeys(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list SSH public keys: %w", err)
	}
	if len(keys) == 0 {
		_, err := fmt.Fprintln(output, "No public SSH keys are registered.")
		return "", err
	}

	index, selected, err := csio.SelectOption(input, output, keys, func(key api.SSHPublicKey) string {
		return fmt.Sprintf("%q (%s)", key.Title, key.PublicKeyFingerprint)
	}, "Enter the number of the SSH key to delete")
	if err != nil || !selected {
		return "", err
	}
	return keys[index].PublicKeyFingerprint, nil
}
