// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package io_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Io Suite")
}
