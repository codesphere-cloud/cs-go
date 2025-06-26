// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package io_test

import (
	"bytes"
	//"io"
	"strings"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	csio "github.com/codesphere-cloud/cs-go/pkg/io"
)

var _ = Describe("StreamOutput", func() {
	var (
		wg     sync.WaitGroup
		input  string
		output bytes.Buffer
	)
	BeforeEach(func() {
		input = "fake-output\n"
	})
	It("Streams output from input to output", func() {
		stringReader := strings.NewReader(input)
		csio.StreamOutput(&wg, stringReader, &output)
		wg.Wait()

		Expect(output.String()).To(Equal(input))
	})
})
