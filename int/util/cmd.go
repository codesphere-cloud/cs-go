// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bytes"
	"fmt"
	. "github.com/onsi/gomega"
	"os/exec"
)

func RunCommandInBackground(outputBuffer *bytes.Buffer, args ...string) *exec.Cmd {
	command := exec.Command("../cs", args...)
	fmt.Println(args)
	command.Stdout = outputBuffer
	command.Stderr = outputBuffer

	go func() {
		err := command.Start()
		Expect(err).NotTo(HaveOccurred())
	}()
	return command
}
