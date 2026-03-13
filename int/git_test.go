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

var _ = Describe("Git Pull Integration Tests", Label("git"), func() {
	var (
		teamId        string
		workspaceName string
		workspaceId   string
	)

	BeforeEach(func() {
		teamId, _ = intutil.FailIfMissingEnvVars()
		workspaceName = intutil.NewWorkspaceName("git")
	})

	AfterEach(func() {
		if workspaceId != "" {
			By(fmt.Sprintf("Cleaning up: deleting workspace %s (ID: %s)", workspaceName, workspaceId))
			intutil.CleanupWorkspace(workspaceId)
			workspaceId = ""
		}
	})

	Context("Git Pull Command", func() {
		BeforeEach(func() {
			By("Creating a workspace")
			workspaceId = intutil.CreateTestWorkspace(teamId, workspaceName)
		})

		It("should execute git pull command", func() {
			By("Running git pull")
			output, exitCode := intutil.RunCommandWithExitCode(
				"git", "pull",
				"-w", workspaceId,
			)
			log.Printf("Git pull output: %s (exit code: %d)\n", output, exitCode)

			Expect(output).NotTo(BeEmpty())
		})
	})
})
