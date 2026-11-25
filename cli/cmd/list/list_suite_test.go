// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package list_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

func TestList(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "List Suite")
}

// newTestCommand returns a command carrying a background context, like cobra
// provides to RunE when a command is executed.
func newTestCommand() *cobra.Command {
	c := &cobra.Command{}
	c.SetContext(context.Background())
	return c
}
