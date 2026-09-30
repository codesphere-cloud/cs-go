// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package list_test

import (
	"encoding/json"
	"errors"
	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
	subject "github.com/codesphere-cloud/cs-go/cli/cmd/list"
	"github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.yaml.in/yaml/v2"
	"io"
	"os"
	"time"
)

var _ = Describe("ListOrganizationMembers", func() {
	const orgId = "8316ee2f-87f8-424e-b925-7382dc50d662"
	var client *cmd.MockClient
	var env *cmd.MockEnv
	var opts *cmd.GlobalOptions
	var c *subject.ListOrganizationMembersCmd
	BeforeEach(func() {
		client = cmd.NewMockClient(GinkgoT())
		env = cmd.NewMockEnv(GinkgoT())
		opts = &cmd.GlobalOptions{OrgId: orgId, Env: env}
		c = &subject.ListOrganizationMembersCmd{Opts: &subject.ListOptions{RootOptions: opts}, ClientFactory: func(shared.RootOptions) (cmd.Client, error) { return client, nil }}
	})
	Context("output formats", func() {
		var name, email string
		var members []api.OrganizationMember
		BeforeEach(func() {
			name, email = "Example User", "user@example.com"
			members = []api.OrganizationMember{{UserId: 0, OrganizationId: orgId, Role: "admin", Pending: true, CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Name: &name, Email: &email}, {UserId: 1, OrganizationId: orgId, Role: "member"}}
			client.EXPECT().ListOrganizationMembers(orgId).Return(members, nil).Once()
		})
		captureOutput := func() []byte {
			GinkgoHelper()
			original := os.Stdout
			reader, writer, err := os.Pipe()
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(reader.Close)
			DeferCleanup(func() {
				os.Stdout = original
				// Close on early test failure; the successful path already closes the writer.
				_ = writer.Close()
			})
			os.Stdout = writer
			Expect(c.RunE(nil, nil)).To(Succeed())
			Expect(writer.Close()).To(Succeed())
			os.Stdout = original
			output, err := io.ReadAll(reader)
			Expect(err).NotTo(HaveOccurred())
			return output
		}
		It("renders members as a table", func() {
			c.Opts.OutputFormat = shared.OutputFormatTable
			output := captureOutput()
			Expect(string(output)).To(ContainSubstring(name))
			Expect(string(output)).To(ContainSubstring(email))
			Expect(string(output)).To(ContainSubstring("admin"))
			Expect(string(output)).To(ContainSubstring("member"))
		})
		It("renders members as JSON", func() {
			c.Opts.OutputFormat = shared.OutputFormatJSON
			output := captureOutput()
			var decoded []api.OrganizationMember
			Expect(json.Unmarshal(output, &decoded)).To(Succeed())
			Expect(decoded).To(Equal(members))
		})
		It("renders members as YAML", func() {
			c.Opts.OutputFormat = shared.OutputFormatYAML
			output := captureOutput()
			var decoded []api.OrganizationMember
			Expect(yaml.Unmarshal(output, &decoded)).To(Succeed())
			Expect(decoded).To(Equal(members))
		})
	})
	It("handles empty results", func() {
		client.EXPECT().ListOrganizationMembers(orgId).Return([]api.OrganizationMember{}, nil).Once()
		Expect(c.RunE(nil, nil)).To(Succeed())
	})
	It("uses environment organization scope", func() {
		opts.OrgId = ""
		env.EXPECT().GetOrgId().Return(orgId).Once()
		client.EXPECT().ListOrganizationMembers(orgId).Return([]api.OrganizationMember{}, nil).Once()
		Expect(c.RunE(nil, nil)).To(Succeed())
	})
	It("requires organization scope", func() {
		opts.OrgId = ""
		env.EXPECT().GetOrgId().Return("").Once()
		Expect(c.RunE(nil, nil)).To(MatchError(ContainSubstring("organization ID not set")))
	})
	It("rejects invalid organization scope", func() {
		opts.OrgId = "invalid"
		Expect(c.RunE(nil, nil)).To(HaveOccurred())
	})
	It("reports API failures", func() {
		client.EXPECT().ListOrganizationMembers(orgId).Return(nil, errors.New("forbidden")).Once()
		Expect(c.RunE(nil, nil)).To(MatchError("failed to list organization members: forbidden"))
	})
	It("reports client creation failures", func() {
		c.ClientFactory = func(shared.RootOptions) (cmd.Client, error) { return nil, errors.New("no token") }
		Expect(c.RunE(nil, nil)).To(MatchError("failed to create Codesphere client: no token"))
	})
})
