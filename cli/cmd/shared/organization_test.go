// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package shared_test

import (
	"github.com/codesphere-cloud/cs-go/cli/cmd"
	"github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RequireOrganizationID", func() {
	const orgID = "8316ee2f-87f8-424e-b925-7382dc50d662"
	It("uses the explicit organization ID", func() {
		env := cmd.NewMockEnv(GinkgoT())
		id, err := shared.RequireOrganizationID(&cmd.GlobalOptions{OrgId: orgID, Env: env})
		Expect(err).NotTo(HaveOccurred())
		Expect(id).To(Equal(orgID))
	})
	It("falls back to the environment", func() {
		env := cmd.NewMockEnv(GinkgoT())
		env.EXPECT().GetOrgId().Return(orgID).Once()
		id, err := shared.RequireOrganizationID(&cmd.GlobalOptions{Env: env})
		Expect(err).NotTo(HaveOccurred())
		Expect(id).To(Equal(orgID))
	})
	It("explains how to supply a missing organization ID", func() {
		env := cmd.NewMockEnv(GinkgoT())
		env.EXPECT().GetOrgId().Return("").Once()
		id, err := shared.RequireOrganizationID(&cmd.GlobalOptions{Env: env})
		Expect(err).To(MatchError("organization ID not set, use --org or CS_ORG_ID to set it"))
		Expect(id).To(BeEmpty())
	})
	It("reports an invalid organization ID", func() {
		id, err := shared.RequireOrganizationID(&cmd.GlobalOptions{OrgId: "invalid"})
		Expect(err).To(MatchError(ContainSubstring("failed to get organization ID: invalid organization ID format")))
		Expect(id).To(BeEmpty())
	})
})

var _ = Describe("ValidateOrganizationRole", func() {
	It("accepts admin", func() {
		Expect(shared.ValidateOrganizationRole("admin")).To(Succeed())
	})
	It("accepts member", func() {
		Expect(shared.ValidateOrganizationRole("member")).To(Succeed())
	})
	It("includes the invalid role in the error", func() {
		Expect(shared.ValidateOrganizationRole("owner")).To(MatchError(`invalid organization role "owner": must be admin or member`))
	})
	It("identifies an empty role", func() {
		Expect(shared.ValidateOrganizationRole("")).To(MatchError(`invalid organization role "": must be admin or member`))
	})
})
