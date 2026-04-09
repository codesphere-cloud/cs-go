// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLinters(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Linters Suite")
}

var _ = Describe("Linters", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "linter-test-")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tempDir)).To(Succeed())
	})

	Context("runLinter", func() {
		It("should return ErrToolNotFound when the tool is not found on PATH", func() {
			err := runLinter("nonexistent-tool-xyz", nil, "somefile.txt")
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, ErrToolNotFound)).To(BeTrue())
		})

		It("should return an error when the tool exits with a non-zero code", func() {
			tmpFile := filepath.Join(tempDir, "test.txt")
			Expect(os.WriteFile(tmpFile, []byte("hello"), 0644)).To(Succeed())

			err := runLinter("false", nil, tmpFile)
			Expect(err).To(HaveOccurred())
		})

		It("should succeed when the tool runs without errors", func() {
			tmpFile := filepath.Join(tempDir, "test.txt")
			Expect(os.WriteFile(tmpFile, []byte("hello"), 0644)).To(Succeed())

			err := runLinter("true", nil, tmpFile)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should pass additional arguments to the tool", func() {
			tmpFile := filepath.Join(tempDir, "test.txt")
			Expect(os.WriteFile(tmpFile, []byte("hello world"), 0644)).To(Succeed())

			// grep -q "hello" <file> succeeds when the pattern matches
			err := runLinter("grep", []string{"-q", "hello"}, tmpFile)
			Expect(err).NotTo(HaveOccurred())

			// grep -q "notfound" <file> fails when the pattern does not match
			err = runLinter("grep", []string{"-q", "notfound"}, tmpFile)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("LintDockerfile", func() {
		It("should return ErrToolNotFound when hadolint is not on PATH", func() {
			origPath := os.Getenv("PATH")
			Expect(os.Setenv("PATH", tempDir)).To(Succeed())
			defer func() { Expect(os.Setenv("PATH", origPath)).To(Succeed()) }()

			err := LintDockerfile(filepath.Join(tempDir, "Dockerfile"))
			Expect(errors.Is(err, ErrToolNotFound)).To(BeTrue())
			Expect(err.Error()).To(ContainSubstring(HadolintTool))
		})
	})

	Context("LintShellScript", func() {
		It("should return ErrToolNotFound when shellcheck is not on PATH", func() {
			origPath := os.Getenv("PATH")
			Expect(os.Setenv("PATH", tempDir)).To(Succeed())
			defer func() { Expect(os.Setenv("PATH", origPath)).To(Succeed()) }()

			err := LintShellScript(filepath.Join(tempDir, "script.sh"))
			Expect(errors.Is(err, ErrToolNotFound)).To(BeTrue())
			Expect(err.Error()).To(ContainSubstring(ShellcheckTool))
		})
	})

	Context("LintKubernetesManifest", func() {
		It("should return ErrToolNotFound when kubeconform is not on PATH", func() {
			origPath := os.Getenv("PATH")
			Expect(os.Setenv("PATH", tempDir)).To(Succeed())
			defer func() { Expect(os.Setenv("PATH", origPath)).To(Succeed()) }()

			err := LintKubernetesManifest(filepath.Join(tempDir, "manifest.yml"))
			Expect(errors.Is(err, ErrToolNotFound)).To(BeTrue())
			Expect(err.Error()).To(ContainSubstring(KubeconformTool))
		})
	})
})
