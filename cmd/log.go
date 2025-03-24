/*
Copyright © 2025 Alex Klein <alex@codesphere.com>
*/
package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type LogCmd struct {
	cmd   *cobra.Command
	scope LogCmdScope
}

type LogCmdScope struct {
	workspaceId *int32
	server      *string
	step        *int32
	replica     *string
	host        *string
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Kind      string `json:"kind"`
	Data      string `json:"data"`
}

type ErrResponse struct {
	Status  int    `json:"status"`
	Title   string `json:"title"`
	Detail  string `json:"detail"`
	TraceId string `json:"traceId"`
}

type SSE struct {
	event string
	data  string
}

func addLogCmd(rootCmd *cobra.Command) {
	logCmd := LogCmd{
		cmd: &cobra.Command{
			Use:   "log",
			Short: "Retrieve Run logs from services",
			Long: `You can retrieve logs based on the given scope.

	If you provide the step number and server, it returns all logs from
	all replicas of that server.

	If you provide a specific replica id, it will return the logs of
	that replica only.`,
			Example: `
	Get logs from a server
		log -w 637128 -s app
	Get logs from a replica
		log -w 637128 -r workspace-213d7a8c-48b4-42e2-8f70-c905ab04abb5-58d657cdc5-m8rrp
	Get logs from a self-hosted Codesphere installation:
		log --host codesphere.acme.com -w 637128 -s app`,
		},
	}
	logCmd.cmd.RunE = logCmd.RunE
	logCmd.parseLogCmdFlags()
	rootCmd.AddCommand(logCmd.cmd)
}

func (logCmd *LogCmd) parseLogCmdFlags() {
	logCmd.scope = LogCmdScope{
		host:        logCmd.cmd.Flags().String("host", "codesphere.com", "Hostname of Codesphere installation"),
		workspaceId: logCmd.cmd.Flags().Int32P("workspace-id", "w", 0, "ID of Codesphere workspace"),
		server:      logCmd.cmd.Flags().StringP("server", "s", "codesphere-ide", "Name of the landscape server"),
		step:        logCmd.cmd.Flags().Int32P("step", "n", 0, "Index of execution step (default 0)"),
		replica:     logCmd.cmd.Flags().StringP("replica", "r", "", "ID of server replica"),
	}
}

func (logCmd *LogCmd) RunE(_ *cobra.Command, args []string) error {
	if *logCmd.scope.workspaceId == 0 {
		return errors.New("Workspace ID required, but not provided.")
	}

	apiToken := os.Getenv("CS_TOKEN")
	if apiToken == "" {
		return errors.New("CS_TOKEN env var required, but not set.")
	}

	if *logCmd.scope.replica != "" {
		if *logCmd.scope.server != "codesphere-ide" {
			slog.Warn(
				"Ignoring server flag (providing replica ID is sufficient).",
				"replica", *logCmd.scope.replica,
				"server", *logCmd.scope.server,
			)
		}
		return printLogsOfReplica(apiToken, &logCmd.scope)
	}
	if *logCmd.scope.server != "" {
		return printLogsOfServer(apiToken, &logCmd.scope)
	}
	return errors.New("Server name must not be empty")
}

func printLogsOfReplica(apiToken string, scope *LogCmdScope) error {
	endpoint := fmt.Sprintf(
		"https://%s/api/workspaces/%d/logs/run/%d/replica/%s",
		*scope.host,
		*scope.workspaceId,
		*scope.step,
		*scope.replica,
	)
	return printLogsOfEndpoint(apiToken, endpoint)
}

func printLogsOfServer(apiToken string, scope *LogCmdScope) error {
	endpoint := fmt.Sprintf(
		"https://%s/api/workspaces/%d/logs/run/%d/server/%s",
		*scope.host,
		*scope.workspaceId,
		*scope.step,
		*scope.server,
	)
	return printLogsOfEndpoint(apiToken, endpoint)
}

func printLogsOfEndpoint(apiToken string, endpoint string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("Failed to construct request: %s", err)
	}

	// Set the Accept header to indicate SSE
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Authorization", "Bearer "+apiToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to request logs: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Log server responded with non-ok code: %d", resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)

	for {
		sse := SSE{event: "", data: ""}
		for { // SSE parsing
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return fmt.Errorf("Failed to parse log: %s", err)
			}

			line = strings.TrimSpace(line)

			if strings.HasPrefix(line, "data:") {
				data := strings.TrimPrefix(line, "data:")
				data = strings.TrimSpace(data)
				if sse.data != "" {
					sse.data += "\n" + data
				} else {
					sse.data = data
				}
			} else if strings.HasPrefix(line, "event:") {
				event := strings.TrimPrefix(line, "event:")
				event = strings.TrimSpace(event)
				if sse.event != "" {
					slog.Warn(
						"Received multiple event types in same SSE.",
						"old", sse.event,
						"new", event,
					)
				}
				sse.event = event
			} else if strings.HasPrefix(line, "id:") {
				// not implemented/used
			} else if strings.HasPrefix(line, "retry:") {
				slog.Warn("Received retry event, but not supported.")
			} else if line == "" {
				// empty line marks end of SSE
				break
			}
		} // end SSE parsing
		var log []LogEntry
		err := json.Unmarshal([]byte(sse.data), &log)
		if err != nil {
			var errRes ErrResponse
			json.Unmarshal([]byte(sse.data), &errRes)
			return fmt.Errorf(
				"Server responded with error: %d %s: %s",
				errRes.Status, errRes.Title, errRes.Detail,
			)
		}

		for i := 0; i < len(log); i++ {
			fmt.Print(log[i].Data)
		}
	}
}
