// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"github.com/codesphere-cloud/cs-go/cli/cmd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Team aliases for resource groups", func() {
	It("resolves list teams to list resource-groups", func() {
		leaf, rest, err := cmd.GetRootCmd().Find([]string{"list", "teams"})
		Expect(err).NotTo(HaveOccurred())
		Expect(rest).To(BeEmpty())
		Expect(leaf.Name()).To(Equal("resource-groups"))
	})
	It("resolves create team to create resource-group", func() {
		leaf, _, err := cmd.GetRootCmd().Find([]string{"create", "team"})
		Expect(err).NotTo(HaveOccurred())
		Expect(leaf.Name()).To(Equal("resource-group"))
	})
	It("resolves delete team to delete resource-group", func() {
		leaf, _, err := cmd.GetRootCmd().Find([]string{"delete", "team"})
		Expect(err).NotTo(HaveOccurred())
		Expect(leaf.Name()).To(Equal("resource-group"))
	})
	It("resolves add team-member to add resource-group-member", func() {
		leaf, _, err := cmd.GetRootCmd().Find([]string{"add", "team-member"})
		Expect(err).NotTo(HaveOccurred())
		Expect(leaf.Name()).To(Equal("resource-group-member"))
	})
	It("resolves delete team-member to delete resource-group-member", func() {
		leaf, _, err := cmd.GetRootCmd().Find([]string{"delete", "team-member"})
		Expect(err).NotTo(HaveOccurred())
		Expect(leaf.Name()).To(Equal("resource-group-member"))
	})
	It("resolves list team-members to list resource-group-members", func() {
		leaf, _, err := cmd.GetRootCmd().Find([]string{"list", "team-members"})
		Expect(err).NotTo(HaveOccurred())
		Expect(leaf.Name()).To(Equal("resource-group-members"))
	})
	It("sets the resource group flag via --team on child commands", func() {
		leaf, _, err := cmd.GetRootCmd().Find([]string{"list", "resource-group-members"})
		Expect(err).NotTo(HaveOccurred())
		Expect(leaf.ParseFlags([]string{"--team", "42"})).To(Succeed())
		id, err := leaf.Flags().GetInt("resource-group")
		Expect(err).NotTo(HaveOccurred())
		Expect(id).To(Equal(42))
	})
	It("uses the last value when both --resource-group and --team are set", func() {
		root := cmd.GetRootCmd()
		Expect(root.ParseFlags([]string{"--resource-group", "1", "--team", "2"})).To(Succeed())
		id, err := root.PersistentFlags().GetInt("resource-group")
		Expect(err).NotTo(HaveOccurred())
		Expect(id).To(Equal(2))
	})
	It("keeps -t as the shorthand", func() {
		root := cmd.GetRootCmd()
		Expect(root.ParseFlags([]string{"-t", "7"})).To(Succeed())
		id, err := root.PersistentFlags().GetInt("resource-group")
		Expect(err).NotTo(HaveOccurred())
		Expect(id).To(Equal(7))
	})
	It("shows --resource-group in the global flag help", func() {
		flag := cmd.GetRootCmd().PersistentFlags().Lookup("team")
		Expect(flag).NotTo(BeNil())
		Expect(flag.Name).To(Equal("resource-group"))
		Expect(flag.Shorthand).To(Equal("t"))
	})
	It("keeps the deprecated team create path", func() {
		leaf, rest, err := cmd.GetRootCmd().Find([]string{"team", "create"})
		Expect(err).NotTo(HaveOccurred())
		Expect(rest).To(BeEmpty())
		Expect(leaf.Name()).To(Equal("create"))
		Expect(leaf.Deprecated).NotTo(BeEmpty())
	})
	It("does not add aliases to deprecated legacy commands", func() {
		leaf, rest, err := cmd.GetRootCmd().Find([]string{"team", "team"})
		Expect(err).NotTo(HaveOccurred())
		Expect(leaf.Name()).To(Equal("team"))
		Expect(rest).To(Equal([]string{"team"}))
	})
})
