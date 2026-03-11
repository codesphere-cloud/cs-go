// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"os"

	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func FailIfMissingEnvVars() (teamId, token string) {
	teamId = os.Getenv("CS_TEAM_ID")
	if teamId == "" {
		ginkgo.Fail("CS_TEAM_ID environment variable not set")
	}

	token = os.Getenv("CS_TOKEN")
	if token == "" {
		ginkgo.Fail("CS_TOKEN environment variable not set")
	}

	return teamId, token
}

// WithClearedWorkspaceEnv temporarily unsets CS_WORKSPACE_ID and WORKSPACE_ID,
// calls fn, then restores the original values.
func WithClearedWorkspaceEnv(fn func()) {
	originalWsId := os.Getenv("CS_WORKSPACE_ID")
	originalWsIdFallback := os.Getenv("WORKSPACE_ID")
	_ = os.Unsetenv("CS_WORKSPACE_ID")
	_ = os.Unsetenv("WORKSPACE_ID")
	defer func() {
		_ = os.Setenv("CS_WORKSPACE_ID", originalWsId)
		_ = os.Setenv("WORKSPACE_ID", originalWsIdFallback)
	}()
	fn()
}

// ExpectNonExistentWorkspaceError asserts that a command failed with a "not found" style error.
func ExpectNonExistentWorkspaceError(output string, exitCode int) {
	ginkgo.GinkgoHelper()
	gomega.Expect(exitCode).NotTo(gomega.Equal(0))
	gomega.Expect(output).To(gomega.Or(
		gomega.ContainSubstring("failed to get workspace"),
		gomega.ContainSubstring("not found"),
		gomega.ContainSubstring("404"),
	))
}
