// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package list_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/spf13/cobra"

	"github.com/codesphere-cloud/cs-go/cli/cmd"
	listcmd "github.com/codesphere-cloud/cs-go/cli/cmd/list"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func writeSSELogs(w http.ResponseWriter, entries []listcmd.LogEntry) {
	payload, _ := json.Marshal(entries)
	fmt.Fprintf(w, "event: message\ndata: %s\n\n", payload)
}

var _ = Describe("ListLandscapeLogs", func() {
	var (
		mockEnv    *cmd.MockEnv
		globalOpts *cmd.GlobalOptions
		listOpts   *listcmd.ListOptions
		parentCmd  *cobra.Command
		server     *httptest.Server
		mux        *http.ServeMux
		wsId       int

		originalToken string
		originalApi   string
	)

	BeforeEach(func() {
		mockEnv = cmd.NewMockEnv(GinkgoT())
		wsId = 42

		mux = http.NewServeMux()
		server = httptest.NewServer(mux)

		// cs.GetPipelineStatus/cs.SetAuthoriziationHeader read CS_API/CS_TOKEN
		// directly from the OS environment, independent of RootOptions.
		originalApi = os.Getenv("CS_API")
		originalToken = os.Getenv("CS_TOKEN")
		_ = os.Setenv("CS_API", server.URL)
		_ = os.Setenv("CS_TOKEN", "test-token")

		globalOpts = &cmd.GlobalOptions{
			Env:         mockEnv,
			WorkspaceId: wsId,
			ApiUrl:      server.URL,
		}
		listOpts = &listcmd.ListOptions{RootOptions: globalOpts}

		parentCmd = &cobra.Command{Use: "list"}
		listcmd.AddListLandscapeLogsCmd(parentCmd, listOpts)
	})

	AfterEach(func() {
		server.Close()
		_ = os.Setenv("CS_API", originalApi)
		_ = os.Setenv("CS_TOKEN", originalToken)
		mockEnv.AssertExpectations(GinkgoT())
	})

	// captureLogOutput temporarily redirects the stdlib "log" package output
	// (used by the command to print log lines) so it can be asserted on.
	captureLogOutput := func(run func()) string {
		r, w, _ := os.Pipe()
		oldOutput := log.Writer()
		log.SetOutput(w)

		run()

		log.SetOutput(oldOutput)
		_ = w.Close()
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		return buf.String()
	}

	Context("RunE execution flow", func() {
		It("fails when workspace ID is unavailable", func() {
			globalOpts.WorkspaceId = -1
			mockEnv.EXPECT().GetWorkspaceId().Return(-1, errors.New("CS_WORKSPACE_ID env var required, but not set")).Once()

			parentCmd.SetArgs([]string{"landscape-logs"})
			err := parentCmd.Execute()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get workspace ID"))
		})

		It("retrieves logs scoped to a server", func() {
			mux.HandleFunc(fmt.Sprintf("GET /workspaces/%d/logs/run/0/server/app", wsId), func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Header.Get("Authorization")).To(Equal("Bearer test-token"))
				writeSSELogs(w, []listcmd.LogEntry{{Timestamp: "t1", Kind: "stdout", Data: "server log line"}})
			})

			parentCmd.SetArgs([]string{"landscape-logs", "-s", "app"})
			output := captureLogOutput(func() {
				err := parentCmd.Execute()
				Expect(err).NotTo(HaveOccurred())
			})
			Expect(output).To(ContainSubstring("server log line"))
		})

		It("retrieves logs scoped to a replica", func() {
			mux.HandleFunc(fmt.Sprintf("GET /workspaces/%d/logs/run/0/replica/replica-1", wsId), func(w http.ResponseWriter, r *http.Request) {
				writeSSELogs(w, []listcmd.LogEntry{{Timestamp: "t2", Kind: "stdout", Data: "replica log line"}})
			})

			parentCmd.SetArgs([]string{"landscape-logs", "-r", "replica-1"})
			output := captureLogOutput(func() {
				err := parentCmd.Execute()
				Expect(err).NotTo(HaveOccurred())
			})
			Expect(output).To(ContainSubstring("replica log line"))
		})

		It("retrieves logs scoped to a stage", func() {
			mux.HandleFunc(fmt.Sprintf("GET /workspaces/%d/logs/build/2", wsId), func(w http.ResponseWriter, r *http.Request) {
				writeSSELogs(w, []listcmd.LogEntry{{Timestamp: "t3", Kind: "stdout", Data: "stage log line"}})
			})

			parentCmd.SetArgs([]string{"landscape-logs", "--stage", "build", "-n", "2"})
			output := captureLogOutput(func() {
				err := parentCmd.Execute()
				Expect(err).NotTo(HaveOccurred())
			})
			Expect(output).To(ContainSubstring("stage log line"))
		})

		It("retrieves logs of all replicas when no scope is given", func() {
			mux.HandleFunc(fmt.Sprintf("GET /workspaces/%d/pipeline/run", wsId), func(w http.ResponseWriter, r *http.Request) {
				status := []cs.ReplicaStatus{
					{State: "running", Steps: []cs.Step{{State: "done"}}, Replica: "replica-1", Server: "app"},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(status)
			})
			mux.HandleFunc(fmt.Sprintf("GET /workspaces/%d/logs/run/0/replica/replica-1", wsId), func(w http.ResponseWriter, r *http.Request) {
				writeSSELogs(w, []listcmd.LogEntry{{Timestamp: "t4", Kind: "stdout", Data: "all logs line"}})
			})

			parentCmd.SetArgs([]string{"landscape-logs"})
			output := captureLogOutput(func() {
				err := parentCmd.Execute()
				Expect(err).NotTo(HaveOccurred())
			})
			Expect(output).To(ContainSubstring("all logs line"))
		})
	})
})

var _ = Describe("AddListLandscapeLogsCmd", func() {
	var (
		parentCmd *cobra.Command
		listOpts  *listcmd.ListOptions
	)

	BeforeEach(func() {
		parentCmd = &cobra.Command{Use: "list"}
		listOpts = &listcmd.ListOptions{RootOptions: &cmd.GlobalOptions{}}
	})

	It("adds the landscape-logs command with correct properties", func() {
		listcmd.AddListLandscapeLogsCmd(parentCmd, listOpts)

		var logsCmd *cobra.Command
		for _, c := range parentCmd.Commands() {
			if c.Use == "landscape-logs" {
				logsCmd = c
				break
			}
		}

		Expect(logsCmd).NotTo(BeNil())
		Expect(logsCmd.Short).To(Equal("Retrieve run logs from services"))
		Expect(logsCmd.RunE).NotTo(BeNil())
		Expect(logsCmd.Flags().Lookup("server")).NotTo(BeNil())
		Expect(logsCmd.Flags().Lookup("stage")).NotTo(BeNil())
		Expect(logsCmd.Flags().Lookup("step")).NotTo(BeNil())
		Expect(logsCmd.Flags().Lookup("replica")).NotTo(BeNil())

		stage, err := logsCmd.Flags().GetString("stage")
		Expect(err).NotTo(HaveOccurred())
		Expect(stage).To(Equal("run"))
	})
})
