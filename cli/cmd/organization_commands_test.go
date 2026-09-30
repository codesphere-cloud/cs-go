// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"github.com/codesphere-cloud/cs-go/cli/cmd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Organization command registration", func() {
	It("registers list members", func() {
		leaf, rest, err := cmd.GetRootCmd().Find([]string{"list", "organization-members"})
		Expect(err).NotTo(HaveOccurred())
		Expect(rest).To(BeEmpty())
		Expect(leaf.RunE).NotTo(BeNil())
		Expect(leaf.Short).NotTo(BeEmpty())
		Expect(leaf.Example).NotTo(BeEmpty())
	})
	It("registers add member", func() {
		leaf, rest, err := cmd.GetRootCmd().Find([]string{"add", "organization-member"})
		Expect(err).NotTo(HaveOccurred())
		Expect(rest).To(BeEmpty())
		Expect(leaf.RunE).NotTo(BeNil())
		Expect(leaf.Short).NotTo(BeEmpty())
		Expect(leaf.Example).NotTo(BeEmpty())
	})
	It("registers remove member", func() {
		leaf, rest, err := cmd.GetRootCmd().Find([]string{"delete", "organization-member"})
		Expect(err).NotTo(HaveOccurred())
		Expect(rest).To(BeEmpty())
		Expect(leaf.RunE).NotTo(BeNil())
		Expect(leaf.Short).NotTo(BeEmpty())
		Expect(leaf.Example).NotTo(BeEmpty())
	})
	It("registers update role", func() {
		leaf, rest, err := cmd.GetRootCmd().Find([]string{"update", "org-member"})
		Expect(err).NotTo(HaveOccurred())
		Expect(rest).To(BeEmpty())
		Expect(leaf.RunE).NotTo(BeNil())
		Expect(leaf.Short).NotTo(BeEmpty())
		Expect(leaf.Example).NotTo(BeEmpty())
	})
	It("registers update role alias", func() {
		leaf, rest, err := cmd.GetRootCmd().Find([]string{"update", "organization-member"})
		Expect(err).NotTo(HaveOccurred())
		Expect(rest).To(BeEmpty())
		Expect(leaf.RunE).NotTo(BeNil())
		Expect(leaf.Short).NotTo(BeEmpty())
		Expect(leaf.Example).NotTo(BeEmpty())
	})
	It("registers plural organizations", func() {
		leaf, rest, err := cmd.GetRootCmd().Find([]string{"list", "organizations"})
		Expect(err).NotTo(HaveOccurred())
		Expect(rest).To(BeEmpty())
		Expect(leaf.RunE).NotTo(BeNil())
		Expect(leaf.Short).NotTo(BeEmpty())
		Expect(leaf.Example).NotTo(BeEmpty())
	})
	It("registers singular organization", func() {
		leaf, rest, err := cmd.GetRootCmd().Find([]string{"list", "organization"})
		Expect(err).NotTo(HaveOccurred())
		Expect(rest).To(BeEmpty())
		Expect(leaf.RunE).NotTo(BeNil())
		Expect(leaf.Short).NotTo(BeEmpty())
		Expect(leaf.Example).NotTo(BeEmpty())
	})
	It("registers short organization", func() {
		leaf, rest, err := cmd.GetRootCmd().Find([]string{"list", "org"})
		Expect(err).NotTo(HaveOccurred())
		Expect(rest).To(BeEmpty())
		Expect(leaf.RunE).NotTo(BeNil())
		Expect(leaf.Short).NotTo(BeEmpty())
		Expect(leaf.Example).NotTo(BeEmpty())
	})
	It("registers short organizations", func() {
		leaf, rest, err := cmd.GetRootCmd().Find([]string{"list", "orgs"})
		Expect(err).NotTo(HaveOccurred())
		Expect(rest).To(BeEmpty())
		Expect(leaf.RunE).NotTo(BeNil())
		Expect(leaf.Short).NotTo(BeEmpty())
		Expect(leaf.Example).NotTo(BeEmpty())
	})
	It("registers short organization members", func() {
		leaf, rest, err := cmd.GetRootCmd().Find([]string{"list", "org-members"})
		Expect(err).NotTo(HaveOccurred())
		Expect(rest).To(BeEmpty())
		Expect(leaf.RunE).NotTo(BeNil())
		Expect(leaf.Short).NotTo(BeEmpty())
		Expect(leaf.Example).NotTo(BeEmpty())
	})
	It("requires the email flag before running", func() {
		root := cmd.GetRootCmd()
		root.SetArgs([]string{"add", "organization-member"})
		Expect(root.Execute()).To(MatchError(ContainSubstring(`required flag(s) "email"`)))
	})
	It("requires the user flag before running", func() {
		root := cmd.GetRootCmd()
		root.SetArgs([]string{"delete", "organization-member"})
		Expect(root.Execute()).To(MatchError(ContainSubstring(`required flag(s) "user"`)))
	})
	It("requires the role flag before running", func() {
		root := cmd.GetRootCmd()
		root.SetArgs([]string{"update", "org-member", "--user", "0"})
		Expect(root.Execute()).To(MatchError(ContainSubstring(`required flag(s) "role"`)))
	})
	It("rejects unexpected arguments", func() {
		root := cmd.GetRootCmd()
		root.SetArgs([]string{"list", "organization-members", "unexpected"})
		Expect(root.Execute()).To(HaveOccurred())
	})
	It("retains CLI self-update", func() {
		leaf, _, err := cmd.GetRootCmd().Find([]string{"update"})
		Expect(err).NotTo(HaveOccurred())
		Expect(leaf.Short).To(Equal("Update Codesphere CLI"))
		Expect(leaf.RunE).NotTo(BeNil())
	})
	It("uses -g for organization and -o for output", func() {
		root := cmd.GetRootCmd()
		leaf, _, err := root.Find([]string{"list", "organization-members"})
		Expect(err).NotTo(HaveOccurred())
		Expect(leaf.ParseFlags([]string{"-g", "8316ee2f-87f8-424e-b925-7382dc50d662", "-o", "json"})).To(Succeed())
		orgID, err := leaf.Flags().GetString("org")
		Expect(err).NotTo(HaveOccurred())
		Expect(orgID).To(Equal("8316ee2f-87f8-424e-b925-7382dc50d662"))
		output, err := leaf.Flags().GetString("output")
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(Equal("json"))
	})
	It("rejects fractional user IDs before running", func() {
		root := cmd.GetRootCmd()
		root.SetArgs([]string{"delete", "organization-member", "--user", "1.5"})
		Expect(root.Execute()).To(MatchError(ContainSubstring("invalid syntax")))
	})

	It("accepts the legacy -O shorthand on child commands", func() {
		root := cmd.GetRootCmd()
		leaf, _, err := root.Find([]string{"list", "organization-members"})
		Expect(err).NotTo(HaveOccurred())
		Expect(leaf.ParseFlags([]string{"-O", "8316ee2f-87f8-424e-b925-7382dc50d662"})).To(Succeed())
		orgID, err := leaf.Flags().GetString("org")
		Expect(err).NotTo(HaveOccurred())
		Expect(orgID).To(Equal("8316ee2f-87f8-424e-b925-7382dc50d662"))
	})
	It("uses the last organization flag across all aliases", func() {
		root := cmd.GetRootCmd()
		Expect(root.ParseFlags([]string{"-O", "legacy", "--org", "long", "-g", "preferred"})).To(Succeed())
		orgID, err := root.PersistentFlags().GetString("org")
		Expect(err).NotTo(HaveOccurred())
		Expect(orgID).To(Equal("preferred"))
		Expect(root.ParseFlags([]string{"-g", "preferred", "-O", "legacy-last"})).To(Succeed())
		orgID, err = root.PersistentFlags().GetString("org")
		Expect(err).NotTo(HaveOccurred())
		Expect(orgID).To(Equal("legacy-last"))
	})

})
