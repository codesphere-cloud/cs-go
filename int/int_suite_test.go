// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package int_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Int Suite")
}
