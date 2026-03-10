// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package int_test

import (
	"fmt"
	"log"

	intutil "github.com/codesphere-cloud/cs-go/int/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Command Error Handling Tests", Label("error-handling"), func() {
	It("should fail gracefully with non-existent workspace for all commands", func() {
		testCases := []struct {
			commandName string
			args        []string
		}{
			{"open workspace", []string{"open", "workspace", "-w", "99999999"}},
			{"log", []string{"log", "-w", "99999999"}},
			{"start pipeline", []string{"start", "pipeline", "-w", "99999999"}},
			{"git pull", []string{"git", "pull", "-w", "99999999"}},
			{"set-env", []string{"set-env", "-w", "99999999", "TEST_VAR=test"}},
			{"wake-up", []string{"wake-up", "-w", "99999999"}},
			{"curl", []string{"curl", "/", "-w", "99999999"}},
		}

		for _, tc := range testCases {
			By(fmt.Sprintf("Testing %s with non-existent workspace", tc.commandName))
			output, exitCode := intutil.RunCommandWithExitCode(tc.args...)
			log.Printf("%s non-existent workspace output: %s (exit code: %d)\n", tc.commandName, output, exitCode)
			Expect(exitCode).NotTo(Equal(0))
		}
	})
})

var _ = Describe("Command Error Handling Tests - Additional", Label("error-handling"), func() {
	It("should fail gracefully with non-existent workspace for all commands", func() {
		testCases := []struct {
			commandName string
			args        []string
		}{
			{"open workspace", []string{"open", "workspace", "-w", "99999999"}},
			{"log", []string{"log", "-w", "99999999"}},
			{"start pipeline", []string{"start", "pipeline", "-w", "99999999"}},
			{"git pull", []string{"git", "pull", "-w", "99999999"}},
			{"set-env", []string{"set-env", "-w", "99999999", "TEST_VAR=test"}},
		}

		for _, tc := range testCases {
			By(fmt.Sprintf("Testing %s with non-existent workspace", tc.commandName))
			output, exitCode := intutil.RunCommandWithExitCode(tc.args...)
			log.Printf("%s non-existent workspace output: %s (exit code: %d)\n", tc.commandName, output, exitCode)
			Expect(exitCode).NotTo(Equal(0))
		}
	})
})
