package cmd_test

import (
	. "github.com/onsi/ginkgo/v2"
	//. "github.com/onsi/gomega"
	//"github.com/codesphere-cloud/cs-go/cli/cmd"
)

var _ = Describe("Cli/Cmd/Monitor", func() {

	Context("Command exits after 10 seconds with exit code 0", func() {
		It("Doesn't return an error", func() {

		})
	})
	Context("Command exits after 0.3 seconds", func() {
		Context("Command doesn't return an error", func() {
			It("Restarts immediately", func() {

			})

		})
		Context("Command returns an error", func() {
			It("Restarts after a delay of 5 seconds", func() {

			})
		})
	})
})
