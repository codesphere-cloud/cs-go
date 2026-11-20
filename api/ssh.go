// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"github.com/codesphere-cloud/cs-go/api/errors"
	"github.com/codesphere-cloud/cs-go/api/openapi_client"
)

// UploadSSHPublicKey registers a public SSH key for the authenticated user.
func (c *Client) UploadSSHPublicKey(publicKey string) error {
	response, err := c.api.SshAPI.SshUploadPublicKey(c.ctx).
		SshUploadPublicKeyRequest(openapi_client.SshUploadPublicKeyRequest{PublicKey: publicKey}).
		Execute()
	return errors.FormatAPIError(response, err)
}

// ListSSHPublicKeys returns the authenticated user's public SSH keys.
func (c *Client) ListSSHPublicKeys() ([]SSHPublicKey, error) {
	keys, response, err := c.api.SshAPI.SshListPublicKeys(c.ctx).Execute()
	return keys, errors.FormatAPIError(response, err)
}

// DeleteSSHPublicKey removes a public SSH key by fingerprint.
func (c *Client) DeleteSSHPublicKey(fingerprint string) error {
	response, err := c.api.SshAPI.SshDeletePublicKey(c.ctx).
		SshDeletePublicKeyRequest(openapi_client.SshDeletePublicKeyRequest{PublicKeyFingerprint: fingerprint}).
		Execute()
	return errors.FormatAPIError(response, err)
}
