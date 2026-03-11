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

var _ = Describe("Wake Up Workspace Integration Tests", Label("wakeup"), func() {
	var (
		teamId        string
		workspaceName string
		workspaceId   string
	)

	BeforeEach(func() {
		teamId, _ = intutil.FailIfMissingEnvVars()
		workspaceName = intutil.NewWorkspaceName("wakeup")
	})

	AfterEach(func() {
		if workspaceId != "" {
			By(fmt.Sprintf("Cleaning up: deleting workspace %s (ID: %s)", workspaceName, workspaceId))
			intutil.CleanupWorkspace(workspaceId)
			workspaceId = ""
		}
	})

	Context("Wake Up Command", func() {
		BeforeEach(func() {
			By("Creating a workspace for wake-up testing")
			workspaceId = intutil.CreateTestWorkspace(teamId, workspaceName)

			By("Waiting for workspace to be fully provisioned")
			time.Sleep(intutil.PostCreateWaitTime)
		})

		It("should wake up workspace successfully", func() {
			By("Waking up the workspace")
			output := intutil.RunCommand(
				"wake-up",
				"-w", workspaceId,
			)
			log.Printf("Wake up workspace output: %s\n", output)

			Expect(output).To(Or(
				ContainSubstring("Waking up workspace"),
				// The workspace might already be running
				ContainSubstring("is already running"),
			))
			Expect(output).To(ContainSubstring(workspaceId))
		})

		It("should respect custom timeout", func() {
			By("Waking up workspace with custom timeout")
			output, exitCode := intutil.RunCommandWithExitCode(
				"wake-up",
				"-w", workspaceId,
				"--timeout", "5s",
			)
			log.Printf("Wake up with timeout output: %s (exit code: %d)\n", output, exitCode)

			Expect(output).To(Or(
				ContainSubstring("Waking up workspace"),
				// The workspace might already be running
				ContainSubstring("is already running"),
			))
			Expect(exitCode).To(Equal(0))
		})

		It("should work with workspace ID from environment variable", func() {
			By("Setting CS_WORKSPACE_ID environment variable")
			originalWsId := os.Getenv("CS_WORKSPACE_ID")
			_ = os.Setenv("CS_WORKSPACE_ID", workspaceId)
			defer func() { _ = os.Setenv("CS_WORKSPACE_ID", originalWsId) }()

			By("Waking up workspace using environment variable")
			output := intutil.RunCommand("wake-up")
			log.Printf("Wake up with env var output: %s\n", output)

			Expect(output).To(Or(
				ContainSubstring("Waking up workspace"),
				// The workspace might already be running
				ContainSubstring("is already running"),
			))
			Expect(output).To(ContainSubstring(workspaceId))
		})
	})

	Context("Wake Up Error Handling", func() {
		It("should fail when workspace ID is missing", func() {
			By("Attempting to wake up workspace without ID")
			intutil.WithClearedWorkspaceEnv(func() {
				output, exitCode := intutil.RunCommandWithExitCode("wake-up")
				log.Printf("Wake up without workspace ID output: %s (exit code: %d)\n", output, exitCode)

				Expect(exitCode).NotTo(Equal(0))
				Expect(output).To(Or(
					ContainSubstring("workspace"),
					ContainSubstring("required"),
					ContainSubstring("not set"),
				))
			})
		})

		It("should handle workspace without dev domain gracefully", func() {
			By("Creating a workspace (which might not have dev domain configured)")
			workspaceId = intutil.CreateTestWorkspace(teamId, workspaceName)

			By("Attempting to wake up the workspace")
			wakeupOutput, wakeupExitCode := intutil.RunCommandWithExitCode(
				"wake-up",
				"-w", workspaceId,
			)
			log.Printf("Wake up workspace output: %s (exit code: %d)\n", wakeupOutput, wakeupExitCode)

			if wakeupExitCode != 0 {
				Expect(wakeupOutput).To(Or(
					ContainSubstring("development domain"),
					ContainSubstring("dev domain"),
					ContainSubstring("failed to wake up"),
				))
			}
		})
	})

	Context("Wake Up Command Help", func() {
		It("should display help information", func() {
			By("Running wake-up --help")
			output := intutil.RunCommand("wake-up", "--help")
			log.Printf("Wake up help output: %s\n", output)

			Expect(output).To(ContainSubstring("Wake up an on-demand workspace"))
			Expect(output).To(ContainSubstring("--timeout"))
			Expect(output).To(ContainSubstring("-w, --workspace"))
		})
	})
})
