// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cs_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cs Suite")
}
