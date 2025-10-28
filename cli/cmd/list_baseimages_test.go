// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
)

var _ = Describe("ListBaseimagesCmd", func() {
	var (
		c          cmd.ListBaseimagesCmd
		globalOpts cmd.GlobalOptions
		mockEnv    *cmd.MockEnv
		mockClient *cmd.MockClient
	)

	BeforeEach(func() {
		mockEnv = cmd.NewMockEnv(GinkgoT())
		mockClient = cmd.NewMockClient(GinkgoT())
		globalOpts = cmd.GlobalOptions{
			Env: mockEnv,
		}
		c = cmd.ListBaseimagesCmd{
			Opts: globalOpts,
		}
	})

	AfterEach(func() {
		mockEnv.AssertExpectations(GinkgoT())
		mockClient.AssertExpectations(GinkgoT())
	})

	Context("RunE method", func() {
		It("fails when client creation fails due to missing API token", func() {
			// Mock failure to get API token
			mockEnv.EXPECT().GetApiToken().Return("", errors.New("CS_TOKEN env var required, but not set"))

			err := c.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to create Codesphere client"))
		})

		It("fails when client creation fails due to invalid API URL", func() {
			// Mock successful token but invalid URL
			mockEnv.EXPECT().GetApiToken().Return("test-token", nil)
			mockEnv.EXPECT().GetApiUrl().Return("ht tp://invalid url with spaces")

			err := c.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to create Codesphere client"))
		})

		It("successfully creates client and fails on list baseimages with real client", func() {
			// Mock successful client creation
			mockEnv.EXPECT().GetApiToken().Return("test-token", nil)
			mockEnv.EXPECT().GetApiUrl().Return("https://api.codesphere.com")

			err := c.RunE(nil, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to list baseimages"))
		})
	})

	Context("ListBaseimages method", func() {
		It("successfully lists baseimages", func() {
			supportedUntil, _ := time.Parse("2006-01-02", "2025-12-31")
			expectedBaseimages := []api.Baseimage{
				{Id: "ubuntu-20", Name: "Ubuntu 20.04", SupportedUntil: supportedUntil},
				{Id: "node-18", Name: "Node.js 18", SupportedUntil: supportedUntil},
			}

			mockClient.EXPECT().ListBaseimages().Return(expectedBaseimages, nil)

			err := c.ListBaseimages(mockClient)
			Expect(err).To(BeNil())
		})

		It("returns error when no baseimages available but client fails", func() {
			mockClient.EXPECT().ListBaseimages().Return([]api.Baseimage{}, errors.New("API error"))

			err := c.ListBaseimages(mockClient)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to list baseimages: API error"))
		})

		It("succeeds when empty list is returned", func() {
			mockClient.EXPECT().ListBaseimages().Return([]api.Baseimage{}, nil)

			err := c.ListBaseimages(mockClient)
			Expect(err).To(BeNil())
		})
	})
})

var _ = Describe("AddListBaseimagesCmd", func() {
	var (
		parentCmd  *cobra.Command
		globalOpts cmd.GlobalOptions
	)

	BeforeEach(func() {
		parentCmd = &cobra.Command{Use: "list"}
		globalOpts = cmd.GlobalOptions{}
	})

	It("adds the baseimages command with correct properties", func() {
		cmd.AddListBaseimagesCmd(parentCmd, globalOpts)

		var baseimagesCmd *cobra.Command
		for _, c := range parentCmd.Commands() {
			if c.Use == "baseimages" {
				baseimagesCmd = c
				break
			}
		}

		Expect(baseimagesCmd).NotTo(BeNil())
		Expect(baseimagesCmd.Use).To(Equal("baseimages"))
		Expect(baseimagesCmd.Short).To(Equal("List baseimages"))
		Expect(baseimagesCmd.Long).To(Equal("List baseimages available in Codesphere for workspace creation"))
		Expect(baseimagesCmd.RunE).NotTo(BeNil())
	})
})
