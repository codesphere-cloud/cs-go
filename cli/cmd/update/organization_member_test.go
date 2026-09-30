// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package update_test

import (
	"errors"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
	"github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	subject "github.com/codesphere-cloud/cs-go/cli/cmd/update"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ChangeOrganizationMemberRole", func() {
	const orgId = "8316ee2f-87f8-424e-b925-7382dc50d662"
	var client *cmd.MockClient
	var env *cmd.MockEnv
	var opts *cmd.GlobalOptions
	var c *subject.UpdateOrganizationMemberCmd
	BeforeEach(func() {
		client = cmd.NewMockClient(GinkgoT())
		env = cmd.NewMockEnv(GinkgoT())
		opts = &cmd.GlobalOptions{OrgId: orgId, Env: env}
		c = &subject.UpdateOrganizationMemberCmd{Opts: subject.UpdateOrganizationMemberOpts{RootOptions: opts, UserId: 0, Role: "admin"}, ClientFactory: func(shared.RootOptions) (cmd.Client, error) { return client, nil }}
	})
	It("calls the API with the organization and member parameters", func() {
		client.EXPECT().ChangeOrganizationMemberRole(orgId, 0, "admin").Return(nil).Once()
		Expect(c.RunE(nil, nil)).To(Succeed())
	})
	It("resolves organization from the environment", func() {
		opts.OrgId = ""
		env.EXPECT().GetOrgId().Return(orgId).Once()
		client.EXPECT().ChangeOrganizationMemberRole(orgId, 0, "admin").Return(nil).Once()
		Expect(c.RunE(nil, nil)).To(Succeed())
	})
	It("requires an organization", func() {
		opts.OrgId = ""
		env.EXPECT().GetOrgId().Return("").Once()
		Expect(c.RunE(nil, nil)).To(MatchError(ContainSubstring("organization ID not set")))
	})
	It("rejects malformed organization IDs", func() {
		opts.OrgId = "invalid"
		Expect(c.RunE(nil, nil)).To(MatchError(ContainSubstring("invalid organization ID")))
	})
	It("rejects invalid member input without an API call", func() {
		c.Opts.UserId = -1
		Expect(c.RunE(nil, nil)).To(HaveOccurred())
	})
	It("rejects invalid roles", func() {
		c.Opts.Role = "owner"
		Expect(c.RunE(nil, nil)).To(MatchError(`invalid organization role "owner": must be admin or member`))
	})
	It("preserves API failures", func() {
		failure := errors.New("forbidden")
		client.EXPECT().ChangeOrganizationMemberRole(orgId, 0, "admin").Return(failure).Once()
		Expect(c.RunE(nil, nil)).To(MatchError(ContainSubstring("forbidden")))
	})
	It("reports client creation failures", func() {
		c.ClientFactory = func(shared.RootOptions) (cmd.Client, error) { return nil, errors.New("no token") }
		Expect(c.RunE(nil, nil)).To(MatchError("failed to create Codesphere client: no token"))
	})
})
