// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
)

var _ = Describe("StartPipeline", func() {
	var (
		mockClient *cmd.MockClient
		mockTime   *api.MockTime
		c          *cmd.StartPipelineCmd
		wsId       int
		timeout    time.Duration
		stages     []string
		profile    string
		verbose    bool
	)

	BeforeEach(func() {
		mockClient = cmd.NewMockClient(GinkgoT())
		mockTime = api.NewMockTime(GinkgoT())
		wsId = 21
		profile = ""
		timeout = 30 * time.Second
		verbose = false
	})

	JustBeforeEach(func() {
		c = &cmd.StartPipelineCmd{
			Opts: cmd.StartPipelineOpts{
				GlobalOptions: cmd.GlobalOptions{
					WorkspaceId: &wsId,
					Verbose:     &verbose,
				},
				Profile: &profile,
				Timeout: &timeout,
			},
			Time: mockTime,
		}
	})

	Context("invalid pipeline stage specified", func() {
		BeforeEach(func() {
			stages = []string{"warmup", "run", "stretch"}
		})

		It("fails before executing any stage", func() {
			err := c.StartPipelineStages(mockClient, wsId, stages)
			Expect(err).To(MatchError("invalid pipeline stage: " + stages[0]))
		})
	})

	Context("valid pipeline stages specified", func() {
		var (
			reportedStatusSuccess []api.PipelineStatus
			reportedStatusRunning []api.PipelineStatus
			reportedStatusWaiting []api.PipelineStatus
			reportedStatusFailure []api.PipelineStatus
		)

		BeforeEach(func() {
			stages = []string{"prepare", "test", "run"}
			reportedStatusSuccess = PreparePipelineStatus("success")
			reportedStatusRunning = RunPipelineStatus("running")
			reportedStatusWaiting = PreparePipelineStatus("waiting")
			reportedStatusFailure = RunPipelineStatus("failure")
		})

		Context("stages start sequentially", func() {
			BeforeEach(func() {
				currentTime := time.Unix(1746190963, 0)
				mockTime.EXPECT().Now().RunAndReturn(func() time.Time {
					return currentTime
				}).Maybe()
				mockTime.EXPECT().Sleep(mock.Anything).Run(func(t time.Duration) {
					currentTime = currentTime.Add(t)
				}).Maybe()
			})

			Context("immediately successful", func() {
				JustBeforeEach(func() {
					prepareStartCall := mockClient.EXPECT().StartPipelineStage(wsId, profile, stages[0]).Return(nil).Call
					prepareStatusCall := mockClient.EXPECT().GetPipelineState(wsId, stages[0]).Return(reportedStatusSuccess, nil).NotBefore(prepareStartCall)

					testStartCall := mockClient.EXPECT().StartPipelineStage(wsId, profile, stages[1]).Return(nil).NotBefore(prepareStatusCall)
					testStatusCall := mockClient.EXPECT().GetPipelineState(wsId, stages[1]).Return(reportedStatusSuccess, nil).NotBefore(testStartCall)

					runStartCall := mockClient.EXPECT().StartPipelineStage(wsId, profile, stages[2]).Return(nil).NotBefore(testStatusCall)
					mockClient.EXPECT().GetPipelineState(wsId, stages[2]).Return(reportedStatusRunning, nil).NotBefore(runStartCall)
				})

				Context("uses a custom profile", func() {
					BeforeEach(func() {
						profile = "prod"
					})
					It("starts all 3 stages sequentially", func() {
						err := c.StartPipelineStages(mockClient, wsId, stages)
						Expect(err).NotTo(HaveOccurred())
					})
				})

				It("starts all 3 stages sequentially", func() {
					err := c.StartPipelineStages(mockClient, wsId, stages)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("it takes more than 10 seconds to start each stage", func() {
				It("starts all 3 stages sequentially, waiting for each stage to finish", func() {
					prepareStartCall := mockClient.EXPECT().StartPipelineStage(wsId, profile, stages[0]).Return(nil).Call
					prepareStatusCall := mockClient.EXPECT().GetPipelineState(wsId, stages[0]).Return(reportedStatusRunning, nil).Times(2).NotBefore(prepareStartCall)
					prepareStatusCallSuccess := mockClient.EXPECT().GetPipelineState(wsId, stages[0]).Return(reportedStatusSuccess, nil).NotBefore(prepareStatusCall)

					testStartCall := mockClient.EXPECT().StartPipelineStage(wsId, profile, stages[1]).Return(nil).NotBefore(prepareStatusCallSuccess)
					testStatusCall := mockClient.EXPECT().GetPipelineState(wsId, stages[1]).Return(reportedStatusRunning, nil).Times(2).NotBefore(testStartCall)
					testStatusCallSuccess := mockClient.EXPECT().GetPipelineState(wsId, stages[1]).Return(reportedStatusSuccess, nil).NotBefore(testStatusCall)

					runStartCall := mockClient.EXPECT().StartPipelineStage(wsId, profile, stages[2]).Return(nil).NotBefore(testStatusCallSuccess)
					mockClient.EXPECT().GetPipelineState(wsId, stages[2]).Return(reportedStatusWaiting, nil).Times(2).NotBefore(runStartCall)
					mockClient.EXPECT().GetPipelineState(wsId, stages[2]).Return(reportedStatusRunning, nil).NotBefore(runStartCall)

					err := c.StartPipelineStages(mockClient, wsId, stages)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("it takes the test stage more than the timeout complete", func() {
				It("returns a timeout error", func() {
					prepareStartCall := mockClient.EXPECT().StartPipelineStage(wsId, profile, stages[0]).Return(nil).Call
					prepareStatusCall := mockClient.EXPECT().GetPipelineState(wsId, stages[0]).Return(reportedStatusSuccess, nil).NotBefore(prepareStartCall)

					testStartCall := mockClient.EXPECT().StartPipelineStage(wsId, profile, stages[1]).Return(nil).NotBefore(prepareStatusCall)

					//this should result in a timeout
					mockClient.EXPECT().GetPipelineState(wsId, stages[1]).Return(reportedStatusRunning, nil).Times(8).NotBefore(testStartCall)

					err := c.StartPipelineStages(mockClient, wsId, stages)
					Expect(err).To(MatchError("failed waiting for stage test to finish: timed out waiting for pipeline stage test to be complete"))
				})
			})

			Context("prepare stage fails", func() {
				It("propagates the failure", func() {
					prepareStartCall := mockClient.EXPECT().StartPipelineStage(wsId, profile, stages[0]).Return(nil).Call
					mockClient.EXPECT().GetPipelineState(wsId, stages[0]).Return(reportedStatusFailure, nil).NotBefore(prepareStartCall)

					err := c.StartPipelineStages(mockClient, wsId, stages)
					Expect(err).To(MatchError("failed waiting for stage prepare to finish: stage prepare failed: server A, replica 0 reached unexpected state failure"))
				})
			})
		})
	})
})

func PreparePipelineStatus(state string) []api.PipelineStatus {
	return []api.PipelineStatus{{
		State:   "waiting",
		Replica: "0",
		Server:  "A",
	}, {
		State:   state,
		Replica: "0",
		Server:  "codesphere-ide",
	}}
}
func RunPipelineStatus(state string) []api.PipelineStatus {
	return []api.PipelineStatus{{
		State:   state,
		Replica: "0",
		Server:  "A",
	}, {
		State:   "waiting",
		Replica: "0",
		Server:  "codesphere-ide",
	}}
}
