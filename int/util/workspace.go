// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/codesphere-cloud/cs-go/api"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// CreateTestWorkspace creates a workspace with standard settings and returns the workspace ID.
// It fails the test if creation fails or no workspace ID is returned.
func CreateTestWorkspace(teamId, workspaceName string) string {
	ginkgo.GinkgoHelper()
	output := RunCommand(
		"create", "workspace", workspaceName,
		"-t", teamId,
		"-p", DefaultPlanId,
		"--timeout", DefaultCreateTimeout,
	)
	log.Printf("Create workspace output: %s\n", output)
	gomega.Expect(output).To(gomega.ContainSubstring(WorkspaceCreatedOutput))
	workspaceId := ExtractWorkspaceId(output)
	gomega.Expect(workspaceId).NotTo(gomega.BeEmpty())
	return workspaceId
}

// NewWorkspaceName generates a unique workspace name with the given prefix.
func NewWorkspaceName(prefix string) string {
	return fmt.Sprintf("cli-%s-test-%d", prefix, time.Now().Unix())
}

func CheckBillingStatus(teamId string) (bool, string) {
	testName := "billing-check-temp"
	output, exitCode := RunCommandWithExitCode(
		"create", "workspace", testName,
		"-t", teamId,
		"-p", "8",
		"--timeout", "1s",
	)

	if strings.Contains(output, "402") && strings.Contains(output, "Missing billing information") {
		return false, "Team does not have billing information configured (payment method and address required)"
	}

	if exitCode == 0 || strings.Contains(output, "Workspace created") {
		if wsId := ExtractWorkspaceId(output); wsId != "" {
			CleanupWorkspace(wsId)
		}
		return true, ""
	}

	return true, ""
}

func RunCommand(args ...string) string {
	output, _ := RunCommandWithExitCode(args...)
	return output
}

func RunCommandWithExitCode(args ...string) (string, int) {
	command := exec.Command("../cs", args...)

	command.Env = os.Environ()

	var outputBuffer bytes.Buffer
	command.Stdout = &outputBuffer
	command.Stderr = &outputBuffer

	err := command.Run()

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return outputBuffer.String(), exitCode
}

func ExtractWorkspaceId(output string) string {
	re := regexp.MustCompile(`ID:\s*(\d+)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

func CleanupWorkspace(workspaceId string) {
	if workspaceId == "" {
		return
	}

	output, exitCode := RunCommandWithExitCode("delete", "workspace", "-w", workspaceId, "--yes")
	if exitCode != 0 {
		log.Printf("Warning: Failed to cleanup workspace %s (exit code %d): %s\n", workspaceId, exitCode, output)
	} else {
		log.Printf("Cleanup workspace %s: success\n", workspaceId)
	}
}

func WaitForWorkspaceRunning(client *api.Client, workspaceId int, timeout time.Duration) error {
	return client.WaitForWorkspaceRunning(&api.Workspace{Id: workspaceId}, timeout)
}

func ScaleWorkspace(client *api.Client, workspaceId int, replicas int) error {
	return client.ScaleWorkspace(workspaceId, replicas)
}

func VerifyWorkspaceExists(workspaceId, teamId string) bool {
	output := RunCommand("list", "workspaces", "-t", teamId)
	return strings.Contains(output, workspaceId)
}

func VerifyWorkspaceDeleted(workspaceId, teamId string) bool {
	output := RunCommand("list", "workspaces", "-t", teamId)
	return !strings.Contains(output, workspaceId)
}

func ExtractTeamId(output string) string {
	re := regexp.MustCompile(`ID:\s*(\d+)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) >= 2 {
		return matches[1]
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "integration-test-") {
			re = regexp.MustCompile(`\s+(\d+)\s+`)
			matches = re.FindStringSubmatch(line)
			if len(matches) >= 2 {
				return matches[1]
			}
		}
	}

	return ""
}

func CleanupTeam(teamId string) {
	if teamId == "" {
		return
	}

	output, exitCode := RunCommandWithExitCode("delete", "team", "-t", teamId, "--force")
	if exitCode != 0 {
		log.Printf("Warning: Failed to cleanup team %s (exit code %d): %s\n", teamId, exitCode, output)
	} else {
		log.Printf("Cleanup team %s: %s\n", teamId, output)
	}
}
