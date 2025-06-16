package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	csio "github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var (
	// Metrics
	lastRestartTimestamp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cs_monitor_last_restart_timestamp_seconds",
			Help: "Timestamp of the last command execution completion.",
		},
		[]string{"return_code"},
	)
	totalRestarts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cs_monitor_total_restarts_total",
			Help: "Total number of command executions completed.",
		},
		[]string{"return_code"},
	)
)

// RunCommandWithMetrics continuously runs the given command, streams its output,
// and reports Prometheus metrics. It implements a restart delay of 5 seconds
// if a non-zero exit code occurs within 1 second of the command starting.
// The metrics will be served on the specified listenAddr.
func (c *MonitorCmd) RunCommandWithMetrics(ctx context.Context, cmdArgs []string) error {
	if len(cmdArgs) == 0 {
		return errors.New("no command specified")
	}

	fmt.Printf("Starting monitor with command: %v\n", cmdArgs)
	customRegistry := prometheus.NewRegistry()
	customRegistry.MustRegister(lastRestartTimestamp)
	customRegistry.MustRegister(totalRestarts)

	// Setup HTTP server for Prometheus metrics
	http.Handle("/metrics", promhttp.HandlerFor(customRegistry, promhttp.HandlerOpts{}))

	// Add redirect handler for "/" to "/metrics"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/metrics", http.StatusMovedPermanently)
	})

	go func() {
		fmt.Printf("Prometheus metrics will be available on %s/metrics\n", *c.opts.ListenAddress)
		err := http.ListenAndServe(*c.opts.ListenAddress, nil)
		if err != nil {
			panic(fmt.Errorf("error starting metrics server on %s: %v", *c.opts.ListenAddress, err))
		}
	}()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Context cancelled, stopping command runner.")
			return nil
		default:
			// Initialize metric with start time
			lastRestartTimestamp.WithLabelValues("").Set(float64(time.Now().Unix()))
			fmt.Printf("Running command: %s\n", cmdArgs)
			startTime := time.Now()
			returnCode := executeCommand(ctx, cmdArgs)
			duration := time.Since(startTime)
			strReturnCode := strconv.Itoa(returnCode)

			// Update Prometheus metrics
			lastRestartTimestamp.WithLabelValues(strReturnCode).Set(float64(time.Now().Unix()))
			totalRestarts.WithLabelValues(strReturnCode).Inc()

			fmt.Printf("Command finished with return code: %d (took %v)\n", returnCode, duration)

			// Implement the restart delay logic
			if returnCode != 0 && duration < 1*time.Second {
				fmt.Printf("Command exited with non-zero code (%d) in less than 1 second (%v). Waiting 5 seconds before next restart...\n", returnCode, duration)
				select {
				case <-time.After(5 * time.Second):
				case <-ctx.Done():
					fmt.Println("Context cancelled during delay, stopping command runner.")
					return nil
				}
			}
			fmt.Println("cs monitor: restarting.")
		}
	}
}

// executeCommand executes the given command, streams its output, and returns its exit code.
func executeCommand(ctx context.Context, cmdArgs []string) int {
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)

	// Create pipes for stdout and stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Error creating stdout pipe: %v\n", err)
		return -1 // Indicate an internal error
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("Error creating stderr pipe: %v\n", err)
		return -1 // Indicate an internal error
	}

	// Use a WaitGroup to ensure all goroutines finish before returning
	var wg sync.WaitGroup

	// Stream stdout
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		if err := scanner.Err(); err != nil && err != io.EOF { // io.EOF is expected at end of stream
			fmt.Printf("Error reading stdout: %v\n", err)
		}
	}()

	// Stream stderr
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			fmt.Fprintln(os.Stderr, scanner.Text())
		}
		if err := scanner.Err(); err != nil && err != io.EOF { // io.EOF is expected at end of stream
			fmt.Printf("Error reading stderr: %v\n", err)
		}
	}()

	// Start the command
	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting command %v: %v\n", cmdArgs, err)
		return -1 // Indicate an internal error
	}

	// Wait for the command to finish
	err = cmd.Wait()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// The command exited with a non-zero status
			return exitError.ExitCode()
		}
		// Other errors (e.g., command not found, context cancelled before start)
		fmt.Printf("Command %v failed with error: %v\n", cmdArgs, err)
		return -1 // Indicate an internal error or command failure
	}

	wg.Wait() // Wait for stdout/stderr streaming to complete
	return 0
}

type MonitorCmd struct {
	cmd  *cobra.Command
	opts MonitorOpts
}

type MonitorOpts struct {
	GlobalOptions
	ListenAddress *string
}

func (c *MonitorCmd) RunE(_ *cobra.Command, args []string) error {
	// Allow graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		s := <-sigChan
		fmt.Printf("Received signal %s, initiating graceful shutdown...\n", s)
		cancel()
	}()

	return c.RunCommandWithMetrics(ctx, args)
}

func AddMonitorCmd(rootCmd *cobra.Command) {
	monitor := MonitorCmd{
		cmd: &cobra.Command{
			Use:   "monitor",
			Short: "Monitor a command and report health information",
			Long: csio.Long(`Loops over running a command and report information in a health endpoint.

				Codesphere watches for health information of an application on port 3000, which is the default endpoint for this command.
				You can specify a different port if your application is running on port 3000.

				The monitor command keeps restarting an application and reports the metrics about the restarts in prometheus metrics format.
				Metrics reported are
				* cs_monitor_last_restart_timestamp_seconds - Timestamp of the last command execution completion
				* cs_monitor_total_restarts_total - Total number of command executions completed`),
			Example: csio.FormatExampleCommands("monitor", []csio.Example{
				{Cmd: "-- npm start", Desc: "monitor application that ist started by npm"},
				{Cmd: "--address 8080 -- ./my-app -p 3000 ", Desc: "monitor application from local binary on port 3000, expose metrics on port 8080"},
			}),
		},
	}
	monitor.opts.ListenAddress = monitor.cmd.Flags().String("address", ":3000", "Help message for toggle")
	rootCmd.AddCommand(monitor.cmd)
	monitor.cmd.RunE = monitor.RunE
}
