// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package int_test

import (
	"os"
	"testing"

	intutil "github.com/codesphere-cloud/cs-go/int/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Int Suite")
}

var _ = AfterSuite(func() {
	teamId := os.Getenv("CS_TEAM_ID")
	if teamId == "" {
		return
	}

	GinkgoWriter.Println("Running global cleanup for any orphaned test workspaces...")

	prefixes := []string{
		"cli-git-test-",
		"cli-pipeline-test-",
		"cli-log-test-",
		"cli-open-test-",
		"cli-setenv-test-",
		"cli-edge-test-",
		"cli-very-long-workspace-name-test-",
	}

	for _, prefix := range prefixes {
		intutil.CleanupAllWorkspacesInTeam(teamId, prefix)
	}

	GinkgoWriter.Println("Global cleanup complete")
})
