// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package exporter_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestExporter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Exporter Suite")
}
