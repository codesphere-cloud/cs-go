package cmd_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/codesphere-cloud/cs-go/cli/cmd"
)

var _ = Describe("Exec", func() {
	var (
		mockEnv    *cmd.MockEnv
		mockClient *cmd.MockClient
		e          *cmd.ExecCmd
		wsId       int
		command    string
		workDir    string
		envVars    []string
	)

	BeforeEach(func() {
		envVars = []string{}
		mockClient = cmd.NewMockClient(GinkgoT())
		mockEnv = cmd.NewMockEnv(GinkgoT())
		wsId = 42
		command = "ls -al"
	})

	JustBeforeEach(func() {
		e = &cmd.ExecCmd{
			Opts: cmd.ExecOptions{
				GlobalOptions: cmd.GlobalOptions{
					Env:         mockEnv,
					WorkspaceId: &wsId,
				},
				EnvVar:  &envVars,
				WorkDir: &workDir,
			},
		}
	})

	Context("No workdir, no env vars set", func() {
		It("executes the command", func() {
			mockClient.EXPECT().ExecCommand(wsId, command, "", map[string]string{}).Return("stdout", "stderr", nil)
			err := e.ExecCommand(mockClient, command)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("env vars set", func() {
		BeforeEach(func() {
			envVars = []string{"a=b", "b=c"}
		})
		It("executes the command with env vars", func() {
			mockClient.EXPECT().ExecCommand(wsId, command, "", map[string]string{"a": "b", "b": "c"}).Return("stdout", "stderr", nil)
			err := e.ExecCommand(mockClient, command)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("work dir set", func() {
		BeforeEach(func() {
			workDir = "user"
		})
		It("executes the command with workdir", func() {
			mockClient.EXPECT().ExecCommand(wsId, command, workDir, map[string]string{}).Return("stdout", "stderr", nil)
			err := e.ExecCommand(mockClient, command)
			Expect(err).ToNot(HaveOccurred())
		})
	})

})
