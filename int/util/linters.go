// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"os/exec"
)

const (
	HadolintTool    = "hadolint"
	ShellcheckTool  = "shellcheck"
	KubeconformTool = "kubeconform"
)

// runLinter executes a linting tool on the file at the given path.
// Returns an error if the tool is not found on $PATH or if linting reports issues.
func runLinter(tool string, args []string, path string) error {
	if _, err := exec.LookPath(tool); err != nil {
		return fmt.Errorf("%s not found on $PATH: %w", tool, err)
	}
	cmdArgs := append(args, path)
	cmd := exec.Command(tool, cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s errors in %s:\n%s", tool, path, string(output))
	}
	return nil
}

// LintDockerfile runs hadolint on the Dockerfile at the given path.
func LintDockerfile(path string) error {
	return runLinter(HadolintTool, []string{"--failure-threshold", "error"}, path)
}

// LintShellScript runs shellcheck on the shell script at the given path.
func LintShellScript(path string) error {
	return runLinter(ShellcheckTool, []string{"-S", "error"}, path)
}

// LintKubernetesManifest runs kubeconform on the Kubernetes manifest at the given path.
func LintKubernetesManifest(path string) error {
	return runLinter(KubeconformTool, nil, path)
}
