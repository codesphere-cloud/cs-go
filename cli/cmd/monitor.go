// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/codesphere-cloud/cs-go/api"
	csio "github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

type MonitorCmd struct {
	Cmd  *cobra.Command
	Opts MonitorOpts
	Time api.Time
	Http csio.HttpServer
	Exec csio.Exec
}

type MonitorOpts struct {
	GlobalOptions
	ListenAddress      *string
	MaxRestarts        *int
	Forward            *string
	CaCertFile         *string
	InsecureSkipVerify *bool
}

func (c *MonitorCmd) RunE(_ *cobra.Command, args []string) error {
	// Allow graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		s := <-sigChan
		slog.Info("initiating graceful shutdown...", "signal", s)
		cancel()
	}()

	return c.RunCommandWithHealthcheck(ctx, args)
}

// RunCommandWithMetrics continuously runs the given command, streams its output,
// and reports Prometheus metrics. It implements a restart delay of 5 seconds
// if a non-zero exit code occurs within 1 second of the command starting.
// The metrics will be served on the specified listenAddr.
func (c *MonitorCmd) RunCommandWithHealthcheck(ctx context.Context, cmdArgs []string) error {
	if len(cmdArgs) == 0 {
		return errors.New("no command specified")
	}

	slog.Info("starting monitor", "command", strings.Join(cmdArgs, " "))

	totalRestarts := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cs_monitor_restarts_total",
			Help: "Total number of command executions completed per return code. Empty return_code count is initialized with 1.",
		},
		[]string{"return_code"},
	)

	c.startHealthcheckEndpoint(totalRestarts)

	//MaxRestarts required for being able to unit test
	for i := 0; i <= *c.Opts.MaxRestarts || *c.Opts.MaxRestarts == -1; i++ {
		select {
		case <-ctx.Done():
			slog.Info("stopping command runner.")
			return nil
		default:
			startTime := c.Time.Now()
			returnCode, err := c.Exec.ExecuteCommand(ctx, cmdArgs)
			if err != nil {
				return fmt.Errorf("error executing command %s: %w", cmdArgs, err)
			}
			duration := c.Time.Now().Sub(startTime)
			strReturnCode := strconv.Itoa(returnCode)

			totalRestarts.WithLabelValues(strReturnCode).Inc()

			slog.Info("command exited", "returnCode", returnCode, "duration", duration)

			if *c.Opts.MaxRestarts >= 0 && *c.Opts.MaxRestarts < (i+1) {
				slog.Info("maximum number of restarts reached, exiting.")
				return nil
			}
			// Delay in case of fast non-zero exit
			if returnCode != 0 && duration < 1*time.Second {
				slog.Info("command exited with non-zero code in less than 1 second. Waiting 5 seconds before next restart", "returnCode", returnCode, "commandDuration", duration)
				c.Time.Sleep(5 * time.Second)
			}
			slog.Info("cs monitor: restarting.")
		}
	}
	return nil
}

func (c *MonitorCmd) startHealthcheckEndpoint(totalRestarts *prometheus.CounterVec) {

	if c.Opts.Forward != nil && *c.Opts.Forward != "" {
		c.startHttpProxy()
		return
	}

	c.startMetricsEndpoint(totalRestarts)
}

func (c *MonitorCmd) startMetricsEndpoint(totalRestarts *prometheus.CounterVec) {
	customRegistry := prometheus.NewRegistry()
	customRegistry.MustRegister(totalRestarts)

	// Initialize metric
	totalRestarts.WithLabelValues("").Inc()

	c.Http.Handle("/metrics", promhttp.HandlerFor(customRegistry, promhttp.HandlerOpts{}))

	c.Http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c.Http.Redirect(w, r, "/metrics", http.StatusMovedPermanently)
	})

	go func() {
		slog.Info("starting prometheus metrics", "listen address", *c.Opts.ListenAddress)
		err := c.Http.ListenAndServe(*c.Opts.ListenAddress, nil)
		if err != nil {
			panic(fmt.Errorf("error starting metrics server on %s: %v", *c.Opts.ListenAddress, err))
		}
	}()
}

func (c *MonitorCmd) startHttpProxy() {
	proxy, err := c.NewProxy()
	if err != nil {
		panic(fmt.Errorf("failed to create proxy: %w", err))
	}

	slog.Info("HTTP proxy listening", "listen address", *c.Opts.ListenAddress, "targetURL", proxy.targetURL.String())

	go func() {
		err := c.Http.ListenAndServe(*c.Opts.ListenAddress, proxy)
		if err != nil {
			panic(fmt.Errorf("error starting proxy server: %w", err))
		}
	}()
}

// Proxy represents our HTTP proxy server with a fixed target URL.
type Proxy struct {
	targetURL  *url.URL
	httpClient *http.Client
}

// NewProxy creates and returns a new Proxy instance.
func (c *MonitorCmd) NewProxy() (*Proxy, error) {
	parsedURL, err := url.Parse(*c.Opts.Forward)
	if err != nil {
		return nil, fmt.Errorf("error parsing target URL '%s': %w", *c.Opts.Forward, err)
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("target URL '%s' must include a scheme (http/https) and host", *c.Opts.Forward)
	}

	// Configure TLS for the outgoing client
	tlsConfig := &tls.Config{
		InsecureSkipVerify: *c.Opts.InsecureSkipVerify,
	}

	// Load custom CA certificate to trust the target server's certificate
	if c.Opts.CaCertFile != nil && *c.Opts.CaCertFile != "" {
		caCert, err := os.ReadFile(*c.Opts.CaCertFile)
		if err != nil {
			return nil, fmt.Errorf("error reading CA certificate file '%s': %w", *c.Opts.CaCertFile, err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate from '%s'", *c.Opts.CaCertFile)
		}
		tlsConfig.RootCAs = caCertPool
		slog.Info("Loaded custom CA certificate to trust target server", "caFile", c.Opts.CaCertFile)
	}

	// Create a custom Transport to use our TLS configuration
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
		Proxy:           http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// Create the HTTP client with the custom transport
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}

	return &Proxy{targetURL: parsedURL, httpClient: httpClient}, nil
}

// ServeHTTP implements the http.Handler interface for our Proxy.
// This method will be called for every incoming HTTP request.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Scheme = p.targetURL.Scheme
	r.URL.Host = p.targetURL.Host

	// Append the client's path
	r.URL.Path = p.targetURL.Path + r.URL.Path

	// If the target URL itself has a query, use it. Otherwise, keep the client's query.
	if p.targetURL.RawQuery != "" {
		r.URL.RawQuery = p.targetURL.RawQuery
	}
	r.RequestURI = "" // RequestURI must be empty when acting as a client.

	// Remove hop-by-hop headers that shouldn't be forwarded.
	// These headers are for the connection between the client and the proxy,
	// not for the connection between the proxy and the target.
	r.Header.Del("Proxy-Connection")
	r.Header.Del("Keep-Alive")
	r.Header.Del("Te")         // Transfer-Encoding
	r.Header.Del("Trailers")   // Trailer header
	r.Header.Del("Upgrade")    // Upgrade header
	r.Header.Del("Connection") // Remove Connection header as well

	resp, err := p.httpClient.Do(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error forwarding request to %s: %v", p.targetURL, err), http.StatusBadGateway)
		slog.Error("Error forwarding request", "targetURL", p.targetURL.String(), "error", err)
		return
	}

	// Copy the response headers from the target server to our client.
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Set the status code.
	w.WriteHeader(resp.StatusCode)

	// Copy the response body from the target server to our client.
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		// Note: At this point, headers have been sent, so we can't
		// send an HTTP error code back. We just log the issue.
		slog.Error("Error copying response body for %s: %v", p.targetURL.String(), err)
	}
	_ = resp.Body.Close()
}

func AddMonitorCmd(rootCmd *cobra.Command, opts GlobalOptions) {
	monitor := MonitorCmd{
		Cmd: &cobra.Command{
			Use:   "monitor",
			Short: "Monitor a command and report health information",
			Long: csio.Long(`Loops over running a command and report information in a health endpoint.

				Codesphere watches for health information of an application on port 3000, which is the default endpoint for this command.
				You can specify a different port if your application is running on port 3000.

				The monitor command keeps restarting an application and reports the metrics about the restarts in prometheus metrics format.
				Metrics reported are
				* cs_monitor_total_restarts_total - Total number of command executions completed`),
			Example: csio.FormatExampleCommands("monitor", []csio.Example{
				{Cmd: "-- npm start", Desc: "monitor application that ist started by npm"},
				{Cmd: "--address 8080 -- ./my-app -p 3000 ", Desc: "monitor application from local binary on port 3000, expose metrics on port 8080"},
				{Cmd: "--forward http://localhost:8080/my-healthcheck -- ./my-app --healthcheck :8080", Desc: "forward health-check to application health endpoint"},
				{Cmd: "--forward --insecure-skip-verify -- ./my-app --healthcheck https://localhost:8443", Desc: "forward health-check to application health endpoint, ignore invalid TLS certs"},
				{Cmd: "--forward --ca-cert-file ca.crt -- ./my-app --healthcheck https://localhost:8443", Desc: "forward health-check to application health endpoint, using custom CA cert, e.g. for self-signed certs"},
			}),
		},
		Time: &api.RealTime{},
		Opts: MonitorOpts{GlobalOptions: opts},
		Http: &csio.RealHttpServer{},
		Exec: &csio.RealExec{},
	}
	monitor.Opts.ListenAddress = monitor.Cmd.Flags().String("address", ":3000", "Custom listen address for the metrics endpoint")
	monitor.Opts.MaxRestarts = monitor.Cmd.Flags().Int("max-restarts", -1, "Maximum number of restarts before exiting")
	monitor.Opts.Forward = monitor.Cmd.Flags().String("forward", "", "Forward healthcheck requests to application health endpoint")
	monitor.Opts.InsecureSkipVerify = monitor.Cmd.Flags().Bool("insecure-skip-verify", false, "Skip TLS validation (only relevant for --forward option when healthcheck is exposed as HTTPS endpoint with custom certificate)")
	monitor.Opts.CaCertFile = monitor.Cmd.Flags().String("ca-cert-file", "", "TLS CA certificate (only relevant for --forward option when healthcheck is exposed as HTTPS enpoint with custom certificate)")
	rootCmd.AddCommand(monitor.Cmd)
	monitor.Cmd.RunE = monitor.RunE
}
