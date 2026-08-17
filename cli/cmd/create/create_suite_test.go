// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package create_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCreate(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Create Suite")
}
