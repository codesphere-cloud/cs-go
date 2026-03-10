// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/codesphere-cloud/cs-go/cli/cmd"
)

var _ = Describe("ScaleWorkspace", func() {
	var (
		mockClient *cmd.MockClient
		c          *cmd.ScaleWorkspaceCmd
		wsId       int
	)

	JustBeforeEach(func() {
		mockClient = cmd.NewMockClient(GinkgoT())
		wsId = 42
		c = &cmd.ScaleWorkspaceCmd{
			Opts: cmd.ScaleWorkspaceOpts{
				GlobalOptions: &cmd.GlobalOptions{
					WorkspaceId: wsId,
				},
			},
		}
	})

	Context("ScaleWorkspaceServices", func() {
		It("should scale landscape services", func() {
			c.Opts.Services = []string{"frontend=2", "backend=3"}

			mockClient.EXPECT().ScaleLandscapeServices(wsId, map[string]int{"frontend": 2, "backend": 3}).Return(nil)

			err := c.ScaleWorkspaceServices(mockClient, wsId)

			Expect(err).ToNot(HaveOccurred())
		})

		It("should return error if ScaleLandscapeServices fails", func() {
			c.Opts.Services = []string{"web=1"}

			mockClient.EXPECT().ScaleLandscapeServices(wsId, map[string]int{"web": 1}).Return(fmt.Errorf("scale error"))

			err := c.ScaleWorkspaceServices(mockClient, wsId)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to scale landscape services"))
		})

		It("should return error for invalid service format", func() {
			c.Opts.Services = []string{"invalid"}

			err := c.ScaleWorkspaceServices(mockClient, wsId)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to parse services"))
		})

		It("should allow scaling to 0 replicas", func() {
			c.Opts.Services = []string{"api=0"}

			mockClient.EXPECT().ScaleLandscapeServices(wsId, map[string]int{"api": 0}).Return(nil)

			err := c.ScaleWorkspaceServices(mockClient, wsId)

			Expect(err).ToNot(HaveOccurred())
		})

		It("should return error for negative replica count", func() {
			c.Opts.Services = []string{"web=-1"}

			err := c.ScaleWorkspaceServices(mockClient, wsId)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("replica count must be non-negative"))
		})

		It("should return error for empty service name", func() {
			c.Opts.Services = []string{"=2"}

			err := c.ScaleWorkspaceServices(mockClient, wsId)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("empty service name"))
		})

		It("should return error for non-numeric replica count", func() {
			c.Opts.Services = []string{"web=abc"}

			err := c.ScaleWorkspaceServices(mockClient, wsId)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid replica count"))
		})
	})
})
