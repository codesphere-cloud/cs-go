package cmd

import (
	"context"
	"errors"
	"fmt"
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
	Http HttpServer
}

type MonitorOpts struct {
	GlobalOptions
	ListenAddress *string
	MaxRestarts   *int
}

type HttpServer interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	Handle(pattern string, handler http.Handler)
	Redirect(w http.ResponseWriter, r *http.Request, url string, code int)
}

type RealHttpServer struct{}

func (*RealHttpServer) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, handler)
}

func (*RealHttpServer) Handle(pattern string, handler http.Handler) {
	http.Handle(pattern, handler)
}
func (*RealHttpServer) Redirect(w http.ResponseWriter, r *http.Request, url string, code int) {
	http.Redirect(w, r, url, code)
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

// RunCommandWithMetrics continuously runs the given command, streams its output,
// and reports Prometheus metrics. It implements a restart delay of 5 seconds
// if a non-zero exit code occurs within 1 second of the command starting.
// The metrics will be served on the specified listenAddr.
func (c *MonitorCmd) RunCommandWithMetrics(ctx context.Context, cmdArgs []string) error {
	if len(cmdArgs) == 0 {
		return errors.New("no command specified")
	}

	fmt.Printf("Starting monitor with command: %v\n", strings.Join(cmdArgs, " "))

	lastRestartTimestampMetric, totalRestartsMetric := c.startHealthcheckEndpoint()

	//MaxRestarts required for being able to unit test
	for i := 0; i <= *c.Opts.MaxRestarts || *c.Opts.MaxRestarts == -1; i++ {
		select {
		case <-ctx.Done():
			fmt.Println("Stopping command runner.")
			return nil
		default:
			fmt.Printf("Running command: %s\n", cmdArgs)
			startTime := c.Time.Now()
			returnCode, err := csio.ExecuteCommand(ctx, cmdArgs)
			if err != nil {
				return fmt.Errorf("error executing command %s: %w", cmdArgs, err)
			}
			duration := time.Since(startTime)
			strReturnCode := strconv.Itoa(returnCode)

			lastRestartTimestampMetric.WithLabelValues(strReturnCode).Set(float64(c.Time.Now().Unix()))
			totalRestartsMetric.WithLabelValues(strReturnCode).Inc()

			fmt.Printf("Command finished with return code: %d (took %v)\n", returnCode, duration)

			// Delay in case of fast non-zero exit
			if returnCode != 0 && duration < 1*time.Second {
				fmt.Printf("Command exited with non-zero code (%d) in less than 1 second (%v). Waiting 5 seconds before next restart...\n", returnCode, duration)
				c.Time.Sleep(5 * time.Second)
			}
			fmt.Println("cs monitor: restarting.")
		}
	}
	fmt.Println("Maximum number of restarts reached, exiting.")
	return nil
}

func (c *MonitorCmd) startHealthcheckEndpoint() (*prometheus.GaugeVec, *prometheus.CounterVec) {
	lastRestartTimestamp := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cs_monitor_last_restart_timestamp_seconds",
			Help: "Timestamp of the last command execution completion.",
		},
		[]string{"return_code"},
	)
	totalRestarts := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cs_monitor_total_restarts_total",
			Help: "Total number of command executions completed.",
		},
		[]string{"return_code"},
	)

	customRegistry := prometheus.NewRegistry()
	customRegistry.MustRegister(lastRestartTimestamp)
	customRegistry.MustRegister(totalRestarts)

	// Initialize metric with start time
	lastRestartTimestamp.WithLabelValues("").Set(float64(time.Now().Unix()))

	http.Handle("/metrics", promhttp.HandlerFor(customRegistry, promhttp.HandlerOpts{}))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/metrics", http.StatusMovedPermanently)
	})

	go func() {
		fmt.Printf("Prometheus metrics will be available on %s/metrics\n", *c.Opts.ListenAddress)
		err := http.ListenAndServe(*c.Opts.ListenAddress, nil)
		if err != nil {
			panic(fmt.Errorf("error starting metrics server on %s: %v", *c.Opts.ListenAddress, err))
		}
	}()
	return lastRestartTimestamp, totalRestarts
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
				* cs_monitor_last_restart_timestamp_seconds - Timestamp of the last command execution completion
				* cs_monitor_total_restarts_total - Total number of command executions completed`),
			Example: csio.FormatExampleCommands("monitor", []csio.Example{
				{Cmd: "-- npm start", Desc: "monitor application that ist started by npm"},
				{Cmd: "--address 8080 -- ./my-app -p 3000 ", Desc: "monitor application from local binary on port 3000, expose metrics on port 8080"},
			}),
		},
		Time: &api.RealTime{},
		Opts: MonitorOpts{GlobalOptions: opts},
		Http: &RealHttpServer{},
	}
	monitor.Opts.ListenAddress = monitor.cmd.Flags().String("address", ":3000", "Custom listen address for the metrics endpoint")
	monitor.Opts.MaxRestarts = monitor.cmd.Flags().Int("max-restarts", -1, "Maximum number of restarts before exiting")
	rootCmd.AddCommand(monitor.cmd)
	monitor.cmd.RunE = monitor.RunE
}
