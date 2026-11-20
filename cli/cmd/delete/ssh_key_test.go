// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package delete_test

import (
	"bytes"
	"strings"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
	"github.com/codesphere-cloud/cs-go/cli/cmd/delete"
	"github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("Delete SSH key", func() {
	var client *cmd.MockClient
	var command delete.DeleteSSHKeyCmd
	var leaf *cobra.Command
	var output *bytes.Buffer
	keys := []api.SSHPublicKey{
		{Title: "Laptop", PublicKeyFingerprint: "SHA256:first"},
		{Title: "Desktop", PublicKeyFingerprint: "SHA256:second"},
	}

	BeforeEach(func() {
		client = cmd.NewMockClient(GinkgoT())
		provider := shared.NewMockClientProvider(GinkgoT())
		provider.EXPECT().NewClient().Return(client, nil).Once()
		command = delete.DeleteSSHKeyCmd{ClientProvider: provider}
		leaf = &cobra.Command{}
		output = &bytes.Buffer{}
		leaf.SetErr(output)
	})

	It("deletes a supplied fingerprint directly", func() {
		client.EXPECT().DeleteSSHPublicKey("SHA256:second").Return(nil).Once()

		Expect(command.RunE(leaf, []string{"SHA256:second"})).To(Succeed())
		Expect(output.String()).To(BeEmpty())
	})

	It("lists keys and deletes the selected fingerprint", func() {
		client.EXPECT().ListSSHPublicKeys().Return(keys, nil).Once()
		client.EXPECT().DeleteSSHPublicKey("SHA256:second").Return(nil).Once()
		leaf.SetIn(strings.NewReader("2\n"))

		Expect(command.RunE(leaf, nil)).To(Succeed())
		Expect(output.String()).To(ContainSubstring(`1. "Laptop" (SHA256:first)`))
		Expect(output.String()).To(ContainSubstring(`2. "Desktop" (SHA256:second)`))
	})

	It("cancels on an empty selection", func() {
		client.EXPECT().ListSSHPublicKeys().Return(keys, nil).Once()
		leaf.SetIn(strings.NewReader("\n"))

		Expect(command.RunE(leaf, nil)).To(Succeed())
	})

	It("rejects a selection outside the displayed list", func() {
		client.EXPECT().ListSSHPublicKeys().Return(keys, nil).Once()
		leaf.SetIn(strings.NewReader("3\n"))

		Expect(command.RunE(leaf, nil)).To(MatchError(ContainSubstring("enter a number from 1 to 2")))
	})

	It("reports an empty key list without prompting", func() {
		client.EXPECT().ListSSHPublicKeys().Return(nil, nil).Once()

		Expect(command.RunE(leaf, nil)).To(Succeed())
		Expect(output.String()).To(Equal("No public SSH keys are registered.\n"))
	})
})
