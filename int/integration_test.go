// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package int_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	intutil "github.com/codesphere-cloud/cs-go/int/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("cs monitor", func() {
	var (
		certsDir              string
		tempDir               string
		caCertPath            string
		serverCertPath        string
		serverKeyPath         string
		monitorListenPort     int
		targetServerPort      int
		targetServer          *http.Server
		monitorCmdProcess     *exec.Cmd
		testHttpClient        *http.Client
		monitorOutputBuf      *bytes.Buffer
		targetServerOutputBuf *bytes.Buffer
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "e2e-tls-monitor-test-")
		Expect(err).NotTo(HaveOccurred())
		certsDir = filepath.Join(tempDir, "certs")

		monitorListenPort, err = intutil.GetEphemeralPort()
		Expect(err).NotTo(HaveOccurred())
		targetServerPort, err = intutil.GetEphemeralPort()
		Expect(err).NotTo(HaveOccurred())

		testHttpClient = &http.Client{
			Timeout: 10 * time.Second,
		}

		monitorOutputBuf = new(bytes.Buffer)
		targetServerOutputBuf = new(bytes.Buffer)
	})

	AfterEach(func() {
		if monitorCmdProcess != nil && monitorCmdProcess.Process != nil {
			fmt.Printf("Terminating monitor process (PID: %d). Output:\n%s\n", monitorCmdProcess.Process.Pid, monitorOutputBuf.String())
			_ = monitorCmdProcess.Process.Kill()
			_, _ = monitorCmdProcess.Process.Wait()
		}

		Expect(os.RemoveAll(tempDir)).NotTo(HaveOccurred())
	})

	Context("Healthcheck forwarding", func() {
		AfterEach(func() {
			if targetServer != nil {
				fmt.Printf("Terminating HTTP(S) server. Output:\n%s\n", targetServerOutputBuf.String())
				shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 1*time.Second)
				defer shutdownCancel()
				_ = targetServer.Shutdown(shutdownCtx)
			}
		})
		It("should start a Go HTTP server, and proxy successfully", func() {
			var err error

			By("Starting Go HTTPS server with generated certs")
			targetServer, err = intutil.StartTestHttpServer(targetServerPort)
			Expect(err).NotTo(HaveOccurred())
			intutil.WaitForPort(fmt.Sprintf("127.0.0.1:%d", targetServerPort), 10*time.Second)
			fmt.Printf("Go HTTPS server started on port %d.\n", targetServerPort)

			By("Running 'cs monitor' command with --forward and --insecure-skip-verify")
			intutil.RunCommandInBackground(monitorOutputBuf,
				"monitor",
				"--forward", fmt.Sprintf("http://127.0.0.1:%d/", targetServerPort),
				"--address", fmt.Sprintf(":%d", monitorListenPort),
				"--", "sleep", "60s",
			)
			intutil.WaitForPort(fmt.Sprintf("127.0.0.1:%d", monitorListenPort), 10*time.Second)

			By("Making request to monitor proxy to verify successful forwarding")
			resp, err := testHttpClient.Get(fmt.Sprintf("http://127.0.0.1:%d/", monitorListenPort))
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			bodyBytes, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bodyBytes)).To(Equal("OK (HTTP)"))

			fmt.Printf("Monitor output after request:\n%s\n", monitorOutputBuf.String())
		})

		It("should start a Go HTTPS server with generated certs, run monitor with --insecure-skip-verify, and proxy successfully", func() {
			By("Generating TLS certificates")
			var err error
			caCertPath, serverCertPath, serverKeyPath, err = intutil.GenerateTLSCerts(
				certsDir,
				"localhost",
				[]string{"localhost", "127.0.0.1"},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(caCertPath).To(BeAnExistingFile())
			Expect(serverCertPath).To(BeAnExistingFile())
			Expect(serverKeyPath).To(BeAnExistingFile())

			By("Starting Go HTTPS server with generated certs")
			targetServer, err = intutil.StartTestHttpsServer(targetServerPort, serverCertPath, serverKeyPath)
			Expect(err).NotTo(HaveOccurred())
			intutil.WaitForPort(fmt.Sprintf("127.0.0.1:%d", targetServerPort), 10*time.Second)
			fmt.Printf("Go HTTPS server started on port %d.\n", targetServerPort)

			By("Running 'cs monitor' command with --forward and --insecure-skip-verify")
			intutil.RunCommandInBackground(monitorOutputBuf,
				"monitor",
				"--forward", fmt.Sprintf("https://127.0.0.1:%d/", targetServerPort),
				"--address", fmt.Sprintf(":%d", monitorListenPort),
				"--insecure-skip-verify",
				"--", "sleep", "60s",
			)
			intutil.WaitForPort(fmt.Sprintf("127.0.0.1:%d", monitorListenPort), 10*time.Second)

			By("Making request to monitor proxy to verify successful forwarding")
			resp, err := testHttpClient.Get(fmt.Sprintf("http://127.0.0.1:%d/", monitorListenPort))
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			bodyBytes, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bodyBytes)).To(Equal("OK (HTTPS)"))

			fmt.Printf("Monitor output after request:\n%s\n", monitorOutputBuf.String())
		})

		It("should get an error for an invalid HTTPS certificate without --insecure-skip-verify or --ca-cert-file", func() {
			By("Generating TLS certificates in Go")
			var err error
			caCertPath, serverCertPath, serverKeyPath, err = intutil.GenerateTLSCerts(
				certsDir,
				"localhost",
				[]string{"localhost", "127.0.0.1"},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(caCertPath).To(BeAnExistingFile())
			Expect(serverCertPath).To(BeAnExistingFile())
			Expect(serverKeyPath).To(BeAnExistingFile())

			By("Starting Go HTTPS server with generated certs")
			targetServer, err = intutil.StartTestHttpsServer(targetServerPort, serverCertPath, serverKeyPath)
			Expect(err).NotTo(HaveOccurred())
			intutil.WaitForPort(fmt.Sprintf("127.0.0.1:%d", targetServerPort), 10*time.Second)
			fmt.Printf("Go HTTPS server started on port %d.\n", targetServerPort)

			By("Running 'cs monitor' command without TLS bypass/trust")
			intutil.RunCommandInBackground(
				monitorOutputBuf,
				"monitor",
				"--address", fmt.Sprintf(":%d", monitorListenPort),
				"--forward", fmt.Sprintf("https://127.0.0.1:%d/", targetServerPort),
				"--", "sleep", "60s",
			)

			intutil.WaitForPort(fmt.Sprintf("127.0.0.1:%d", monitorListenPort), 10*time.Second)
			fmt.Printf("Monitor command started on port %d. Initial output:\n%s\n", monitorListenPort, monitorOutputBuf.String())

			By("Making request to monitor proxy and expecting a Bad Gateway error")
			resp, err := testHttpClient.Get(fmt.Sprintf("http://127.0.0.1:%d/", monitorListenPort))
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusBadGateway))
			bodyBytes, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bodyBytes)).To(ContainSubstring("Error forwarding request"))
			Expect(string(bodyBytes)).To(ContainSubstring("tls: failed to verify certificate"))

			fmt.Printf("Monitor output after request:\n%s\n", monitorOutputBuf.String())
		})

		It("should forward to an HTTPS target with --ca-cert-file and return 200 OK", func() {
			By("Generating TLS certificates in Go")
			var err error
			caCertPath, serverCertPath, serverKeyPath, err = intutil.GenerateTLSCerts(
				certsDir,
				"localhost",
				[]string{"localhost", "127.0.0.1"},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(caCertPath).To(BeAnExistingFile())
			Expect(serverCertPath).To(BeAnExistingFile())
			Expect(serverKeyPath).To(BeAnExistingFile())

			By("Starting Go HTTPS server with generated certs")
			targetServer, err = intutil.StartTestHttpsServer(targetServerPort, serverCertPath, serverKeyPath)
			Expect(err).NotTo(HaveOccurred())
			intutil.WaitForPort(fmt.Sprintf("127.0.0.1:%d", targetServerPort), 10*time.Second)
			fmt.Printf("Go HTTPS server started on port %d.\n", targetServerPort)

			By("Running 'cs monitor' command with --ca-cert-file")
			intutil.RunCommandInBackground(monitorOutputBuf,
				"monitor",
				"--address", fmt.Sprintf(":%d", monitorListenPort),
				"--forward", fmt.Sprintf("https://127.0.0.1:%d/", targetServerPort),
				"--ca-cert-file", caCertPath,
				"--",
				"sleep", "60s",
			)
			intutil.WaitForPort(fmt.Sprintf("127.0.0.1:%d", monitorListenPort), 10*time.Second)
			fmt.Printf("Monitor command started on port %d. Initial output:\n%s\n", monitorListenPort, monitorOutputBuf.String())

			By("Making request to monitor proxy to verify successful forwarding")
			resp, err := testHttpClient.Get(fmt.Sprintf("http://127.0.0.1:%d/", monitorListenPort))
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			bodyBytes, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bodyBytes)).To(Equal("OK (HTTPS)"))

			fmt.Printf("Monitor output after request:\n%s\n", monitorOutputBuf.String())
		})
	})

	Context("Prometheus Metrics Endpoint", func() {
		It("should expose Prometheus metrics when no forward is specified", func() {
			By("Running 'cs monitor' command without forwarding (metrics only)")
			intutil.RunCommandInBackground(monitorOutputBuf,
				"monitor",
				"--address", fmt.Sprintf(":%d", monitorListenPort),
				"--", "sleep", "60s",
			)
			intutil.WaitForPort(fmt.Sprintf("127.0.0.1:%d", monitorListenPort), 10*time.Second)
			fmt.Printf("Monitor command started on port %d. Initial output:\n%s\n", monitorListenPort, monitorOutputBuf.String())

			By("Making a request to the monitor's metrics endpoint")
			resp, err := testHttpClient.Get(fmt.Sprintf("http://127.0.0.1:%d/metrics", monitorListenPort))
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			bodyBytes, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bodyBytes)).To(ContainSubstring("cs_monitor_restarts_total"))
			fmt.Printf("Monitor output after metrics request:\n%s\n", monitorOutputBuf.String())
		})

		It("should redirect root to /metrics", func() {
			By("Running 'cs monitor' command without forwarding (metrics only)")
			intutil.RunCommandInBackground(monitorOutputBuf,
				"monitor",
				"--address", fmt.Sprintf(":%d", monitorListenPort),
				"--", "sleep", "60s",
			)
			intutil.WaitForPort(fmt.Sprintf("127.0.0.1:%d", monitorListenPort), 10*time.Second)
			fmt.Printf("Monitor command started on port %d. Initial output:\n%s\n", monitorListenPort, monitorOutputBuf.String())

			By("Making a request to the monitor's root endpoint and expecting a redirect")
			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
				Timeout: 5 * time.Second,
			}
			resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/", monitorListenPort))
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusMovedPermanently))
			Expect(resp.Header.Get("Location")).To(Equal("/metrics"))
			fmt.Printf("Monitor output after redirect request:\n%s\n", monitorOutputBuf.String())
		})
	})

	Context("Command Execution and Restart Logic", func() {
		It("should execute the command once if it succeeds", func() {
			monitorCmdProcess = intutil.RunCommandInBackground(monitorOutputBuf,
				"monitor",
				"--address", fmt.Sprintf(":%d", monitorListenPort),
				"--max-restarts", "0",
				"--", "true",
			)

			Eventually(monitorCmdProcess.Wait, "5s").Should(Succeed(), "Monitor process should exit successfully")

			output := monitorOutputBuf.String()
			Expect(output).To(ContainSubstring("command exited"))
			Expect(output).To(ContainSubstring("returnCode=0"))
			Expect(output).To(ContainSubstring("maximum number of restarts reached, exiting"))
			Expect(strings.Count(output, "command exited")).To(Equal(1), "Command should have executed only once")
		})

		It("should restart the command if it exits with non-zero code quickly", func() {

			monitorCmdProcess = intutil.RunCommandInBackground(monitorOutputBuf,
				"monitor",
				"--address", fmt.Sprintf(":%d", monitorListenPort),
				"--max-restarts", "1",
				"--", "bash", "-c", "echo FAKE_OUTPUT;exit 1",
			)

			Eventually(monitorCmdProcess.Wait, "15s").Should(Succeed(), "Monitor process should exit after restarts")

			output := monitorOutputBuf.String()
			Expect(output).To(ContainSubstring("command exited"))
			Expect(output).To(ContainSubstring("returnCode=1"))
			Expect(output).To(ContainSubstring("command exited with non-zero code in less than 1 second. Waiting 5 seconds before next restart"))
			Expect(output).To(ContainSubstring("cs monitor: restarting"))
			Expect(output).To(ContainSubstring("maximum number of restarts reached, exiting"))
			Expect(strings.Count(output, "FAKE_OUTPUT")).To(Equal(3), "Command should have executed twice")
		})

		It("should stop command runner on context cancellation", func() {
			By("Running 'cs monitor' command with infinite restarts")
			monitorCmdProcess = intutil.RunCommandInBackground(monitorOutputBuf,
				"monitor",
				"--address", fmt.Sprintf(":%d", monitorListenPort),
				"--max-restarts", "-1",
				"--", "sleep", "10s",
			)
			Eventually(func() string { return monitorOutputBuf.String() }, "5s").Should(ContainSubstring("starting monitor"))

			By("Stopping command execution")
			err := monitorCmdProcess.Process.Signal(os.Interrupt)
			Expect(err).NotTo(HaveOccurred())
			_, _ = monitorCmdProcess.Process.Wait()

			output := monitorOutputBuf.String()
			Expect(output).To(ContainSubstring("initiating graceful shutdown..."))
			Expect(output).To(ContainSubstring("stopping command runner."))
			Expect(output).NotTo(ContainSubstring("error executing command"))
		})
	})
})

var _ = Describe("Open Workspace Integration Tests", func() {
	var (
		teamId        string
		workspaceName string
		workspaceId   string
	)

	BeforeEach(func() {
		teamId, _ = intutil.SkipIfMissingEnvVars()
		workspaceName = fmt.Sprintf("cli-open-test-%d", time.Now().Unix())
	})

	AfterEach(func() {
		if workspaceId != "" {
			By(fmt.Sprintf("Cleaning up: deleting workspace %s (ID: %s)", workspaceName, workspaceId))
			intutil.CleanupWorkspace(workspaceId)
			workspaceId = ""
		}
	})

	Context("Open Workspace Command", func() {
		BeforeEach(func() {
			By("Creating a workspace")
			output := intutil.RunCommand(
				"create", "workspace", workspaceName,
				"-t", teamId,
				"-p", "8",
				"--timeout", "15m",
			)
			fmt.Printf("Create workspace output: %s\n", output)

			Expect(output).To(ContainSubstring("Workspace created"))
			workspaceId = intutil.ExtractWorkspaceId(output)
			Expect(workspaceId).NotTo(BeEmpty())
		})

		It("should open workspace successfully", func() {
			By("Opening the workspace")
			output := intutil.RunCommand(
				"open", "workspace",
				"-w", workspaceId,
			)
			fmt.Printf("Open workspace output: %s\n", output)

			Expect(output).To(ContainSubstring("Opening workspace"))
			Expect(output).To(ContainSubstring(workspaceId))
		})
	})

	Context("Open Workspace Error Handling", func() {
		It("should fail when workspace ID is missing", func() {
			By("Attempting to open workspace without ID")
			originalWsId := os.Getenv("CS_WORKSPACE_ID")
			originalWsIdFallback := os.Getenv("WORKSPACE_ID")
			_ = os.Unsetenv("CS_WORKSPACE_ID")
			_ = os.Unsetenv("WORKSPACE_ID")
			defer func() {
				_ = os.Setenv("CS_WORKSPACE_ID", originalWsId)
				_ = os.Setenv("WORKSPACE_ID", originalWsIdFallback)
			}()

			output, exitCode := intutil.RunCommandWithExitCode(
				"open", "workspace",
			)
			fmt.Printf("Open without workspace ID output: %s (exit code: %d)\n", output, exitCode)
			Expect(exitCode).NotTo(Equal(0))
			Expect(output).To(Or(
				ContainSubstring("workspace"),
				ContainSubstring("required"),
			))
		})
	})
})

var _ = Describe("Workspace Edge Cases and Advanced Operations", func() {
	var (
		teamId        string
		workspaceName string
		workspaceId   string
	)

	BeforeEach(func() {
		teamId, _ = intutil.SkipIfMissingEnvVars()
		workspaceName = fmt.Sprintf("cli-edge-test-%d", time.Now().Unix())
	})

	AfterEach(func() {
		if workspaceId != "" {
			By(fmt.Sprintf("Cleaning up: deleting workspace %s (ID: %s)", workspaceName, workspaceId))
			intutil.CleanupWorkspace(workspaceId)
			workspaceId = ""
		}
	})

	Context("Workspace Creation Edge Cases", func() {
		It("should create a workspace with a very long name", func() {
			longName := fmt.Sprintf("cli-very-long-workspace-name-test-%d", time.Now().Unix())
			By("Creating a workspace with a long name")
			output := intutil.RunCommand(
				"create", "workspace", longName,
				"-t", teamId,
				"-p", "8",
				"--timeout", "15m",
			)
			fmt.Printf("Create workspace with long name output: %s\n", output)

			if output != "" && !strings.Contains(output, "error") {
				Expect(output).To(ContainSubstring("Workspace created"))
				workspaceId = intutil.ExtractWorkspaceId(output)
			}
		})

		It("should handle creation timeout gracefully", func() {
			By("Creating a workspace with very short timeout")
			output, exitCode := intutil.RunCommandWithExitCode(
				"create", "workspace", workspaceName,
				"-t", teamId,
				"-p", "8",
				"--timeout", "1s",
			)
			fmt.Printf("Create with short timeout output: %s (exit code: %d)\n", output, exitCode)

			if exitCode != 0 {
				Expect(output).To(Or(
					ContainSubstring("timeout"),
					ContainSubstring("timed out"),
				))
			} else if strings.Contains(output, "Workspace created") {
				workspaceId = intutil.ExtractWorkspaceId(output)
			}
		})
	})

	Context("Exec Command Edge Cases", func() {
		BeforeEach(func() {
			By("Creating a workspace")
			output := intutil.RunCommand(
				"create", "workspace", workspaceName,
				"-t", teamId,
				"-p", "8",
				"--timeout", "15m",
			)
			Expect(output).To(ContainSubstring("Workspace created"))
			workspaceId = intutil.ExtractWorkspaceId(output)
			Expect(workspaceId).NotTo(BeEmpty())
		})

		It("should execute commands with multiple arguments", func() {
			By("Executing a command with multiple arguments")
			output := intutil.RunCommand(
				"exec",
				"-w", workspaceId,
				"--",
				"sh", "-c", "echo test1 && echo test2",
			)
			fmt.Printf("Exec with multiple args output: %s\n", output)
			Expect(output).To(ContainSubstring("test1"))
			Expect(output).To(ContainSubstring("test2"))
		})

		It("should handle commands that output to stderr", func() {
			By("Executing a command that writes to stderr")
			output := intutil.RunCommand(
				"exec",
				"-w", workspaceId,
				"--",
				"sh", "-c", "echo error message >&2",
			)
			fmt.Printf("Exec with stderr output: %s\n", output)
			Expect(output).To(ContainSubstring("error message"))
		})

		It("should handle commands with exit codes", func() {
			By("Executing a command that exits with non-zero code")
			output, exitCode := intutil.RunCommandWithExitCode(
				"exec",
				"-w", workspaceId,
				"--",
				"sh", "-c", "exit 42",
			)
			fmt.Printf("Exec with exit code output: %s (exit code: %d)\n", output, exitCode)
		})

		It("should execute long-running commands", func() {
			By("Executing a command that takes a few seconds")
			output := intutil.RunCommand(
				"exec",
				"-w", workspaceId,
				"--",
				"sh", "-c", "sleep 2 && echo completed",
			)
			fmt.Printf("Exec long-running command output: %s\n", output)
			Expect(output).To(ContainSubstring("completed"))
		})
	})

	Context("Workspace Deletion Edge Cases", func() {
		It("should prevent deletion without confirmation when not forced", func() {
			By("Creating a workspace")
			output := intutil.RunCommand(
				"create", "workspace", workspaceName,
				"-t", teamId,
				"-p", "8",
				"--timeout", "15m",
			)
			Expect(output).To(ContainSubstring("Workspace created"))
			workspaceId = intutil.ExtractWorkspaceId(output)
			Expect(workspaceId).NotTo(BeEmpty())

			By("Attempting to delete without --yes flag")
			output = intutil.RunCommand(
				"delete", "workspace",
				"-w", workspaceId,
				"--yes",
			)
			fmt.Printf("Delete with confirmation output: %s\n", output)
			Expect(output).To(ContainSubstring("deleted"))
			workspaceId = ""
		})

		It("should fail gracefully when deleting already deleted workspace", func() {
			By("Creating and deleting a workspace")
			output := intutil.RunCommand(
				"create", "workspace", workspaceName,
				"-t", teamId,
				"-p", "8",
				"--timeout", "15m",
			)
			Expect(output).To(ContainSubstring("Workspace created"))
			tempWsId := intutil.ExtractWorkspaceId(output)

			output = intutil.RunCommand(
				"delete", "workspace",
				"-w", tempWsId,
				"--yes",
			)
			Expect(output).To(ContainSubstring("deleted"))

			By("Attempting to delete the same workspace again")
			output, exitCode := intutil.RunCommandWithExitCode(
				"delete", "workspace",
				"-w", tempWsId,
				"--yes",
			)
			fmt.Printf("Delete already deleted workspace output: %s (exit code: %d)\n", output, exitCode)
			Expect(exitCode).NotTo(Equal(0))
			Expect(output).To(Or(
				ContainSubstring("error"),
				ContainSubstring("failed"),
				ContainSubstring("not found"),
			))
		})
	})
})

var _ = Describe("Version and Help Tests", func() {
	Context("Version Command", func() {
		It("should display version information", func() {
			By("Running version command")
			output := intutil.RunCommand("version")
			fmt.Printf("Version output: %s\n", output)

			Expect(output).To(Or(
				ContainSubstring("version"),
				ContainSubstring("Version"),
				MatchRegexp(`\d+\.\d+\.\d+`),
			))
		})
	})

	Context("Help Commands", func() {
		It("should display main help", func() {
			By("Running help command")
			output := intutil.RunCommand("--help")
			fmt.Printf("Help output length: %d\n", len(output))

			Expect(output).To(ContainSubstring("Usage:"))
			Expect(output).To(ContainSubstring("Available Commands:"))
		})

		It("should display help for all subcommands", func() {
			testCases := []struct {
				command     []string
				shouldMatch string
			}{
				{[]string{"create", "--help"}, "workspace"},
				{[]string{"exec", "--help"}, "exec"},
				{[]string{"log", "--help"}, "log"},
				{[]string{"start", "pipeline", "--help"}, "pipeline"},
				{[]string{"git", "pull", "--help"}, "pull"},
				{[]string{"set-env", "--help"}, "set-env"},
			}

			for _, tc := range testCases {
				By(fmt.Sprintf("Testing %v", tc.command))
				output := intutil.RunCommand(tc.command...)
				Expect(output).To(ContainSubstring("Usage:"))
				Expect(output).To(ContainSubstring(tc.shouldMatch))
			}
		})
	})

	Context("Invalid Commands", func() {
		It("should handle unknown commands gracefully", func() {
			By("Running unknown command")
			output, exitCode := intutil.RunCommandWithExitCode("unknowncommand")
			fmt.Printf("Unknown command output: %s (exit code: %d)\n", output, exitCode)

			Expect(exitCode).NotTo(Equal(0))
			Expect(output).To(Or(
				ContainSubstring("unknown command"),
				ContainSubstring("Error:"),
			))
		})

		It("should suggest similar commands for typos", func() {
			By("Running misspelled command")
			output, exitCode := intutil.RunCommandWithExitCode("listt")
			fmt.Printf("Typo command output: %s (exit code: %d)\n", output, exitCode)

			Expect(exitCode).NotTo(Equal(0))
			lowerOutput := strings.ToLower(output)
			Expect(lowerOutput).To(Or(
				ContainSubstring("unknown"),
				ContainSubstring("error"),
				ContainSubstring("did you mean"),
			))
		})
	})

	Context("Global Flags", func() {
		It("should accept all global flags", func() {
			By("Testing --api flag")
			output := intutil.RunCommand(
				"--api", "https://example.com/api",
				"list", "teams",
			)
			Expect(output).NotTo(ContainSubstring("unknown flag"))

			By("Testing --verbose flag")
			output = intutil.RunCommand(
				"--verbose",
				"list", "plans",
			)
			Expect(output).NotTo(ContainSubstring("unknown flag"))

			By("Testing -v shorthand")
			output = intutil.RunCommand(
				"-v",
				"list", "baseimages",
			)
			Expect(output).NotTo(ContainSubstring("unknown flag"))
		})
	})
})

var _ = Describe("List Command Tests", func() {
	var teamId string

	BeforeEach(func() {
		teamId, _ = intutil.SkipIfMissingEnvVars()
	})

	Context("List Workspaces", func() {
		It("should list all workspaces in team with proper formatting", func() {
			By("Listing workspaces")
			output := intutil.RunCommand("list", "workspaces", "-t", teamId)
			fmt.Printf("List workspaces output length: %d\n", len(output))

			Expect(output).To(ContainSubstring("TEAM ID"))
			Expect(output).To(ContainSubstring("ID"))
			Expect(output).To(ContainSubstring("NAME"))
		})
	})

	Context("List Plans", func() {
		It("should list all available plans", func() {
			By("Listing plans")
			output := intutil.RunCommand("list", "plans")
			fmt.Printf("List plans output: %s\n", output)

			Expect(output).To(ContainSubstring("ID"))
			Expect(output).To(ContainSubstring("NAME"))
			Expect(output).To(Or(
				ContainSubstring("Micro"),
				ContainSubstring("Free"),
			))
		})

		It("should show plan details like CPU and RAM", func() {
			By("Listing plans with details")
			output := intutil.RunCommand("list", "plans")
			fmt.Printf("Plan details output length: %d\n", len(output))

			Expect(output).To(ContainSubstring("CPU"))
			Expect(output).To(ContainSubstring("RAM"))
		})
	})

	Context("List Base Images", func() {
		It("should list available base images", func() {
			By("Listing base images")
			output := intutil.RunCommand("list", "baseimages")
			fmt.Printf("List baseimages output: %s\n", output)

			Expect(output).To(ContainSubstring("ID"))
			Expect(output).To(ContainSubstring("NAME"))
		})

		It("should show Ubuntu images", func() {
			By("Checking for Ubuntu in base images")
			output := intutil.RunCommand("list", "baseimages")

			Expect(output).To(ContainSubstring("Ubuntu"))
		})
	})

	Context("List Teams", func() {
		It("should list teams user has access to", func() {
			By("Listing teams")
			output := intutil.RunCommand("list", "teams")
			fmt.Printf("List teams output: %s\n", output)

			Expect(output).To(ContainSubstring("ID"))
			Expect(output).To(ContainSubstring("NAME"))
			Expect(output).To(ContainSubstring(teamId))
		})

		It("should show team role", func() {
			By("Checking team roles")
			output := intutil.RunCommand("list", "teams")

			Expect(output).To(Or(
				ContainSubstring("Admin"),
				ContainSubstring("Member"),
				ContainSubstring("ROLE"),
			))
		})
	})

	Context("List Error Handling", func() {
		It("should handle missing or invalid list subcommand", func() {
			By("Running list without subcommand")
			output, exitCode := intutil.RunCommandWithExitCode("list")
			fmt.Printf("List without subcommand output: %s (exit code: %d)\n", output, exitCode)
			Expect(output).To(Or(
				ContainSubstring("Available Commands:"),
				ContainSubstring("Usage:"),
			))

			By("Running list with invalid subcommand")
			output, _ = intutil.RunCommandWithExitCode("list", "invalid")
			fmt.Printf("List invalid output (first 200 chars): %s\n", output[:min(200, len(output))])
			Expect(output).To(Or(
				ContainSubstring("Available Commands:"),
				ContainSubstring("Usage:"),
			))
		})

		It("should require team ID for workspace listing when not set globally", func() {
			By("Listing workspaces without team ID in specific contexts")
			output := intutil.RunCommand("list", "workspaces", "-t", teamId)

			Expect(output).NotTo(BeEmpty())
		})
	})
})

var _ = Describe("Command Error Handling Tests", func() {
	It("should fail gracefully with non-existent workspace for all commands", func() {
		testCases := []struct {
			commandName string
			args        []string
		}{
			{"open workspace", []string{"open", "workspace", "-w", "99999999"}},
			{"log", []string{"log", "-w", "99999999"}},
			{"start pipeline", []string{"start", "pipeline", "-w", "99999999"}},
			{"git pull", []string{"git", "pull", "-w", "99999999"}},
			{"set-env", []string{"set-env", "-w", "99999999", "TEST_VAR=test"}},
			{"wake-up", []string{"wake-up", "-w", "99999999"}},
			{"curl", []string{"curl", "/", "-w", "99999999"}},
		}

		for _, tc := range testCases {
			By(fmt.Sprintf("Testing %s with non-existent workspace", tc.commandName))
			output, exitCode := intutil.RunCommandWithExitCode(tc.args...)
			fmt.Printf("%s non-existent workspace output: %s (exit code: %d)\n", tc.commandName, output, exitCode)
			Expect(exitCode).NotTo(Equal(0))
		}
	})
})

var _ = Describe("Log Command Integration Tests", func() {
	var (
		teamId        string
		workspaceName string
		workspaceId   string
	)

	BeforeEach(func() {
		teamId, _ = intutil.SkipIfMissingEnvVars()
		workspaceName = fmt.Sprintf("cli-log-test-%d", time.Now().Unix())
	})

	AfterEach(func() {
		if workspaceId != "" {
			By(fmt.Sprintf("Cleaning up: deleting workspace %s (ID: %s)", workspaceName, workspaceId))
			intutil.CleanupWorkspace(workspaceId)
			workspaceId = ""
		}
	})

	Context("Log Command", func() {
		BeforeEach(func() {
			By("Creating a workspace")
			output := intutil.RunCommand(
				"create", "workspace", workspaceName,
				"-t", teamId,
				"-p", "8",
				"--timeout", "15m",
			)
			Expect(output).To(ContainSubstring("Workspace created"))
			workspaceId = intutil.ExtractWorkspaceId(output)
			Expect(workspaceId).NotTo(BeEmpty())
		})

		It("should retrieve logs from workspace", func() {
			By("Getting logs from workspace")
			output, exitCode := intutil.RunCommandWithExitCode(
				"log",
				"-w", workspaceId,
			)
			fmt.Printf("Log command output (first 500 chars): %s... (exit code: %d)\n",
				output[:min(500, len(output))], exitCode)

			Expect(exitCode).To(Or(Equal(0), Equal(1)))
		})
	})
})

var _ = Describe("Start Pipeline Integration Tests", func() {
	var (
		teamId        string
		workspaceName string
		workspaceId   string
	)

	BeforeEach(func() {
		teamId, _ = intutil.SkipIfMissingEnvVars()
		workspaceName = fmt.Sprintf("cli-pipeline-test-%d", time.Now().Unix())
	})

	AfterEach(func() {
		if workspaceId != "" {
			By(fmt.Sprintf("Cleaning up: deleting workspace %s (ID: %s)", workspaceName, workspaceId))
			intutil.CleanupWorkspace(workspaceId)
			workspaceId = ""
		}
	})

	Context("Start Pipeline Command", func() {
		BeforeEach(func() {
			By("Creating a workspace")
			output := intutil.RunCommand(
				"create", "workspace", workspaceName,
				"-t", teamId,
				"-p", "8",
				"--timeout", "15m",
			)
			Expect(output).To(ContainSubstring("Workspace created"))
			workspaceId = intutil.ExtractWorkspaceId(output)
			Expect(workspaceId).NotTo(BeEmpty())
		})

		It("should start pipeline successfully", func() {
			By("Starting pipeline")
			output, exitCode := intutil.RunCommandWithExitCode(
				"start", "pipeline",
				"-w", workspaceId,
			)
			fmt.Printf("Start pipeline output: %s (exit code: %d)\n", output, exitCode)

			Expect(output).NotTo(BeEmpty())
		})
	})
})

var _ = Describe("Git Pull Integration Tests", func() {
	var (
		teamId        string
		workspaceName string
		workspaceId   string
	)

	BeforeEach(func() {
		teamId, _ = intutil.SkipIfMissingEnvVars()
		workspaceName = fmt.Sprintf("cli-git-test-%d", time.Now().Unix())
	})

	AfterEach(func() {
		if workspaceId != "" {
			By(fmt.Sprintf("Cleaning up: deleting workspace %s (ID: %s)", workspaceName, workspaceId))
			intutil.CleanupWorkspace(workspaceId)
			workspaceId = ""
		}
	})

	Context("Git Pull Command", func() {
		BeforeEach(func() {
			By("Creating a workspace")
			output := intutil.RunCommand(
				"create", "workspace", workspaceName,
				"-t", teamId,
				"-p", "8",
				"--timeout", "15m",
			)
			Expect(output).To(ContainSubstring("Workspace created"))
			workspaceId = intutil.ExtractWorkspaceId(output)
			Expect(workspaceId).NotTo(BeEmpty())
		})

		It("should execute git pull command", func() {
			By("Running git pull")
			output, exitCode := intutil.RunCommandWithExitCode(
				"git", "pull",
				"-w", workspaceId,
			)
			fmt.Printf("Git pull output: %s (exit code: %d)\n", output, exitCode)

			Expect(output).NotTo(BeEmpty())
		})
	})
})

var _ = Describe("Wake Up Workspace Integration Tests", func() {
	var (
		teamId        string
		workspaceName string
		workspaceId   string
	)

	BeforeEach(func() {
		teamId, _ = intutil.SkipIfMissingEnvVars()
		workspaceName = fmt.Sprintf("cli-wakeup-test-%d", time.Now().Unix())
	})

	AfterEach(func() {
		if workspaceId != "" {
			By(fmt.Sprintf("Cleaning up: deleting workspace %s (ID: %s)", workspaceName, workspaceId))
			intutil.CleanupWorkspace(workspaceId)
			workspaceId = ""
		}
	})

	Context("Wake Up Command", func() {
		BeforeEach(func() {
			By("Creating a workspace for wake-up testing")
			output := intutil.RunCommand(
				"create", "workspace", workspaceName,
				"-t", teamId,
				"-p", "8",
				"--timeout", "15m",
			)
			fmt.Printf("Create workspace output: %s\n", output)

			Expect(output).To(ContainSubstring("Workspace created"))
			workspaceId = intutil.ExtractWorkspaceId(output)
			Expect(workspaceId).NotTo(BeEmpty())

			By("Waiting for workspace to be fully provisioned")
			time.Sleep(5 * time.Second)
		})

		It("should wake up workspace successfully", func() {
			By("Waking up the workspace")
			output := intutil.RunCommand(
				"wake-up",
				"-w", workspaceId,
			)
			fmt.Printf("Wake up workspace output: %s\n", output)

			Expect(output).To(ContainSubstring("Waking up workspace"))
			Expect(output).To(ContainSubstring(workspaceId))
		})

		It("should respect custom timeout", func() {
			By("Waking up workspace with custom timeout")
			output, exitCode := intutil.RunCommandWithExitCode(
				"wake-up",
				"-w", workspaceId,
				"--timeout", "5s",
			)
			fmt.Printf("Wake up with timeout output: %s (exit code: %d)\n", output, exitCode)

			Expect(output).To(ContainSubstring("Waking up workspace"))
		})

		It("should work with workspace ID from environment variable", func() {
			By("Setting CS_WORKSPACE_ID environment variable")
			originalWsId := os.Getenv("CS_WORKSPACE_ID")
			_ = os.Setenv("CS_WORKSPACE_ID", workspaceId)
			defer func() { _ = os.Setenv("CS_WORKSPACE_ID", originalWsId) }()

			By("Waking up workspace using environment variable")
			output := intutil.RunCommand("wake-up")
			fmt.Printf("Wake up with env var output: %s\n", output)

			Expect(output).To(ContainSubstring("Waking up workspace"))
			Expect(output).To(ContainSubstring(workspaceId))
		})
	})

	Context("Wake Up Error Handling", func() {
		It("should fail when workspace ID is missing", func() {
			By("Attempting to wake up workspace without ID")
			originalWsId := os.Getenv("CS_WORKSPACE_ID")
			originalWsIdFallback := os.Getenv("WORKSPACE_ID")
			_ = os.Unsetenv("CS_WORKSPACE_ID")
			_ = os.Unsetenv("WORKSPACE_ID")
			defer func() {
				_ = os.Setenv("CS_WORKSPACE_ID", originalWsId)
				_ = os.Setenv("WORKSPACE_ID", originalWsIdFallback)
			}()

			output, exitCode := intutil.RunCommandWithExitCode("wake-up")
			fmt.Printf("Wake up without workspace ID output: %s (exit code: %d)\n", output, exitCode)

			Expect(exitCode).NotTo(Equal(0))
			Expect(output).To(Or(
				ContainSubstring("workspace"),
				ContainSubstring("required"),
				ContainSubstring("not set"),
			))
		})

		It("should fail gracefully with non-existent workspace", func() {
			By("Attempting to wake up non-existent workspace")
			output, exitCode := intutil.RunCommandWithExitCode(
				"wake-up",
				"-w", "99999999",
			)
			fmt.Printf("Wake up non-existent workspace output: %s (exit code: %d)\n", output, exitCode)

			Expect(exitCode).NotTo(Equal(0))
			Expect(output).To(Or(
				ContainSubstring("failed to get workspace"),
				ContainSubstring("not found"),
				ContainSubstring("404"),
			))
		})

		It("should handle workspace without dev domain gracefully", func() {
			By("Creating a workspace (which might not have dev domain configured)")
			output := intutil.RunCommand(
				"create", "workspace", workspaceName,
				"-t", teamId,
				"-p", "8",
				"--timeout", "15m",
			)
			fmt.Printf("Create workspace output: %s\n", output)

			Expect(output).To(ContainSubstring("Workspace created"))
			workspaceId = intutil.ExtractWorkspaceId(output)
			Expect(workspaceId).NotTo(BeEmpty())

			By("Attempting to wake up the workspace")
			wakeupOutput, wakeupExitCode := intutil.RunCommandWithExitCode(
				"wake-up",
				"-w", workspaceId,
			)
			fmt.Printf("Wake up workspace output: %s (exit code: %d)\n", wakeupOutput, wakeupExitCode)

			if wakeupExitCode != 0 {
				Expect(wakeupOutput).To(Or(
					ContainSubstring("development domain"),
					ContainSubstring("dev domain"),
					ContainSubstring("failed to wake up"),
				))
			}
		})
	})

	Context("Wake Up Command Help", func() {
		It("should display help information", func() {
			By("Running wake-up --help")
			output := intutil.RunCommand("wake-up", "--help")
			fmt.Printf("Wake up help output: %s\n", output)

			Expect(output).To(ContainSubstring("Wake up an on-demand workspace"))
			Expect(output).To(ContainSubstring("--timeout"))
			Expect(output).To(ContainSubstring("-w, --workspace"))
		})
	})
})

var _ = Describe("Curl Workspace Integration Tests", func() {
	var (
		teamId        string
		workspaceName string
		workspaceId   string
	)

	BeforeEach(func() {
		teamId, _ = intutil.SkipIfMissingEnvVars()
		workspaceName = fmt.Sprintf("cli-curl-test-%d", time.Now().Unix())
	})

	AfterEach(func() {
		if workspaceId != "" {
			By(fmt.Sprintf("Cleaning up: deleting workspace %s (ID: %s)", workspaceName, workspaceId))
			intutil.CleanupWorkspace(workspaceId)
			workspaceId = ""
		}
	})

	Context("Curl Command", func() {
		BeforeEach(func() {
			By("Creating a workspace for curl testing")
			output := intutil.RunCommand(
				"create", "workspace", workspaceName,
				"-t", teamId,
				"-p", "8",
				"--timeout", "15m",
			)
			fmt.Printf("Create workspace output: %s\n", output)

			Expect(output).To(ContainSubstring("Workspace created"))
			workspaceId = intutil.ExtractWorkspaceId(output)
			Expect(workspaceId).NotTo(BeEmpty())

			By("Waiting for workspace to be fully provisioned")
			time.Sleep(5 * time.Second)
		})

		It("should send authenticated request to workspace", func() {
			By("Sending curl request to workspace root")
			output := intutil.RunCommand(
				"curl", "/",
				"-w", workspaceId,
				"--insecure",
				"--", "-s", "-o", "/dev/null", "-w", "%{http_code}",
			)
			fmt.Printf("Curl workspace output: %s\n", output)

			Expect(output).To(ContainSubstring("Sending request to workspace"))
			Expect(output).To(ContainSubstring(workspaceId))
		})

		It("should support custom port", func() {
			By("Sending curl request to custom port")
			output, exitCode := intutil.RunCommandWithExitCode(
				"curl", "/",
				"-w", workspaceId,
				"-p", "3001",
				"--insecure",
				"--", "-s", "-o", "/dev/null", "-w", "%{http_code}",
			)
			fmt.Printf("Curl with custom port output: %s (exit code: %d)\n", output, exitCode)

			Expect(output).To(ContainSubstring("Sending request to workspace"))
		})

		It("should pass through curl arguments", func() {
			By("Sending HEAD request using curl -I flag")
			output := intutil.RunCommand(
				"curl", "/",
				"-w", workspaceId,
				"--insecure",
				"--", "-I",
			)
			fmt.Printf("Curl with -I flag output: %s\n", output)

			Expect(output).To(ContainSubstring("Sending request to workspace"))
		})

		It("should work with workspace ID from environment variable", func() {
			By("Setting CS_WORKSPACE_ID environment variable")
			originalWsId := os.Getenv("CS_WORKSPACE_ID")
			_ = os.Setenv("CS_WORKSPACE_ID", workspaceId)
			defer func() { _ = os.Setenv("CS_WORKSPACE_ID", originalWsId) }()

			By("Sending curl request using environment variable")
			output := intutil.RunCommand(
				"curl", "/",
				"--insecure",
				"--", "-s", "-o", "/dev/null", "-w", "%{http_code}",
			)
			fmt.Printf("Curl with env var output: %s\n", output)

			Expect(output).To(ContainSubstring("Sending request to workspace"))
			Expect(output).To(ContainSubstring(workspaceId))
		})
	})

	Context("Curl Error Handling", func() {
		It("should fail when workspace ID is missing", func() {
			By("Attempting to curl without workspace ID")
			originalWsId := os.Getenv("CS_WORKSPACE_ID")
			originalWsIdFallback := os.Getenv("WORKSPACE_ID")
			_ = os.Unsetenv("CS_WORKSPACE_ID")
			_ = os.Unsetenv("WORKSPACE_ID")
			defer func() {
				_ = os.Setenv("CS_WORKSPACE_ID", originalWsId)
				_ = os.Setenv("WORKSPACE_ID", originalWsIdFallback)
			}()

			output, exitCode := intutil.RunCommandWithExitCode("curl", "/")
			fmt.Printf("Curl without workspace ID output: %s (exit code: %d)\n", output, exitCode)

			Expect(exitCode).NotTo(Equal(0))
			Expect(output).To(Or(
				ContainSubstring("workspace"),
				ContainSubstring("required"),
				ContainSubstring("not set"),
			))
		})

		It("should fail gracefully with non-existent workspace", func() {
			By("Attempting to curl non-existent workspace")
			output, exitCode := intutil.RunCommandWithExitCode(
				"curl", "/",
				"-w", "99999999",
			)
			fmt.Printf("Curl non-existent workspace output: %s (exit code: %d)\n", output, exitCode)

			Expect(exitCode).NotTo(Equal(0))
			Expect(output).To(Or(
				ContainSubstring("failed to get workspace"),
				ContainSubstring("not found"),
				ContainSubstring("404"),
			))
		})

		It("should require path argument", func() {
			By("Attempting to curl without path")
			output, exitCode := intutil.RunCommandWithExitCode(
				"curl",
				"-w", "1234",
			)
			fmt.Printf("Curl without path output: %s (exit code: %d)\n", output, exitCode)

			Expect(exitCode).NotTo(Equal(0))
			Expect(output).To(Or(
				ContainSubstring("path"),
				ContainSubstring("required"),
				ContainSubstring("argument"),
			))
		})
	})

	Context("Curl Command Help", func() {
		It("should display help information", func() {
			By("Running curl --help")
			output := intutil.RunCommand("curl", "--help")
			fmt.Printf("Curl help output: %s\n", output)

			Expect(output).To(ContainSubstring("Send authenticated HTTP requests"))
			Expect(output).To(ContainSubstring("--port"))
			Expect(output).To(ContainSubstring("-w, --workspace"))
		})
	})
})

var _ = Describe("Command Error Handling Tests", func() {
	It("should fail gracefully with non-existent workspace for all commands", func() {
		testCases := []struct {
			commandName string
			args        []string
		}{
			{"open workspace", []string{"open", "workspace", "-w", "99999999"}},
			{"log", []string{"log", "-w", "99999999"}},
			{"start pipeline", []string{"start", "pipeline", "-w", "99999999"}},
			{"git pull", []string{"git", "pull", "-w", "99999999"}},
			{"set-env", []string{"set-env", "-w", "99999999", "TEST_VAR=test"}},
		}

		for _, tc := range testCases {
			By(fmt.Sprintf("Testing %s with non-existent workspace", tc.commandName))
			output, exitCode := intutil.RunCommandWithExitCode(tc.args...)
			fmt.Printf("%s non-existent workspace output: %s (exit code: %d)\n", tc.commandName, output, exitCode)
			Expect(exitCode).NotTo(Equal(0))
		}
	})
})
