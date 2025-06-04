// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/codesphere-cloud/cs-go/cli/cmd"
)

var _ = Describe("SetEnvVars", func() {
	var (
		mockEnv    *cmd.MockEnv
		mockClient *cmd.MockClient
		e          *cmd.SetEnvVarCmd
		envVars    []string
		wsId       int
	)

	JustBeforeEach(func() {
		mockEnv = cmd.NewMockEnv(GinkgoT())
		mockClient = cmd.NewMockClient(GinkgoT())
		wsId = 42
		e = &cmd.SetEnvVarCmd{
			Opts: cmd.SetEnvVarOptions{
				GlobalOptions: cmd.GlobalOptions{
					Env:         mockEnv,
					WorkspaceId: &wsId,
				},
				EnvVar: &envVars,
			},
		}
	})

	Context("Multiple env vars", func() {
		BeforeEach(func() {
			envVars = []string{"hello=world", "a=b"}
		})
		It("Sets all env vars passed in", func() {
			expectedVars := map[string]string{"hello": "world", "a": "b"}
			mockClient.EXPECT().SetEnvVarOnWorkspace(wsId, expectedVars).Return(nil)

			err := e.SetEnvironmentVariables(mockClient)
			Expect(err).NotTo(HaveOccurred())
		})

	})

	Context("Single env var", func() {
		BeforeEach(func() {
			envVars = []string{"a=b"}
		})
		It("Sets env var", func() {
			expectedVars := map[string]string{"a": "b"}
			mockClient.EXPECT().SetEnvVarOnWorkspace(wsId, expectedVars).Return(nil)

			err := e.SetEnvironmentVariables(mockClient)
			Expect(err).NotTo(HaveOccurred())
		})

	})

	Context("Malformed env vars", func() {
		BeforeEach(func() {
			envVars = []string{"helloworld", "a=b"}
		})
		It("doesn't set environment variables", func() {
			err := e.SetEnvironmentVariables(mockClient)
			Expect(err).To(MatchError("failed to parse environment variables: invalid environment variable argument: helloworld"))
		})

	})

})
