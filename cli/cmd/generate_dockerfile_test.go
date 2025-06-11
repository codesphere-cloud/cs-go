package cmd_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"

	"github.com/codesphere-cloud/cs-go/cli/cmd"
)

var _ = Describe("GenerateDockerfile", func() {
	var (
		mockEnv    *cmd.MockEnv
		mockClient *cmd.MockClient
		c          *cmd.GenerateDockerfileCmd
		wsId       int
	)

	JustBeforeEach(func() {
		mockClient = cmd.NewMockClient(GinkgoT())
		mockEnv = cmd.NewMockEnv(GinkgoT())
		c = &cmd.GenerateDockerfileCmd{
			Opts: cmd.GenerateDockerfileOpts{
				GlobalOptions: cmd.GlobalOptions{
					Env:         mockEnv,
					WorkspaceId: &wsId,
				},
			},
		}
		fmt.Printf("Using mock client: %v, command: %v\n", mockClient, c)
	})

	// TODO write tests for GenerateDockerfileCmd
})
