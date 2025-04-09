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
	"strconv"
	"strings"
	"sync"

	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/spf13/cobra"
)

type LogCmd struct {
	cmd   *cobra.Command
	scope LogCmdScope
}

type LogCmdScope struct {
	workspaceId *int
	server      *string
	step        *int
	replica     *string
	api         *string
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
	Get all logs of all servers
		log -w 637128
	Get logs from a replica
		log -w 637128 -r workspace-213d7a8c-48b4-42e2-8f70-c905ab04abb5-58d657cdc5-m8rrp
	Get logs from a self-hosted Codesphere installation:
		log --api https://codesphere.acme.com/api -w 637128 -s app`,
		},
	}
	logCmd.cmd.RunE = logCmd.RunE
	logCmd.parseLogCmdFlags()
	rootCmd.AddCommand(logCmd.cmd)
}

func (logCmd *LogCmd) parseLogCmdFlags() {
	logCmd.scope = LogCmdScope{
		api:         logCmd.cmd.Flags().String("api", "", "URL of Codesphere API (can also be CS_API)"),
		workspaceId: logCmd.cmd.Flags().IntP("workspace-id", "w", 0, "ID of Codesphere workspace (can also be CS_WORKSPACE_ID)"),
		server:      logCmd.cmd.Flags().StringP("server", "s", "", "Name of the landscape server"),
		step:        logCmd.cmd.Flags().IntP("step", "n", 0, "Index of execution step (default 0)"),
		replica:     logCmd.cmd.Flags().StringP("replica", "r", "", "ID of server replica"),
	}
}

func (logCmd *LogCmd) RunE(_ *cobra.Command, args []string) (err error) {
	if *logCmd.scope.workspaceId == 0 {
		*logCmd.scope.workspaceId, err = strconv.Atoi(os.Getenv("CS_WORKSPACE_ID"))
		if err != nil {
			return fmt.Errorf("failer to read env var: %e", err)
		}
		if *logCmd.scope.workspaceId == 0 {
			return errors.New("workspace ID required, but not provided")
		}
	}

	if *logCmd.scope.api == "" {
		*logCmd.scope.api = cs.GetApiUrl()
	}

	if *logCmd.scope.replica != "" {
		if *logCmd.scope.server != "codesphere-ide" {
			slog.Warn(
				"Ignoring server flag (providing replica ID is sufficient).",
				"replica", *logCmd.scope.replica,
				"server", *logCmd.scope.server,
			)
		}
		return printLogsOfReplica("", &logCmd.scope)
	}
	if *logCmd.scope.server != "" {
		return printLogsOfServer(&logCmd.scope)
	}

	err = logCmd.printAllLogs()
	if err != nil {
		return fmt.Errorf("failed to print logs: %e", err)
	}

	return nil
}

func (l *LogCmd) printAllLogs() error {
	fmt.Println("Printing logs of all replicas")

	replicas, err := cs.GetPipelineStatus(*l.scope.workspaceId, "run")
	if err != nil {
		return fmt.Errorf("failed to get pipeline status: %e", err)
	}

	var wg sync.WaitGroup
	for _, replica := range replicas {
		for s := range replica.Steps {
			wg.Add(1)
			go func() {
				defer wg.Done()
				scope := l.scope
				*scope.step = s
				*scope.replica = replica.Replica
				prefix := fmt.Sprintf("|%-10s|%s", replica.Server, replica.Replica[len(replica.Replica)-11:])
				err = printLogsOfReplica(prefix, &scope)
				if err != nil {
					fmt.Printf("Error printling logs: %e\n", err)
				}
			}()
		}
	}
	wg.Wait()

	return nil
}

func printLogsOfReplica(prefix string, scope *LogCmdScope) error {
	endpoint := fmt.Sprintf(
		"%s/workspaces/%d/logs/run/%d/replica/%s",
		*scope.api,
		*scope.workspaceId,
		*scope.step,
		*scope.replica,
	)
	return printLogsOfEndpoint(prefix, endpoint)
}

func printLogsOfServer(scope *LogCmdScope) error {
	endpoint := fmt.Sprintf(
		"%s/workspaces/%d/logs/run/%d/server/%s",
		*scope.api,
		*scope.workspaceId,
		*scope.step,
		*scope.server,
	)
	return printLogsOfEndpoint("", endpoint)
}

func printLogsOfEndpoint(prefix string, endpoint string) error {
	fmt.Println(endpoint)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to construct request: %s", err)
	}

	// Set the Accept header to indicate SSE
	req.Header.Set("Accept", "text/event-stream")
	err = cs.SetAuthoriziationHeader(req)
	if err != nil {
		return fmt.Errorf("failed to set header: %e", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request logs: %e", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("log server responded with non-ok code: %d", resp.StatusCode)
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
				return fmt.Errorf("failed to parse log: %s", err)
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
			err = json.Unmarshal([]byte(sse.data), &errRes)
			if err != nil {
				return fmt.Errorf("error reading error json: %e", err)
			}
			return fmt.Errorf(
				"server responded with error: %d %s: %s",
				errRes.Status, errRes.Title, errRes.Detail,
			)
		}

		for i := 0; i < len(log); i++ {
			fmt.Printf("%s%s| %s", log[i].Timestamp, prefix, log[i].Data)
		}
	}
}
