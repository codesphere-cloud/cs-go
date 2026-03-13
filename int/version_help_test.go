// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package int_test

import (
	"fmt"
	"log"
	"strings"

	intutil "github.com/codesphere-cloud/cs-go/int/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version and Help Tests", Label("local"), func() {
	Context("Version Command", func() {
		It("should display version information", func() {
			By("Running version command")
			output := intutil.RunCommand("version")
			log.Printf("Version output: %s\n", output)

			Expect(output).To(Or(
				ContainSubstring("version"),
				ContainSubstring("Version"),
				MatchRegexp(`\d+\.\d+\.\d+`),
			))
		})
	})

	Context("Help Commands", func() {
		It("should display main help", func() {
			By("Running help command")
			output := intutil.RunCommand("--help")
			log.Printf("Help output length: %d\n", len(output))

			Expect(output).To(ContainSubstring("Usage:"))
			Expect(output).To(ContainSubstring("Available Commands:"))
		})

		It("should display help for all subcommands", func() {
			testCases := []struct {
				command     []string
				shouldMatch string
			}{
				{[]string{"create", "--help"}, "workspace"},
				{[]string{"exec", "--help"}, "exec"},
				{[]string{"log", "--help"}, "log"},
				{[]string{"start", "pipeline", "--help"}, "pipeline"},
				{[]string{"git", "pull", "--help"}, "pull"},
				{[]string{"set-env", "--help"}, "set-env"},
			}

			for _, tc := range testCases {
				By(fmt.Sprintf("Testing %v", tc.command))
				output := intutil.RunCommand(tc.command...)
				Expect(output).To(ContainSubstring("Usage:"))
				Expect(output).To(ContainSubstring(tc.shouldMatch))
			}
		})
	})

	Context("Invalid Commands", func() {
		It("should handle unknown commands gracefully", func() {
			By("Running unknown command")
			output, exitCode := intutil.RunCommandWithExitCode("unknowncommand")
			log.Printf("Unknown command output: %s (exit code: %d)\n", output, exitCode)

			Expect(exitCode).NotTo(Equal(0))
			Expect(output).To(Or(
				ContainSubstring("unknown command"),
				ContainSubstring("Error:"),
			))
		})

		It("should suggest similar commands for typos", func() {
			By("Running misspelled command")
			output, exitCode := intutil.RunCommandWithExitCode("listt")
			log.Printf("Typo command output: %s (exit code: %d)\n", output, exitCode)

			Expect(exitCode).NotTo(Equal(0))
			lowerOutput := strings.ToLower(output)
			Expect(lowerOutput).To(Or(
				ContainSubstring("unknown"),
				ContainSubstring("error"),
				ContainSubstring("did you mean"),
			))
		})
	})

	Context("Global Flags", func() {
		It("should accept all global flags", func() {
			By("Testing --api flag")
			output := intutil.RunCommand(
				"--api", "https://example.com/api",
				"list", "teams",
			)
			Expect(output).NotTo(ContainSubstring("unknown flag"))

			By("Testing --verbose flag")
			output = intutil.RunCommand(
				"--verbose",
				"list", "plans",
			)
			Expect(output).NotTo(ContainSubstring("unknown flag"))

			By("Testing -v shorthand")
			output = intutil.RunCommand(
				"-v",
				"list", "baseimages",
			)
			Expect(output).NotTo(ContainSubstring("unknown flag"))
		})
	})
})
