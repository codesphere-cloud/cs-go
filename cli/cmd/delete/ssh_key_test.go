// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package delete_test

import (
	"bytes"
	"context"
	"strings"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
	"github.com/codesphere-cloud/cs-go/cli/cmd/delete"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Delete SSH key", func() {
	var client *cmd.MockClient
	var command delete.DeleteSSHKeyCmd
	var output *bytes.Buffer
	keys := []api.SSHPublicKey{
		{Title: "Laptop", PublicKeyFingerprint: "SHA256:first"},
		{Title: "Desktop", PublicKeyFingerprint: "SHA256:second"},
	}

	BeforeEach(func() {
		client = cmd.NewMockClient(GinkgoT())
		output = &bytes.Buffer{}
	})

	It("deletes a supplied fingerprint directly", func() {
		client.EXPECT().DeleteSSHPublicKey(mock.Anything, "SHA256:second").Return(nil).Once()

		Expect(command.DeleteSSHKey(context.Background(), client, nil, output, []string{"SHA256:second"})).To(Succeed())
		Expect(output.String()).To(BeEmpty())
	})

	It("lists keys and deletes the selected fingerprint", func() {
		client.EXPECT().ListSSHPublicKeys(mock.Anything).Return(keys, nil).Once()
		client.EXPECT().DeleteSSHPublicKey(mock.Anything, "SHA256:second").Return(nil).Once()

		Expect(command.DeleteSSHKey(context.Background(), client, strings.NewReader("2\n"), output, nil)).To(Succeed())
		Expect(output.String()).To(ContainSubstring(`1. "Laptop" (SHA256:first)`))
		Expect(output.String()).To(ContainSubstring(`2. "Desktop" (SHA256:second)`))
	})

	It("cancels on an empty selection", func() {
		client.EXPECT().ListSSHPublicKeys(mock.Anything).Return(keys, nil).Once()

		Expect(command.DeleteSSHKey(context.Background(), client, strings.NewReader("\n"), output, nil)).To(Succeed())
	})

	It("rejects a selection outside the displayed list", func() {
		client.EXPECT().ListSSHPublicKeys(mock.Anything).Return(keys, nil).Once()

		Expect(command.DeleteSSHKey(context.Background(), client, strings.NewReader("3\n"), output, nil)).To(MatchError(ContainSubstring("enter a number from 1 to 2")))
	})

	It("reports an empty key list without prompting", func() {
		client.EXPECT().ListSSHPublicKeys(mock.Anything).Return(nil, nil).Once()

		Expect(command.DeleteSSHKey(context.Background(), client, nil, output, nil)).To(Succeed())
		Expect(output.String()).To(Equal("No public SSH keys are registered.\n"))
	})
})
