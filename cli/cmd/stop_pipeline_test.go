package cmd_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/codesphere-cloud/cs-go/cli/cmd"
)

var _ = Describe("StopPipeline", func() {
	var (
		mockClient *cmd.MockClient
		c          *cmd.StopPipelineCmd
		wsId       int
		stages     []string
	)

	BeforeEach(func() {
		mockClient = cmd.NewMockClient(GinkgoT())
		wsId = 21
	})

	JustBeforeEach(func() {
		c = &cmd.StopPipelineCmd{}
	})

	Context("invalid pipeline stage specified", func() {
		BeforeEach(func() {
			stages = []string{"warmup"}
		})

		It("fails before executing any stage", func() {
			err := c.StopPipelineStages(mockClient, wsId, stages)
			Expect(err).To(MatchError("invalid pipeline stage: " + stages[0]))
		})
	})

	Context("valid pipeline stages specified", func() {
		BeforeEach(func() {
			stages = []string{"prepare", "test", "run"}
		})

		It("stops all stages sequentially", func() {
			mockClient.EXPECT().StopPipelineStage(wsId, stages[0]).Return(nil)
			mockClient.EXPECT().StopPipelineStage(wsId, stages[1]).Return(nil)
			mockClient.EXPECT().StopPipelineStage(wsId, stages[2]).Return(nil)

			err := c.StopPipelineStages(mockClient, wsId, stages)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})