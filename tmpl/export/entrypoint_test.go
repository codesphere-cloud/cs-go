// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package export_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/codesphere-cloud/cs-go/pkg/ci"
	"github.com/codesphere-cloud/cs-go/tmpl/export"
)

var _ = Describe("CreateEntrypoint", func() {
	var (
		entrypointConfig export.EntrypointTemplateConfig
	)

	Context("No run steps are provided", func() {
		JustBeforeEach(func() {
			entrypointConfig = export.EntrypointTemplateConfig{
				RunSteps: []ci.Step{},
			}
		})
		It("should return an error", func() {
			_, err := export.CreateEntrypoint(entrypointConfig)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("at least one run step is required"))
		})
	})

	Context("All values are provided", func() {
		JustBeforeEach(func() {
			entrypointConfig = export.EntrypointTemplateConfig{
				RunSteps: []ci.Step{{
					Name:    "Start web service",
					Command: "npm start",
				}},
			}
		})
		It("Creates an entrypoint script with the correct run steps", func() {
			entrypoint, err := export.CreateEntrypoint(entrypointConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(entrypoint)).To(ContainSubstring("#!/bin/bash"))
			Expect(string(entrypoint)).To(ContainSubstring("# Start web service"))
			Expect(string(entrypoint)).To(ContainSubstring("npm start"))
		})
	})
})
