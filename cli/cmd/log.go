// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	stdlog "log"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/codesphere-cloud/cs-go/api/errors"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	csio "github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type LogCmd struct {
	cmd   *cobra.Command
	scope LogCmdScope
	opts  GlobalOptions
}

type LogCmdScope struct {
	workspaceId int
	server      *string
	stage       *string
	step        *int
	replica     *string
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Kind      string `json:"kind"`
	Data      string `json:"data"`
}

type SSE struct {
	event string
	data  string
}

func AddLogCmd(rootCmd *cobra.Command, opts GlobalOptions) {
	logCmd := LogCmd{
		cmd: &cobra.Command{
			Use:   "log",
			Short: "Retrieve run logs from services",
			Long: csio.Long(`You can retrieve logs based on the given scope.

				If you provide the step number and server, it returns all logs from
				all replicas of that server.

				If you provide a specific replica id, it will return the logs of
				that replica only.`),
			Example: csio.FormatExampleCommands("log", []csio.Example{
				{Cmd: "-w 637128 -s app", Desc: "Get logs from a server"},
				{Cmd: "-w 637128", Desc: "Get all logs of all servers"},
				{Cmd: "-w 637128 -r workspace-213d7a8c-48b4-42e2-8f70-c905ab04abb5-58d657cdc5-m8rrp", Desc: "Get logs from a replica"},
			}),
		},
		opts: opts,
	}
	logCmd.cmd.RunE = logCmd.RunE
	logCmd.parseLogCmdFlags()
	rootCmd.AddCommand(logCmd.cmd)
}

func (logCmd *LogCmd) parseLogCmdFlags() {
	logCmd.scope = LogCmdScope{
		server:  logCmd.cmd.Flags().StringP("server", "s", "", "Name of the landscape server"),
		stage:   logCmd.cmd.Flags().String("stage", "run", "Stage to stream logs from"),
		step:    logCmd.cmd.Flags().IntP("step", "n", 0, "Index of execution step (default 0)"),
		replica: logCmd.cmd.Flags().StringP("replica", "r", "", "ID of server replica"),
	}
}

func (l *LogCmd) RunE(_ *cobra.Command, args []string) (err error) {
	l.scope.workspaceId, err = l.opts.GetWorkspaceId()
	if err != nil {
		return fmt.Errorf("failed to get workspace ID: %w", err)
	}

	if *l.scope.replica != "" {
		if *l.scope.server != "codesphere-ide" {
			slog.Warn(
				"Ignoring server flag (providing replica ID is sufficient).",
				"replica", *l.scope.replica,
				"server", *l.scope.server,
			)
		}
		return l.printLogsOfReplica("")
	}
	if *l.scope.server != "" {
		return l.printLogsOfServer()
	}
	if *l.scope.stage != "run" {
		return l.printLogsOfStage()
	}
	return l.printAllLogs()
}

func (l *LogCmd) printAllLogs() error {
	stdlog.Println("Printing logs of all replicas")

	replicas, err := cs.GetPipelineStatus(l.scope.workspaceId, *l.scope.stage)
	if err != nil {
		return fmt.Errorf("failed to get pipeline status: %w", err)
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
				err = l.printLogsOfReplica(prefix)
				if err != nil {
					stdlog.Printf("Error printling logs: %s\n", err.Error())
				}
			}()
		}
	}
	wg.Wait()

	return nil
}

func (l *LogCmd) printLogsOfStage() error {
	endpoint := fmt.Sprintf(
		"%s/workspaces/%d/logs/%s/%d",
		l.opts.GetApiUrl(),
		l.scope.workspaceId,
		*l.scope.stage,
		*l.scope.step,
	)
	return printLogsOfEndpoint("", endpoint)
}

func (l *LogCmd) printLogsOfReplica(prefix string) error {
	endpoint := fmt.Sprintf(
		"%s/workspaces/%d/logs/run/%d/replica/%s",
		l.opts.GetApiUrl(),
		l.scope.workspaceId,
		*l.scope.step,
		*l.scope.replica,
	)
	return printLogsOfEndpoint(prefix, endpoint)
}

func (l *LogCmd) printLogsOfServer() error {
	endpoint := fmt.Sprintf(
		"%s/workspaces/%d/logs/run/%d/server/%s",
		l.opts.GetApiUrl(),
		l.scope.workspaceId,
		*l.scope.step,
		*l.scope.server,
	)
	return printLogsOfEndpoint("", endpoint)
}

func printLogsOfEndpoint(prefix string, endpoint string) error {
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
		return fmt.Errorf("failed to set header: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request logs: %w", err)
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
			var errRes errors.APIErrorResponse
			err = json.Unmarshal([]byte(sse.data), &errRes)
			if err != nil {
				return fmt.Errorf("error reading error json: %w", err)
			}
			return fmt.Errorf(
				"API error %d %s: %s",
				errRes.Status, errRes.Title, errRes.Detail,
			)
		}

		for i := 0; i < len(log); i++ {
			stdlog.Printf("%s%s| %s", log[i].Timestamp, prefix, log[i].Data)
		}
	}
}
