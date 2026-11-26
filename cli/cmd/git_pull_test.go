// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"errors"
	"time"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("GitPull", func() {
	var (
		client *cmd.MockClient
		pull   *cmd.GitPullCmd
	)

	BeforeEach(func() {
		client = cmd.NewMockClient(GinkgoT())
		remote := "origin"
		branch := "main"
		pull = &cmd.GitPullCmd{
			Opts: cmd.GitPullOpts{
				GlobalOptions: &cmd.GlobalOptions{},
				Remote:        &remote,
				Branch:        &branch,
			},
		}
	})

	It("wakes a stopped workspace before pulling", func() {
		workspace := api.Workspace{Id: 42, Name: "test-workspace"}
		awake := false
		client.EXPECT().GetWorkspace(42).Return(workspace, nil)
		client.EXPECT().WorkspaceStatus(42).Return(&api.WorkspaceStatus{IsRunning: false}, nil)
		client.EXPECT().ScaleWorkspace(42, 1).Return(nil)
		client.EXPECT().WaitForWorkspaceRunning(mock.Anything, mock.Anything).Run(func(_ *api.Workspace, _ time.Duration) {
			awake = true
		}).Return(nil)
		client.EXPECT().GitPull(42, "origin", "main", api.GitPullOptions{}).Run(func(int, string, string, api.GitPullOptions) {
			Expect(awake).To(BeTrue())
		}).Return(nil)

		Expect(pull.PullWorkspace(client, 42)).To(Succeed())
	})

	It("pulls when the workspace is already running", func() {
		workspace := api.Workspace{Id: 42, Name: "test-workspace"}
		client.EXPECT().GetWorkspace(42).Return(workspace, nil)
		client.EXPECT().WorkspaceStatus(42).Return(&api.WorkspaceStatus{IsRunning: true}, nil)
		client.EXPECT().GitPull(42, "origin", "main", api.GitPullOptions{}).Return(nil)

		Expect(pull.PullWorkspace(client, 42)).To(Succeed())
	})

	It("passes the submodule option alone to the pull", func() {
		pull.Opts.RecurseSubmodules = true
		workspace := api.Workspace{Id: 42, Name: "test-workspace"}
		client.EXPECT().GetWorkspace(42).Return(workspace, nil)
		client.EXPECT().WorkspaceStatus(42).Return(&api.WorkspaceStatus{IsRunning: true}, nil)
		client.EXPECT().GitPull(42, "origin", "main", api.GitPullOptions{RecurseSubmodules: true}).Return(nil)

		Expect(pull.PullWorkspace(client, 42)).To(Succeed())
	})

	It("passes the Git LFS option alone to the pull", func() {
		pull.Opts.SkipLfs = true
		workspace := api.Workspace{Id: 42, Name: "test-workspace"}
		client.EXPECT().GetWorkspace(42).Return(workspace, nil)
		client.EXPECT().WorkspaceStatus(42).Return(&api.WorkspaceStatus{IsRunning: true}, nil)
		client.EXPECT().GitPull(42, "origin", "main", api.GitPullOptions{SkipLfs: true}).Return(nil)

		Expect(pull.PullWorkspace(client, 42)).To(Succeed())
	})

	It("offers the submodule and Git LFS flags, both off by default", func() {
		git := &cobra.Command{Use: "git"}
		cmd.AddGitPullCmd(git, &cmd.GlobalOptions{})
		pullCmd, _, err := git.Find([]string{"pull"})
		Expect(err).NotTo(HaveOccurred())

		recurseSubmodules, err := pullCmd.Flags().GetBool("recurse-submodules")
		Expect(err).NotTo(HaveOccurred())
		Expect(recurseSubmodules).To(BeFalse())
		skipLfs, err := pullCmd.Flags().GetBool("skip-lfs")
		Expect(err).NotTo(HaveOccurred())
		Expect(skipLfs).To(BeFalse())
	})

	It("does not pull when waking the workspace fails", func() {
		workspace := api.Workspace{Id: 42, Name: "test-workspace"}
		client.EXPECT().GetWorkspace(42).Return(workspace, nil)
		client.EXPECT().WorkspaceStatus(42).Return(&api.WorkspaceStatus{IsRunning: false}, nil)
		client.EXPECT().ScaleWorkspace(42, 1).Return(errors.New("scale failed"))

		Expect(pull.PullWorkspace(client, 42)).To(MatchError(ContainSubstring("failed to wake up workspace before git pull")))
	})
})
