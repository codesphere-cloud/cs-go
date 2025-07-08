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
		// Setup temporary directory for test artifacts
		tempDir, err = os.MkdirTemp("", "e2e-tls-monitor-test-")
		Expect(err).NotTo(HaveOccurred())
		certsDir = filepath.Join(tempDir, "certs")

		// Get ephemeral ports for the monitor proxy and the Go HTTPS server
		monitorListenPort, err = intutil.GetEphemeralPort()
		Expect(err).NotTo(HaveOccurred())
		targetServerPort, err = intutil.GetEphemeralPort()
		Expect(err).NotTo(HaveOccurred())

		// Initialize a standard HTTP client for testing the monitor proxy.
		testHttpClient = &http.Client{
			Timeout: 10 * time.Second,
		}

		// Buffers to capture stdout/stderr of the processes
		monitorOutputBuf = new(bytes.Buffer)
		targetServerOutputBuf = new(bytes.Buffer)
	})

	AfterEach(func() {
		// Terminate monitor process
		if monitorCmdProcess != nil && monitorCmdProcess.Process != nil {
			fmt.Printf("Terminating monitor process (PID: %d). Output:\n%s\n", monitorCmdProcess.Process.Pid, monitorOutputBuf.String())
			_ = monitorCmdProcess.Process.Kill()
			_, _ = monitorCmdProcess.Process.Wait()
		}

		// Clean up temporary directory
		Expect(os.RemoveAll(tempDir)).NotTo(HaveOccurred())
	})

	Context("Healthcheck forwarding", func() {
		AfterEach(func() {
			// Terminate HTTP server process
			if targetServer != nil {
				fmt.Printf("Terminating HTTP(S) server. Output:\n%s\n", targetServerOutputBuf.String())
				shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 1*time.Second)
				defer shutdownCancel()
				_ = targetServer.Shutdown(shutdownCtx) // Ignore error, it might already be closed
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
				// No --insecure-skip-verify
				// No --ca-cert-file
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
				"--ca-cert-file", caCertPath, // Provide the CA cert
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
			// Disable auto-redirect for the client to check the redirect status
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
				"--", "true", // Command that succeeds immediately
			)

			Eventually(monitorCmdProcess.Wait, "5s").Should(Succeed(), "Monitor process should exit successfully")

			// Verify output indicates single execution and exit
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
				"--", "bash", "-c", "echo FAKE_OUTPUT;exit 1", // Command that fails immediately
			)

			Eventually(monitorCmdProcess.Wait, "15s").Should(Succeed(), "Monitor process should exit after restarts")

			// Verify output indicates restarts and exit
			output := monitorOutputBuf.String()
			Expect(output).To(ContainSubstring("command exited"))
			Expect(output).To(ContainSubstring("returnCode=1"))
			Expect(output).To(ContainSubstring("command exited with non-zero code in less than 1 second. Waiting 5 seconds before next restart"))
			Expect(output).To(ContainSubstring("cs monitor: restarting"))
			Expect(output).To(ContainSubstring("maximum number of restarts reached, exiting"))
			Expect(strings.Count(output, "FAKE_OUTPUT")).To(Equal(3), "Command should have executed twice") // 3 because the command is printed once
		})

		It("should stop command runner on context cancellation", func() {
			By("Running 'cs monitor' command with infinite restarts")
			monitorCmdProcess = intutil.RunCommandInBackground(monitorOutputBuf,
				"monitor",
				"--address", fmt.Sprintf(":%d", monitorListenPort),
				"--max-restarts", "-1", // Infinite
				"--", "sleep", "10s", // A command that runs for a while
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
