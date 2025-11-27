// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"os"

	ginkgo "github.com/onsi/ginkgo/v2"
)

func SkipIfMissingEnvVars() (teamId, token string) {
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
