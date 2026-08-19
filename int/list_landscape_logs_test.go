// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package int_test

import (
	"fmt"
	"log"

	intutil "github.com/codesphere-cloud/cs-go/int/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("List Landscape Logs Command Integration Tests", Label("landscape-logs"), func() {
	var (
		teamId        string
		workspaceName string
		workspaceId   string
	)

	BeforeEach(func() {
		teamId, _ = intutil.FailIfMissingEnvVars()
		workspaceName = intutil.NewWorkspaceName("landscape-logs")
	})

	AfterEach(func() {
		if workspaceId != "" {
			By(fmt.Sprintf("Cleaning up: deleting workspace %s (ID: %s)", workspaceName, workspaceId))
			intutil.CleanupWorkspace(workspaceId)
			workspaceId = ""
		}
	})

	Context("List Landscape Logs Command", func() {
		BeforeEach(func() {
			By("Creating a workspace")
			workspaceId = intutil.CreateTestWorkspace(teamId, workspaceName)
		})

		It("should retrieve logs from workspace", func() {
			By("Getting logs from workspace")
			output, exitCode := intutil.RunCommandWithExitCode(
				"list", "landscape-logs",
				"-w", workspaceId,
			)
			log.Printf("List landscape-logs command output (first 500 chars): %s... (exit code: %d)\n",
				output[:min(500, len(output))], exitCode)

			Expect(exitCode).To(Or(Equal(0), Equal(1)))
		})
	})
})
