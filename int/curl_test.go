// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package int_test

import (
	"fmt"
	"log"
	"os"
	"time"

	intutil "github.com/codesphere-cloud/cs-go/int/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Curl Workspace Integration Tests", Label("curl"), func() {
	var (
		teamId        string
		workspaceName string
		workspaceId   string
	)

	BeforeEach(func() {
		teamId, _ = intutil.FailIfMissingEnvVars()
		workspaceName = intutil.NewWorkspaceName("curl")
	})

	AfterEach(func() {
		if workspaceId != "" {
			By(fmt.Sprintf("Cleaning up: deleting workspace %s (ID: %s)", workspaceName, workspaceId))
			intutil.CleanupWorkspace(workspaceId)
			workspaceId = ""
		}
	})

	Context("Curl Command", func() {
		BeforeEach(func() {
			By("Creating a workspace for curl testing")
			workspaceId = intutil.CreateTestWorkspace(teamId, workspaceName)

			By("Waiting for workspace to be fully provisioned")
			time.Sleep(intutil.PostCreateWaitTime)
		})

		It("should send authenticated request to workspace", func() {
			By("Sending curl request to workspace root")
			output := intutil.RunCommand(
				"curl", "/",
				"-w", workspaceId,
				"--", "-k", "-s", "-o", "/dev/null", "-w", "%{http_code}",
			)
			log.Printf("Curl workspace output: %s\n", output)

			Expect(output).To(ContainSubstring("Sending request to workspace"))
			Expect(output).To(ContainSubstring(workspaceId))
		})

		It("should support custom paths", func() {
			By("Sending curl request to custom path")
			output, exitCode := intutil.RunCommandWithExitCode(
				"curl", "/api/health",
				"-w", workspaceId,
				"--", "-k", "-s", "-o", "/dev/null", "-w", "%{http_code}",
			)
			log.Printf("Curl with custom path output: %s (exit code: %d)\n", output, exitCode)

			Expect(output).To(ContainSubstring("Sending request to workspace"))
		})

		It("should pass through curl arguments", func() {
			By("Sending HEAD request using curl -I flag")
			output := intutil.RunCommand(
				"curl", "/",
				"-w", workspaceId,
				"--", "-k", "-I",
			)
			log.Printf("Curl with -I flag output: %s\n", output)

			Expect(output).To(ContainSubstring("Sending request to workspace"))
		})

		It("should work with workspace ID from environment variable", func() {
			By("Setting CS_WORKSPACE_ID environment variable")
			originalWsId := os.Getenv("CS_WORKSPACE_ID")
			_ = os.Setenv("CS_WORKSPACE_ID", workspaceId)
			defer func() { _ = os.Setenv("CS_WORKSPACE_ID", originalWsId) }()

			By("Sending curl request using environment variable")
			output := intutil.RunCommand(
				"curl", "/",
				"--", "-k", "-s", "-o", "/dev/null", "-w", "%{http_code}",
			)
			log.Printf("Curl with env var output: %s\n", output)

			Expect(output).To(ContainSubstring("Sending request to workspace"))
			Expect(output).To(ContainSubstring(workspaceId))
		})
	})

	Context("Curl Error Handling", func() {
		It("should fail when workspace ID is missing", func() {
			By("Attempting to curl without workspace ID")
			intutil.WithClearedWorkspaceEnv(func() {
				output, exitCode := intutil.RunCommandWithExitCode("curl", "/")
				log.Printf("Curl without workspace ID output: %s (exit code: %d)\n", output, exitCode)

				Expect(exitCode).NotTo(Equal(0))
				Expect(output).To(Or(
					ContainSubstring("workspace"),
					ContainSubstring("required"),
					ContainSubstring("not set"),
				))
			})
		})

		It("should require path argument", func() {
			By("Attempting to curl without path")
			output, exitCode := intutil.RunCommandWithExitCode(
				"curl",
				"-w", "1234",
			)
			log.Printf("Curl without path output: %s (exit code: %d)\n", output, exitCode)

			Expect(exitCode).NotTo(Equal(0))
			Expect(output).To(Or(
				ContainSubstring("path"),
				ContainSubstring("required"),
				ContainSubstring("argument"),
			))
		})
	})

	Context("Curl Command Help", func() {
		It("should display help information", func() {
			By("Running curl --help")
			output := intutil.RunCommand("curl", "--help")
			log.Printf("Curl help output: %s\n", output)

			Expect(output).To(ContainSubstring("Send authenticated HTTP requests"))
			Expect(output).To(ContainSubstring("--timeout"))
			Expect(output).To(ContainSubstring("-w, --workspace"))
		})
	})
})
