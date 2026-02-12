// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
)

var _ = Describe("Curl", func() {
	var (
		mockEnv      *cmd.MockEnv
		mockClient   *cmd.MockClient
		mockExecutor *cmd.MockCommandExecutor
		c            *cmd.CurlCmd
		wsId         int
		teamId       int
		token        string
		devDomain    string
		workspace    api.Workspace
	)

	JustBeforeEach(func() {
		mockClient = cmd.NewMockClient(GinkgoT())
		mockEnv = cmd.NewMockEnv(GinkgoT())
		mockExecutor = cmd.NewMockCommandExecutor(GinkgoT())
		wsId = 42
		teamId = 21
		token = "test-api-token"
		devDomain = "42-3000.dev.5.codesphere.com"
		workspace = api.Workspace{
			Id:        wsId,
			TeamId:    teamId,
			Name:      "test-workspace",
			DevDomain: &devDomain,
		}
		c = &cmd.CurlCmd{
			Opts: cmd.GlobalOptions{
				Env:         mockEnv,
				WorkspaceId: &wsId,
			},
			Timeout:  30 * time.Second,
			Executor: mockExecutor,
		}
	})

	Context("CurlWorkspace", func() {
		It("should construct the correct URL with default port", func() {
			mockClient.EXPECT().GetWorkspace(wsId).Return(workspace, nil)
			mockExecutor.EXPECT().Execute(
				mock.Anything,
				"curl",
				mock.MatchedBy(func(args []string) bool {
					// Verify the args contain the expected header, flag, and URL
					hasHeader := false
					hasFlag := false
					hasURL := false
					for i, arg := range args {
						if arg == "-H" && i+1 < len(args) && args[i+1] == fmt.Sprintf("x-forward-security: %s", token) {
							hasHeader = true
						}
						if arg == "-I" {
							hasFlag = true
						}
						if arg == "https://42-3000.dev.5.codesphere.com/api/health" {
							hasURL = true
						}
					}
					return hasHeader && hasFlag && hasURL
				}),
				mock.Anything,
				mock.Anything,
			).Return(nil)

			err := c.CurlWorkspace(mockClient, wsId, token, "/api/health", []string{"-I"})

			Expect(err).ToNot(HaveOccurred())
		})

		It("should construct the correct URL with custom path", func() {
			mockClient.EXPECT().GetWorkspace(wsId).Return(workspace, nil)
			mockExecutor.EXPECT().Execute(
				mock.Anything,
				"curl",
				mock.MatchedBy(func(args []string) bool {
					// Verify the URL contains the custom path
					hasHeader := false
					hasURL := false
					for i, arg := range args {
						if arg == "-H" && i+1 < len(args) && args[i+1] == fmt.Sprintf("x-forward-security: %s", token) {
							hasHeader = true
						}
						if arg == "https://42-3000.dev.5.codesphere.com/custom/path" {
							hasURL = true
						}
					}
					return hasHeader && hasURL
				}),
				mock.Anything,
				mock.Anything,
			).Return(nil)

			err := c.CurlWorkspace(mockClient, wsId, token, "/custom/path", []string{})

			Expect(err).ToNot(HaveOccurred())
		})

		It("should pass insecure flag when specified", func() {
			c.Insecure = true
			mockClient.EXPECT().GetWorkspace(wsId).Return(workspace, nil)
			mockExecutor.EXPECT().Execute(
				mock.Anything,
				"curl",
				mock.MatchedBy(func(args []string) bool {
					// Verify the insecure flag is present
					hasInsecure := false
					hasHeader := false
					hasURL := false
					for i, arg := range args {
						if arg == "-k" {
							hasInsecure = true
						}
						if arg == "-H" && i+1 < len(args) && args[i+1] == fmt.Sprintf("x-forward-security: %s", token) {
							hasHeader = true
						}
						if arg == "https://42-3000.dev.5.codesphere.com/" {
							hasURL = true
						}
					}
					return hasInsecure && hasHeader && hasURL
				}),
				mock.Anything,
				mock.Anything,
			).Return(nil)

			err := c.CurlWorkspace(mockClient, wsId, token, "/", []string{})

			Expect(err).ToNot(HaveOccurred())
		})

		It("should return error if workspace has no dev domain", func() {
			workspace := api.Workspace{
				Id:        wsId,
				TeamId:    teamId,
				Name:      "test-workspace",
				DevDomain: nil,
			}

			mockClient.EXPECT().GetWorkspace(wsId).Return(workspace, nil)

			err := c.CurlWorkspace(mockClient, wsId, token, "/", []string{})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("does not have a dev domain configured"))
		})

		It("should return error if GetWorkspace fails", func() {
			mockClient.EXPECT().GetWorkspace(wsId).Return(api.Workspace{}, fmt.Errorf("api error"))

			err := c.CurlWorkspace(mockClient, wsId, token, "/", []string{})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get workspace"))
		})

		It("should return error if command execution fails", func() {
			mockClient.EXPECT().GetWorkspace(wsId).Return(workspace, nil)
			mockExecutor.EXPECT().Execute(
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(fmt.Errorf("command failed"))

			err := c.CurlWorkspace(mockClient, wsId, token, "/", []string{})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("command failed"))
		})
	})
})
