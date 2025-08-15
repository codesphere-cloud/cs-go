// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package exporter

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

func isCommandAvailable(name string) bool {
	cmd := exec.Command("command", "-v", name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func BuildImage(ctx context.Context, dockerfile string, tag string, buildContext string) error {
	var cmd *exec.Cmd
	if isCommandAvailable("docker") {

		cmd = exec.CommandContext(ctx, "docker", "build", "-f", dockerfile, "-t", tag, ".")
	} else if isCommandAvailable("podman") {
		cmd = exec.CommandContext(ctx, "podman", "build", "-f", dockerfile, "-t", tag, ".")
	} else {
		return fmt.Errorf("neither 'docker' nor 'podman' command is available")
	}
	cmd.Dir = buildContext
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("build failed with exit status %w", err)
	}

	return nil
}

func PushImage(ctx context.Context, tag string) error {
	var cmd *exec.Cmd
	if isCommandAvailable("docker") {
		cmd = exec.CommandContext(ctx, "docker", "push", tag)
	} else if isCommandAvailable("podman") {
		cmd = exec.CommandContext(ctx, "podman", "push", tag)
	} else {
		return fmt.Errorf("neither 'docker' nor 'podman' command is available")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("push failed with exit status %w", err)
	}

	return nil
}
