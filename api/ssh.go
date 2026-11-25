// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"github.com/codesphere-cloud/cs-go/api/errors"
	"github.com/codesphere-cloud/cs-go/api/openapi_client"
)

// UploadSSHPublicKey registers a public SSH key for the authenticated user.
func (c *Client) UploadSSHPublicKey(ctx context.Context, publicKey string) error {
	response, err := c.api.SshAPI.SshUploadPublicKey(c.withAuth(ctx)).
		SshUploadPublicKeyRequest(openapi_client.SshUploadPublicKeyRequest{PublicKey: publicKey}).
		Execute()
	return errors.FormatAPIError(response, err)
}

// ListSSHPublicKeys returns the authenticated user's public SSH keys.
func (c *Client) ListSSHPublicKeys(ctx context.Context) ([]SSHPublicKey, error) {
	keys, response, err := c.api.SshAPI.SshListPublicKeys(c.withAuth(ctx)).Execute()
	return keys, errors.FormatAPIError(response, err)
}

// DeleteSSHPublicKey removes a public SSH key by fingerprint.
func (c *Client) DeleteSSHPublicKey(ctx context.Context, fingerprint string) error {
	response, err := c.api.SshAPI.SshDeletePublicKey(c.withAuth(ctx)).
		SshDeletePublicKeyRequest(openapi_client.SshDeletePublicKeyRequest{PublicKeyFingerprint: fingerprint}).
		Execute()
	return errors.FormatAPIError(response, err)
}
