// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"github.com/codesphere-cloud/cs-go/cli/cmd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resource group aliases", func() {
	It("resolves list resource-groups to list teams", func() {
		leaf, rest, err := cmd.GetRootCmd().Find([]string{"list", "resource-groups"})
		Expect(err).NotTo(HaveOccurred())
		Expect(rest).To(BeEmpty())
		Expect(leaf.Name()).To(Equal("teams"))
	})
	It("resolves create resource-group to create team", func() {
		leaf, _, err := cmd.GetRootCmd().Find([]string{"create", "resource-group"})
		Expect(err).NotTo(HaveOccurred())
		Expect(leaf.Name()).To(Equal("team"))
	})
	It("resolves delete resource-group to delete team", func() {
		leaf, _, err := cmd.GetRootCmd().Find([]string{"delete", "resource-group"})
		Expect(err).NotTo(HaveOccurred())
		Expect(leaf.Name()).To(Equal("team"))
	})
	It("resolves add resource-group-member to add team-member", func() {
		leaf, _, err := cmd.GetRootCmd().Find([]string{"add", "resource-group-member"})
		Expect(err).NotTo(HaveOccurred())
		Expect(leaf.Name()).To(Equal("team-member"))
	})
	It("resolves delete resource-group-member to delete team-member", func() {
		leaf, _, err := cmd.GetRootCmd().Find([]string{"delete", "resource-group-member"})
		Expect(err).NotTo(HaveOccurred())
		Expect(leaf.Name()).To(Equal("team-member"))
	})
	It("resolves list resource-group-members to list team-members", func() {
		leaf, _, err := cmd.GetRootCmd().Find([]string{"list", "resource-group-members"})
		Expect(err).NotTo(HaveOccurred())
		Expect(leaf.Name()).To(Equal("team-members"))
	})
	It("sets the team flag via --resource-group on child commands", func() {
		leaf, _, err := cmd.GetRootCmd().Find([]string{"list", "team-members"})
		Expect(err).NotTo(HaveOccurred())
		Expect(leaf.ParseFlags([]string{"--resource-group", "42"})).To(Succeed())
		teamID, err := leaf.Flags().GetInt("team")
		Expect(err).NotTo(HaveOccurred())
		Expect(teamID).To(Equal(42))
	})
	It("uses the last value when both --team and --resource-group are set", func() {
		root := cmd.GetRootCmd()
		Expect(root.ParseFlags([]string{"--team", "1", "--resource-group", "2"})).To(Succeed())
		teamID, err := root.PersistentFlags().GetInt("team")
		Expect(err).NotTo(HaveOccurred())
		Expect(teamID).To(Equal(2))
	})
	It("keeps -t working", func() {
		root := cmd.GetRootCmd()
		Expect(root.ParseFlags([]string{"-t", "7"})).To(Succeed())
		teamID, err := root.PersistentFlags().GetInt("resource-group")
		Expect(err).NotTo(HaveOccurred())
		Expect(teamID).To(Equal(7))
	})
	It("does not add resource group aliases to deprecated legacy commands", func() {
		leaf, rest, err := cmd.GetRootCmd().Find([]string{"team", "resource-group"})
		Expect(err).NotTo(HaveOccurred())
		Expect(leaf.Name()).To(Equal("team"))
		Expect(rest).To(Equal([]string{"resource-group"}))
	})
})
