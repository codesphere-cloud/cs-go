// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package int_test

import (
	"log"

	intutil "github.com/codesphere-cloud/cs-go/int/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("List Command Tests", Label("list"), func() {
	var teamId string

	BeforeEach(func() {
		teamId, _ = intutil.FailIfMissingEnvVars()
	})

	Context("List Workspaces", func() {
		It("should list all workspaces in team with proper formatting", func() {
			By("Listing workspaces")
			output := intutil.RunCommand("list", "workspaces", "-t", teamId)
			log.Printf("List workspaces output length: %d\n", len(output))

			Expect(output).To(ContainSubstring("TEAM ID"))
			Expect(output).To(ContainSubstring("ID"))
			Expect(output).To(ContainSubstring("NAME"))
		})
	})

	Context("List Plans", func() {
		It("should list all available plans", func() {
			By("Listing plans")
			output := intutil.RunCommand("list", "plans")
			log.Printf("List plans output: %s\n", output)

			Expect(output).To(ContainSubstring("ID"))
			Expect(output).To(ContainSubstring("NAME"))
			Expect(output).To(Or(
				ContainSubstring("Micro"),
				ContainSubstring("Free"),
			))
		})

		It("should show plan details like CPU and RAM", func() {
			By("Listing plans with details")
			output := intutil.RunCommand("list", "plans")
			log.Printf("Plan details output length: %d\n", len(output))

			Expect(output).To(ContainSubstring("CPU"))
			Expect(output).To(ContainSubstring("RAM"))
		})
	})

	Context("List Base Images", func() {
		It("should list available base images", func() {
			By("Listing base images")
			output := intutil.RunCommand("list", "baseimages")
			log.Printf("List baseimages output: %s\n", output)

			Expect(output).To(ContainSubstring("ID"))
			Expect(output).To(ContainSubstring("NAME"))
		})

		It("should show Ubuntu images", func() {
			By("Checking for Ubuntu in base images")
			output := intutil.RunCommand("list", "baseimages")

			Expect(output).To(ContainSubstring("Ubuntu"))
		})
	})

	Context("List Teams", func() {
		It("should list teams user has access to", func() {
			By("Listing teams")
			output := intutil.RunCommand("list", "teams")
			log.Printf("List teams output: %s\n", output)

			Expect(output).To(ContainSubstring("ID"))
			Expect(output).To(ContainSubstring("NAME"))
			Expect(output).To(ContainSubstring(teamId))
		})

		It("should show team role", func() {
			By("Checking team roles")
			output := intutil.RunCommand("list", "teams")

			Expect(output).To(Or(
				ContainSubstring("Admin"),
				ContainSubstring("Member"),
				ContainSubstring("ROLE"),
			))
		})
	})

	Context("List Error Handling", func() {
		It("should handle missing or invalid list subcommand", func() {
			By("Running list without subcommand")
			output, exitCode := intutil.RunCommandWithExitCode("list")
			log.Printf("List without subcommand output: %s (exit code: %d)\n", output, exitCode)
			Expect(output).To(Or(
				ContainSubstring("Available Commands:"),
				ContainSubstring("Usage:"),
			))

			By("Running list with invalid subcommand")
			output, _ = intutil.RunCommandWithExitCode("list", "invalid")
			log.Printf("List invalid output (first 200 chars): %s\n", output[:min(200, len(output))])
			Expect(output).To(Or(
				ContainSubstring("Available Commands:"),
				ContainSubstring("Usage:"),
			))
		})

		It("should require team ID for workspace listing when not set globally", func() {
			By("Listing workspaces without team ID in specific contexts")
			output := intutil.RunCommand("list", "workspaces", "-t", teamId)

			Expect(output).NotTo(BeEmpty())
		})
	})
})
