// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// LintDockerfile runs hadolint on the Dockerfile at the given path.
// Skips gracefully if hadolint is not installed.
func LintDockerfile(path string) {
	GinkgoHelper()
	if _, err := exec.LookPath("hadolint"); err != nil {
		GinkgoWriter.Println("hadolint not found, skipping Dockerfile lint")
		return
	}
	cmd := exec.Command("hadolint", "--failure-threshold", "error", path)
	output, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "hadolint errors in %s:\n%s", path, string(output))
}

// LintShellScript runs shellcheck on the shell script at the given path.
// Skips gracefully if shellcheck is not installed.
func LintShellScript(path string) {
	GinkgoHelper()
	if _, err := exec.LookPath("shellcheck"); err != nil {
		GinkgoWriter.Println("shellcheck not found, skipping shell script lint")
		return
	}
	cmd := exec.Command("shellcheck", "-S", "error", path)
	output, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "shellcheck errors in %s:\n%s", path, string(output))
}

// LintKubernetesManifest runs kubeconform on the Kubernetes manifest at the given path.
// Skips gracefully if kubeconform is not installed.
func LintKubernetesManifest(path string) {
	GinkgoHelper()
	if _, err := exec.LookPath("kubeconform"); err != nil {
		GinkgoWriter.Println("kubeconform not found, skipping Kubernetes manifest lint")
		return
	}
	cmd := exec.Command("kubeconform", path)
	output, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "kubeconform errors in %s:\n%s", path, string(output))
}
