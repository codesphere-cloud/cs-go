// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"errors"
	"fmt"
	"os/exec"
)

const (
	HadolintTool    = "hadolint"
	ShellcheckTool  = "shellcheck"
	KubeconformTool = "kubeconform"
)

var ErrToolNotFound = errors.New("tool not found on $PATH")

// runLinter executes a linting tool on the file at the given path.
// Returns ErrToolNotFound if the tool is not on $PATH, or a lint error if the tool reports issues.
func runLinter(tool string, args []string, path string) error {
	if _, err := exec.LookPath(tool); err != nil {
		return fmt.Errorf("%w: %s", ErrToolNotFound, tool)
	}
	cmdArgs := append(args, path)
	cmd := exec.Command(tool, cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) == 0 {
			return fmt.Errorf("%s failed for %s: %w", tool, path, err)
		}
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
