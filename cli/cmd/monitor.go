// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
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
	cmd  *cobra.Command
	Opts MonitorOpts
	Time api.Time
	Http csio.HttpServer
	Exec csio.Exec
}

type MonitorOpts struct {
	GlobalOptions
	ListenAddress *string
	MaxRestarts   *int
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

	return c.RunCommandWithMetrics(ctx, args)
}

// RunCommandWithMetrics continuously runs the given command, streams its output,
// and reports Prometheus metrics. It implements a restart delay of 5 seconds
// if a non-zero exit code occurs within 1 second of the command starting.
// The metrics will be served on the specified listenAddr.
func (c *MonitorCmd) RunCommandWithMetrics(ctx context.Context, cmdArgs []string) error {
	if len(cmdArgs) == 0 {
		return errors.New("no command specified")
	}

	slog.Info("starting monitor", "command", strings.Join(cmdArgs, " "))

	totalRestartsMetric := c.startHealthcheckEndpoint()

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

			totalRestartsMetric.WithLabelValues(strReturnCode).Inc()

			slog.Info("command exited", "return code", returnCode, "duration", duration)

			if *c.Opts.MaxRestarts < (i + 1) {
				slog.Info("maximum number of restarts reached, exiting.")
				break
			}
			// Delay in case of fast non-zero exit
			if returnCode != 0 && duration < 1*time.Second {
				slog.Info("command exited with non-zero code in less than 1 second. Waiting 5 seconds before next restart", "return code", returnCode, "command duration", duration)
				c.Time.Sleep(5 * time.Second)
			}
			slog.Info("cs monitor: restarting.")
		}
	}
	return nil
}

func (c *MonitorCmd) startHealthcheckEndpoint() *prometheus.CounterVec {
	totalRestarts := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cs_monitor_restarts_total",
			Help: "Total number of command executions completed per return code. Empty return_code count is initialized with 1.",
		},
		[]string{"return_code"},
	)

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
	return totalRestarts
}

func AddMonitorCmd(rootCmd *cobra.Command, opts GlobalOptions) {
	monitor := MonitorCmd{
		cmd: &cobra.Command{
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
			}),
		},
		Time: &api.RealTime{},
		Opts: MonitorOpts{GlobalOptions: opts},
		Http: &csio.RealHttpServer{},
		Exec: &csio.RealExec{},
	}
	monitor.Opts.ListenAddress = monitor.cmd.Flags().String("address", ":3000", "Custom listen address for the metrics endpoint")
	monitor.Opts.MaxRestarts = monitor.cmd.Flags().Int("max-restarts", -1, "Maximum number of restarts before exiting")
	rootCmd.AddCommand(monitor.cmd)
	monitor.cmd.RunE = monitor.RunE
}
