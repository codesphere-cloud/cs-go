// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMcpServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MCP Server Suite")
}
