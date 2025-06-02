// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Workspace", func() {
	var (
		mockEnv    *cmd.MockEnv
		mockClient *cmd.MockClient
	)

	BeforeEach(func() {
		mockEnv = cmd.NewMockEnv(GinkgoT())
		mockClient = cmd.NewMockClient(GinkgoT())
	})

	It("lists workspaces of single team", func() {
		l := newListWorkspacesCmdWithTeam(mockEnv, 0)
		mockClient.EXPECT().ListWorkspaces(0).Return([]api.Workspace{}, nil)

		w, err := l.ListWorkspaces(mockClient)
		Expect(w).To(Equal([]api.Workspace{}))
		Expect(err).NotTo(HaveOccurred())
	})

	It("lists workspaces of all teams when no team ID is set", func() {
		mockEnv := cmd.NewMockEnv(GinkgoT())
		l := newListWorkspacesCmd(mockEnv)
		mockEnv.EXPECT().GetTeamId().Return(-1, nil)
		mockClient.EXPECT().ListTeams().Return([]api.Team{{Id: 0}, {Id: 1}}, nil)

		expectedWorkspaces := []api.Workspace{
			{Id: 0, Name: "fakeForTeam0"},
			{Id: 1, Name: "fakeForTeam1"},
		}
		mockClient.EXPECT().ListWorkspaces(0).Return([]api.Workspace{expectedWorkspaces[0]}, nil)
		mockClient.EXPECT().ListWorkspaces(1).Return([]api.Workspace{expectedWorkspaces[1]}, nil)

		w, err := l.ListWorkspaces(mockClient)
		Expect(w).To(Equal(expectedWorkspaces))
		Expect(err).NotTo(HaveOccurred())
	})
})

func newListWorkspacesCmdWithTeam(mockEnv *cmd.MockEnv, teamId int) cmd.ListWorkspacesCmd {
	return cmd.ListWorkspacesCmd{
		Opts: cmd.GlobalOptions{
			TeamId: &teamId,
			Env:    mockEnv,
		},
	}
}

func newListWorkspacesCmd(mockEnv *cmd.MockEnv) cmd.ListWorkspacesCmd {
	return cmd.ListWorkspacesCmd{
		Opts: cmd.GlobalOptions{
			Env: mockEnv,
		},
	}
}
