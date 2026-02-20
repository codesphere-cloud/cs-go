// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package pipeline_test

import (
	"context"
	"fmt"
	"io"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	"github.com/codesphere-cloud/cs-go/api"
	openapi "github.com/codesphere-cloud/cs-go/api/openapi_client"
	"github.com/codesphere-cloud/cs-go/pkg/pipeline"
)

func statusWithSteps(server string, state string, stepStates ...string) api.PipelineStatus {
	steps := make([]openapi.WorkspacesPipelineStatus200ResponseInnerStepsInner, len(stepStates))
	for i, s := range stepStates {
		steps[i] = openapi.WorkspacesPipelineStatus200ResponseInnerStepsInner{State: s}
	}
	return api.PipelineStatus{
		State:   state,
		Replica: "0",
		Server:  server,
		Steps:   steps,
	}
}

var _ = Describe("Runner", func() {
	var (
		mockClient *pipeline.MockClient
		mockTime   *api.MockTime
		runner     *pipeline.Runner
		wsId       int
		cfg        pipeline.Config
	)

	BeforeEach(func() {
		wsId = 42
		cfg = pipeline.Config{
			Profile: "",
			Timeout: 30 * time.Second,
			ApiUrl:  "https://codesphere.com/api",
		}
	})

	JustBeforeEach(func() {
		mockClient = pipeline.NewMockClient(GinkgoT())
		mockTime = api.NewMockTime(GinkgoT())
		runner = pipeline.NewRunner(mockClient, mockTime)

		currentTime := time.Unix(1746190963, 0)
		mockTime.EXPECT().Now().RunAndReturn(func() time.Time {
			return currentTime
		}).Maybe()
		mockTime.EXPECT().Sleep(mock.Anything).Run(func(t time.Duration) {
			currentTime = currentTime.Add(t)
		}).Maybe()
	})

	Describe("log streaming during prepare stage", func() {
		Context("with a single step", func() {
			It("calls StreamLogs with step 0", func() {
				mockClient.EXPECT().StartPipelineStage(wsId, cfg.Profile, "prepare").Return(nil)

				pollCount := 0
				mockClient.EXPECT().GetPipelineState(wsId, "prepare").RunAndReturn(
					func(_ int, _ string) ([]api.PipelineStatus, error) {
						pollCount++
						if pollCount == 1 {
							return []api.PipelineStatus{
								statusWithSteps("codesphere-ide", "running", "running"),
							}, nil
						}
						return []api.PipelineStatus{
							statusWithSteps("codesphere-ide", "success", "success"),
						}, nil
					},
				)

				mockClient.EXPECT().StreamLogs(
					mock.Anything, cfg.ApiUrl, wsId, "prepare", 0, mock.Anything,
				).Return(nil)

				err := runner.RunStages(wsId, []string{"prepare"}, cfg)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with multiple steps", func() {
			It("streams each step sequentially", func() {
				mockClient.EXPECT().StartPipelineStage(wsId, cfg.Profile, "prepare").Return(nil)

				pollCount := 0
				mockClient.EXPECT().GetPipelineState(wsId, "prepare").RunAndReturn(
					func(_ int, _ string) ([]api.PipelineStatus, error) {
						pollCount++
						switch pollCount {
						case 1:
							return []api.PipelineStatus{
								statusWithSteps("codesphere-ide", "running", "running", "waiting"),
							}, nil
						case 2:
							return []api.PipelineStatus{
								statusWithSteps("codesphere-ide", "running", "success", "running"),
							}, nil
						default:
							return []api.PipelineStatus{
								statusWithSteps("codesphere-ide", "success", "success", "success"),
							}, nil
						}
					},
				)

				// Step 0 stream
				step0Called := make(chan struct{})
				mockClient.EXPECT().StreamLogs(
					mock.Anything, cfg.ApiUrl, wsId, "prepare", 0, mock.Anything,
				).RunAndReturn(func(_ context.Context, _ string, _ int, _ string, _ int, _ io.Writer) error {
					close(step0Called)
					return nil
				})

				// Step 1 stream — only called after step 0
				mockClient.EXPECT().StreamLogs(
					mock.Anything, cfg.ApiUrl, wsId, "prepare", 1, mock.Anything,
				).RunAndReturn(func(_ context.Context, _ string, _ int, _ string, _ int, _ io.Writer) error {
					select {
					case <-step0Called:
						// good — step 0 was called first
					default:
						return fmt.Errorf("step 1 started before step 0")
					}
					return nil
				})

				err := runner.RunStages(wsId, []string{"prepare"}, cfg)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when ApiUrl is empty", func() {
			It("does not call StreamLogs", func() {
				cfg.ApiUrl = ""

				startCall := mockClient.EXPECT().StartPipelineStage(wsId, cfg.Profile, "prepare").Return(nil).Call
				mockClient.EXPECT().GetPipelineState(wsId, "prepare").Return([]api.PipelineStatus{
					statusWithSteps("codesphere-ide", "success", "success"),
				}, nil).NotBefore(startCall)
				// StreamLogs should NOT be called — mockery will fail if it is

				err := runner.RunStages(wsId, []string{"prepare"}, cfg)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("for the run stage", func() {
			It("does not stream logs", func() {
				syncCall := mockClient.EXPECT().DeployLandscape(wsId, cfg.Profile).Return(nil).Call
				startCall := mockClient.EXPECT().StartPipelineStage(wsId, cfg.Profile, "run").Return(nil).NotBefore(syncCall)
				mockClient.EXPECT().GetPipelineState(wsId, "run").Return([]api.PipelineStatus{
					{State: "running", Replica: "0", Server: "A"},
					{State: "waiting", Replica: "0", Server: "codesphere-ide"},
				}, nil).NotBefore(startCall)
				// StreamLogs should NOT be called

				err := runner.RunStages(wsId, []string{"run"}, cfg)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
