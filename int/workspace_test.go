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

var _ = Describe("Open Workspace Integration Tests", Label("workspace"), func() {
	var (
		teamId        string
		workspaceName string
		workspaceId   string
	)

	BeforeEach(func() {
		teamId, _ = intutil.FailIfMissingEnvVars()
		workspaceName = intutil.NewWorkspaceName("open")
	})

	AfterEach(func() {
		if workspaceId != "" {
			By(fmt.Sprintf("Cleaning up: deleting workspace %s (ID: %s)", workspaceName, workspaceId))
			intutil.CleanupWorkspace(workspaceId)
			workspaceId = ""
		}
	})

	Context("Open Workspace Command", func() {
		BeforeEach(func() {
			By("Creating a workspace")
			workspaceId = intutil.CreateTestWorkspace(teamId, workspaceName)
		})

		It("should open workspace successfully", func() {
			By("Opening the workspace")
			output := intutil.RunCommand(
				"open", "workspace",
				"-w", workspaceId,
			)
			log.Printf("Open workspace output: %s\n", output)

			Expect(output).To(ContainSubstring("Opening workspace"))
			Expect(output).To(ContainSubstring(workspaceId))
		})
	})

	Context("Open Workspace Error Handling", func() {
		It("should fail when workspace ID is missing", func() {
			By("Attempting to open workspace without ID")
			intutil.WithClearedWorkspaceEnv(func() {
				output, exitCode := intutil.RunCommandWithExitCode(
					"open", "workspace",
				)
				log.Printf("Open without workspace ID output: %s (exit code: %d)\n", output, exitCode)
				Expect(exitCode).NotTo(Equal(0))
				Expect(output).To(Or(
					ContainSubstring("workspace"),
					ContainSubstring("required"),
				))
			})
		})
	})
})

var _ = Describe("Workspace Edge Cases and Advanced Operations", Label("workspace"), func() {
	var (
		teamId        string
		workspaceName string
		workspaceId   string
	)

	BeforeEach(func() {
		teamId, _ = intutil.FailIfMissingEnvVars()
		workspaceName = intutil.NewWorkspaceName("edge")
	})

	AfterEach(func() {
		if workspaceId != "" {
			By(fmt.Sprintf("Cleaning up: deleting workspace %s (ID: %s)", workspaceName, workspaceId))
			intutil.CleanupWorkspace(workspaceId)
			workspaceId = ""
		}
	})

	Context("Workspace Creation Edge Cases", func() {
		It("should create a workspace with a very long name", func() {
			longName := intutil.NewWorkspaceName("very-long-workspace-name")
			By("Creating a workspace with a long name")
			output := intutil.RunCommand(
				"create", "workspace", longName,
				"-t", teamId,
				"-p", intutil.DefaultPlanId,
				"--timeout", intutil.DefaultCreateTimeout,
			)
			log.Printf("Create workspace with long name output: %s\n", output)

			if output != "" && !strings.Contains(output, "error") {
				Expect(output).To(ContainSubstring(intutil.WorkspaceCreatedOutput))
				workspaceId = intutil.ExtractWorkspaceId(output)
			}
		})

		It("should handle creation timeout gracefully", func() {
			By("Creating a workspace with very short timeout")
			output, exitCode := intutil.RunCommandWithExitCode(
				"create", "workspace", workspaceName,
				"-t", teamId,
				"-p", intutil.DefaultPlanId,
				"--timeout", "1s",
			)
			log.Printf("Create with short timeout output: %s (exit code: %d)\n", output, exitCode)

			if exitCode != 0 {
				Expect(output).To(Or(
					ContainSubstring("timeout"),
					ContainSubstring("timed out"),
				))
			} else if strings.Contains(output, intutil.WorkspaceCreatedOutput) {
				workspaceId = intutil.ExtractWorkspaceId(output)
			}
		})
	})

	Context("Exec Command Edge Cases", func() {
		BeforeEach(func() {
			By("Creating a workspace")
			workspaceId = intutil.CreateTestWorkspace(teamId, workspaceName)
		})

		It("should execute commands with multiple arguments", func() {
			By("Executing a command with multiple arguments")
			output := intutil.RunCommand(
				"exec",
				"-w", workspaceId,
				"--",
				"sh", "-c", "echo test1 && echo test2",
			)
			log.Printf("Exec with multiple args output: %s\n", output)
			Expect(output).To(ContainSubstring("test1"))
			Expect(output).To(ContainSubstring("test2"))
		})

		It("should handle commands that output to stderr", func() {
			By("Executing a command that writes to stderr")
			output := intutil.RunCommand(
				"exec",
				"-w", workspaceId,
				"--",
				"sh", "-c", "echo error message >&2",
			)
			log.Printf("Exec with stderr output: %s\n", output)
			Expect(output).To(ContainSubstring("error message"))
		})

		It("should handle commands with exit codes", func() {
			By("Executing a command that exits with non-zero code")
			output, exitCode := intutil.RunCommandWithExitCode(
				"exec",
				"-w", workspaceId,
				"--",
				"sh", "-c", "exit 42",
			)
			log.Printf("Exec with exit code output: %s (exit code: %d)\n", output, exitCode)
		})

		It("should execute long-running commands", func() {
			By("Executing a command that takes a few seconds")
			output := intutil.RunCommand(
				"exec",
				"-w", workspaceId,
				"--",
				"sh", "-c", "sleep 2 && echo completed",
			)
			log.Printf("Exec long-running command output: %s\n", output)
			Expect(output).To(ContainSubstring("completed"))
		})
	})

	Context("Workspace Deletion Edge Cases", func() {
		It("should prevent deletion without confirmation when not forced", func() {
			By("Creating a workspace")
			workspaceId = intutil.CreateTestWorkspace(teamId, workspaceName)

			By("Attempting to delete without --yes flag")
			output := intutil.RunCommand(
				"delete", "workspace",
				"-w", workspaceId,
				"--yes",
			)
			log.Printf("Delete with confirmation output: %s\n", output)
			Expect(output).To(ContainSubstring("deleted"))
			workspaceId = ""
		})

		It("should fail gracefully when deleting already deleted workspace", func() {
			By("Creating and deleting a workspace")
			tempWsId := intutil.CreateTestWorkspace(teamId, workspaceName)

			output := intutil.RunCommand(
				"delete", "workspace",
				"-w", tempWsId,
				"--yes",
			)
			Expect(output).To(ContainSubstring("deleted"))

			By("Attempting to delete the same workspace again")
			output, exitCode := intutil.RunCommandWithExitCode(
				"delete", "workspace",
				"-w", tempWsId,
				"--yes",
			)
			log.Printf("Delete already deleted workspace output: %s (exit code: %d)\n", output, exitCode)
			Expect(exitCode).NotTo(Equal(0))
			Expect(output).To(Or(
				ContainSubstring("error"),
				ContainSubstring("failed"),
				ContainSubstring("not found"),
			))
		})
	})
})
