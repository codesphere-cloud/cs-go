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

var _ = Describe("Start Pipeline Integration Tests", Label("pipeline"), func() {
	var (
		teamId        string
		workspaceName string
		workspaceId   string
	)

	BeforeEach(func() {
		teamId, _ = intutil.FailIfMissingEnvVars()
		workspaceName = intutil.NewWorkspaceName("pipeline")
	})

	AfterEach(func() {
		if workspaceId != "" {
			By(fmt.Sprintf("Cleaning up: deleting workspace %s (ID: %s)", workspaceName, workspaceId))
			intutil.CleanupWorkspace(workspaceId)
			workspaceId = ""
		}
	})

	Context("Start Pipeline Command", func() {
		BeforeEach(func() {
			By("Creating a workspace")
			workspaceId = intutil.CreateTestWorkspace(teamId, workspaceName)
		})

		It("should start pipeline successfully", func() {
			By("Starting pipeline")
			output, exitCode := intutil.RunCommandWithExitCode(
				"start", "pipeline",
				"-w", workspaceId,
			)
			log.Printf("Start pipeline output: %s (exit code: %d)\n", output, exitCode)

			Expect(output).NotTo(BeEmpty())
		})
	})
})
