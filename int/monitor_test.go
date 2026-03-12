// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package int_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
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

var _ = Describe("cs monitor", Label("local"), func() {
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
			log.Printf("Terminating monitor process (PID: %d). Output:\n%s\n", monitorCmdProcess.Process.Pid, monitorOutputBuf.String())
			_ = monitorCmdProcess.Process.Kill()
			_, _ = monitorCmdProcess.Process.Wait()
		}

		Expect(os.RemoveAll(tempDir)).NotTo(HaveOccurred())
	})

	Context("Healthcheck forwarding", func() {
		AfterEach(func() {
			if targetServer != nil {
				log.Printf("Terminating HTTP(S) server. Output:\n%s\n", targetServerOutputBuf.String())
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
			log.Printf("Go HTTPS server started on port %d.\n", targetServerPort)

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

			log.Printf("Monitor output after request:\n%s\n", monitorOutputBuf.String())
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
			log.Printf("Go HTTPS server started on port %d.\n", targetServerPort)

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

			log.Printf("Monitor output after request:\n%s\n", monitorOutputBuf.String())
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
			log.Printf("Go HTTPS server started on port %d.\n", targetServerPort)

			By("Running 'cs monitor' command without TLS bypass/trust")
			intutil.RunCommandInBackground(
				monitorOutputBuf,
				"monitor",
				"--address", fmt.Sprintf(":%d", monitorListenPort),
				"--forward", fmt.Sprintf("https://127.0.0.1:%d/", targetServerPort),
				"--", "sleep", "60s",
			)

			intutil.WaitForPort(fmt.Sprintf("127.0.0.1:%d", monitorListenPort), 10*time.Second)
			log.Printf("Monitor command started on port %d. Initial output:\n%s\n", monitorListenPort, monitorOutputBuf.String())

			By("Making request to monitor proxy and expecting a Bad Gateway error")
			resp, err := testHttpClient.Get(fmt.Sprintf("http://127.0.0.1:%d/", monitorListenPort))
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusBadGateway))
			bodyBytes, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bodyBytes)).To(ContainSubstring("Error forwarding request"))
			Expect(string(bodyBytes)).To(ContainSubstring("tls: failed to verify certificate"))

			log.Printf("Monitor output after request:\n%s\n", monitorOutputBuf.String())
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
			log.Printf("Go HTTPS server started on port %d.\n", targetServerPort)

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
			log.Printf("Monitor command started on port %d. Initial output:\n%s\n", monitorListenPort, monitorOutputBuf.String())

			By("Making request to monitor proxy to verify successful forwarding")
			resp, err := testHttpClient.Get(fmt.Sprintf("http://127.0.0.1:%d/", monitorListenPort))
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			bodyBytes, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bodyBytes)).To(Equal("OK (HTTPS)"))

			log.Printf("Monitor output after request:\n%s\n", monitorOutputBuf.String())
		})
	})

	Context("Prometheus Metrics Endpoint", func() {
		It("should expose Prometheus metrics when no forward is specified", func() {
			By("Running 'cs monitor' command without forwarding (metrics only)")
			monitorCmdProcess = intutil.RunCommandInBackground(monitorOutputBuf,
				"monitor",
				"--address", fmt.Sprintf(":%d", monitorListenPort),
				"--", "sleep", "60s",
			)
			intutil.WaitForPort(fmt.Sprintf("127.0.0.1:%d", monitorListenPort), 10*time.Second)
			log.Printf("Monitor command started on port %d. Initial output:\n%s\n", monitorListenPort, monitorOutputBuf.String())

			By("Making a request to the monitor's metrics endpoint")
			resp, err := testHttpClient.Get(fmt.Sprintf("http://127.0.0.1:%d/metrics", monitorListenPort))
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			bodyBytes, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bodyBytes)).To(ContainSubstring("cs_monitor_restarts_total"))
			log.Printf("Monitor output after metrics request:\n%s\n", monitorOutputBuf.String())
		})

		It("should redirect root to /metrics", func() {
			By("Running 'cs monitor' command without forwarding (metrics only)")
			monitorCmdProcess = intutil.RunCommandInBackground(monitorOutputBuf,
				"monitor",
				"--address", fmt.Sprintf(":%d", monitorListenPort),
				"--", "sleep", "60s",
			)
			intutil.WaitForPort(fmt.Sprintf("127.0.0.1:%d", monitorListenPort), 10*time.Second)
			log.Printf("Monitor command started on port %d. Initial output:\n%s\n", monitorListenPort, monitorOutputBuf.String())

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
			log.Printf("Monitor output after redirect request:\n%s\n", monitorOutputBuf.String())
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
