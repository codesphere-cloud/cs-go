// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"errors"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetDomainsCmd", func() {
	var (
		mockClient *cmd.MockClient
		wsId       int
		domains    *api.WorkspaceDomains
	)

	BeforeEach(func() {
		mockClient = cmd.NewMockClient(GinkgoT())
		wsId = 123
		domains = &api.WorkspaceDomains{
			DevDomain:     "https://123-3000.dev.1.codesphere.com/",
			CustomDomains: []string{"custom1.example.com", "custom2.example.com"},
		}
	})

	Context("GetWorkspaceDomains", func() {
		It("retrieves workspace domains successfully", func() {
			mockClient.EXPECT().GetWorkspaceDomains(wsId).Return(domains, nil)

			result, err := mockClient.GetWorkspaceDomains(wsId)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(domains))
			Expect(result.DevDomain).To(Equal("https://123-3000.dev.1.codesphere.com/"))
			Expect(result.CustomDomains).To(HaveLen(2))
		})

		It("handles errors from API", func() {
			expectedErr := errors.New("workspace not found")
			mockClient.EXPECT().GetWorkspaceDomains(wsId).Return(nil, expectedErr)

			result, err := mockClient.GetWorkspaceDomains(wsId)

			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("handles workspace with only devDomain", func() {
			domainsNoCustm := &api.WorkspaceDomains{
				DevDomain:     "https://123-3000.dev.1.codesphere.com/",
				CustomDomains: []string{},
			}
			mockClient.EXPECT().GetWorkspaceDomains(wsId).Return(domainsNoCustm, nil)

			result, err := mockClient.GetWorkspaceDomains(wsId)

			Expect(err).ToNot(HaveOccurred())
			Expect(result.DevDomain).To(Equal("https://123-3000.dev.1.codesphere.com/"))
			Expect(result.CustomDomains).To(BeEmpty())
		})
	})
})
