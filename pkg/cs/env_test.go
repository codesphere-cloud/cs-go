// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cs_test

import (
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Environment.GetTeamId", func() {
	BeforeEach(func() {
		GinkgoT().Setenv("CS_TEAM_ID", "")
		GinkgoT().Setenv("CS_RESOURCE_GROUP_ID", "")
	})

	It("reads CS_TEAM_ID when CS_RESOURCE_GROUP_ID is unset", func() {
		GinkgoT().Setenv("CS_TEAM_ID", "11")
		id, err := cs.NewEnv().GetTeamId()
		Expect(err).NotTo(HaveOccurred())
		Expect(id).To(Equal(11))
	})

	It("reads CS_RESOURCE_GROUP_ID when CS_TEAM_ID is unset", func() {
		GinkgoT().Setenv("CS_RESOURCE_GROUP_ID", "22")
		id, err := cs.NewEnv().GetTeamId()
		Expect(err).NotTo(HaveOccurred())
		Expect(id).To(Equal(22))
	})

	It("prefers CS_RESOURCE_GROUP_ID when both are set", func() {
		GinkgoT().Setenv("CS_TEAM_ID", "11")
		GinkgoT().Setenv("CS_RESOURCE_GROUP_ID", "22")
		id, err := cs.NewEnv().GetTeamId()
		Expect(err).NotTo(HaveOccurred())
		Expect(id).To(Equal(22))
	})

	It("fails on a non-numeric CS_RESOURCE_GROUP_ID instead of falling back", func() {
		GinkgoT().Setenv("CS_TEAM_ID", "11")
		GinkgoT().Setenv("CS_RESOURCE_GROUP_ID", "abc")
		_, err := cs.NewEnv().GetTeamId()
		Expect(err).To(MatchError(ContainSubstring("CS_RESOURCE_GROUP_ID")))
	})

	It("returns -1 when neither is set", func() {
		id, err := cs.NewEnv().GetTeamId()
		Expect(err).NotTo(HaveOccurred())
		Expect(id).To(Equal(-1))
	})
})
