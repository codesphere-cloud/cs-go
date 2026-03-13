// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import "time"

const (
	DefaultPlanId          = "8"
	DefaultCreateTimeout   = "15m"
	NonExistentWorkspaceId = "99999999"
	WorkspaceCreatedOutput = "Workspace created"
)

var PostCreateWaitTime = 5 * time.Second

// WorkspaceNamePrefixes contains all workspace name prefixes used by integration tests.
// This is used for global cleanup in AfterSuite to catch orphaned workspaces.
var WorkspaceNamePrefixes = []string{
	"cli-git-test-",
	"cli-pipeline-test-",
	"cli-log-test-",
	"cli-open-test-",
	"cli-setenv-test-",
	"cli-edge-test-",
	"cli-very-long-workspace-name-test-",
	"cli-wakeup-test-",
	"cli-curl-test-",
}
