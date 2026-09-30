// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package shared_test

import (
	"github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ValidateUserID", func() {
	It("accepts zero", func() {
		Expect(shared.ValidateUserID(0)).To(Succeed())
	})
	It("accepts positive IDs", func() {
		Expect(shared.ValidateUserID(42)).To(Succeed())
	})
	It("rejects negative IDs", func() {
		Expect(shared.ValidateUserID(-1)).To(MatchError("user ID must be non-negative"))
	})
})
