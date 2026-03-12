// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	HadolintTool    = "hadolint"
	ShellcheckTool  = "shellcheck"
	KubeconformTool = "kubeconform"
)

// runLinter executes a linting tool on the file at the given path.
// If the tool is not found on $PATH the lint step is skipped with a warning.
func runLinter(tool string, args []string, path string) {
	GinkgoHelper()
	if _, err := exec.LookPath(tool); err != nil {
		GinkgoWriter.Println(fmt.Sprintf("%s not found, skipping lint for %s", tool, path))
		return
	}
	cmdArgs := append(args, path)
	cmd := exec.Command(tool, cmdArgs...)
	output, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "%s errors in %s:\n%s", tool, path, string(output))
}

// LintDockerfile runs hadolint on the Dockerfile at the given path.
func LintDockerfile(path string) {
	GinkgoHelper()
	runLinter(HadolintTool, []string{"--failure-threshold", "error"}, path)
}

// LintShellScript runs shellcheck on the shell script at the given path.
func LintShellScript(path string) {
	GinkgoHelper()
	runLinter(ShellcheckTool, []string{"-S", "error"}, path)
}

// LintKubernetesManifest runs kubeconform on the Kubernetes manifest at the given path.
func LintKubernetesManifest(path string) {
	GinkgoHelper()
	runLinter(KubeconformTool, nil, path)
}
