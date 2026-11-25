// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package add_test

import (
	"context"
	"os"
	"path/filepath"

	"github.com/codesphere-cloud/cs-go/cli/cmd"
	"github.com/codesphere-cloud/cs-go/cli/cmd/add"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Add SSH key", func() {
	const publicKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAABAgMEBQYHCAkKCwwNDg8QERITFBUWFxgZGhscHR4f laptop"
	var client *cmd.MockClient
	var command add.AddSSHKeyCmd
	var path string

	BeforeEach(func() {
		client = cmd.NewMockClient(GinkgoT())
		path = filepath.Join(GinkgoT().TempDir(), "id_ed25519.pub")
	})

	It("uploads the public key with its comment and trims the line ending", func() {
		Expect(os.WriteFile(path, []byte(publicKey+"\r\n"), 0o600)).To(Succeed())
		client.EXPECT().UploadSSHPublicKey(mock.Anything, publicKey).Return(nil).Once()

		Expect(command.AddSSHKey(context.Background(), client, path)).To(Succeed())
	})

	It("rejects private key files before uploading", func() {
		Expect(os.WriteFile(path, []byte("-----BEGIN OPENSSH PRIVATE KEY-----\nexample\n-----END OPENSSH PRIVATE KEY-----\n"), 0o600)).To(Succeed())

		Expect(command.AddSSHKey(context.Background(), client, path)).To(MatchError(ContainSubstring("must contain a single OpenSSH public key")))
	})

	It("uploads the public key and comment without authorized_keys directives", func() {
		keyWithOptions := `from="192.0.2.10,192.0.2.11",no-port-forwarding ` + publicKey
		Expect(os.WriteFile(path, []byte(keyWithOptions+"\n"), 0o600)).To(Succeed())
		client.EXPECT().UploadSSHPublicKey(mock.Anything, publicKey).Return(nil).Once()

		Expect(command.AddSSHKey(context.Background(), client, path)).To(Succeed())
	})

	It("rejects files containing multiple keys", func() {
		Expect(os.WriteFile(path, []byte(publicKey+"\n"+publicKey+"\n"), 0o600)).To(Succeed())

		Expect(command.AddSSHKey(context.Background(), client, path)).To(MatchError(ContainSubstring("must contain a single OpenSSH public key")))
	})

	It("reports the parser error for a malformed single-line key", func() {
		Expect(os.WriteFile(path, []byte("ssh-ed25519 !invalid!\n"), 0o600)).To(Succeed())

		Expect(command.AddSSHKey(context.Background(), client, path)).To(MatchError(ContainSubstring("ssh: no key found")))
	})
})
