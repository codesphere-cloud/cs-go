// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bytes"
	"log"
	"os"
	"os/exec"

	"github.com/onsi/gomega"
)

func RunCommandInBackground(outputBuffer *bytes.Buffer, args ...string) *exec.Cmd {
	command := exec.Command("../cs", args...)

	command.Env = os.Environ()

	log.Println(args)
	command.Stdout = outputBuffer
	command.Stderr = outputBuffer

	go func() {
		err := command.Start()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}()
	return command
}
