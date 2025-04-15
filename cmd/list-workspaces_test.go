// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"testing"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cmd"
	"github.com/stretchr/testify/assert"
)

func TestListWorkspaces(t *testing.T) {
	l := newListWorkspacesCmdWithTeam(0)
	client := NewMockClient(t)
	client.EXPECT().ListWorkspaces(0).Return([]api.Workspace{}, nil)

	w, err := l.ListWorkspaces(client)
	assert.Equal(t, w, []api.Workspace{}, "should return empty list of workspaces")
	assert.Nil(t, err, "should be nil")
}

func TestListWorkspacesAllTeams(t *testing.T) {
	l := newListWorkspacesCmd()
	client := NewMockClient(t)
	client.EXPECT().ListTeams().Return([]api.Team{{Id: 0}, {Id: 1}}, nil)

	expectedWorkspaces := []api.Workspace{
		{Id: 0, Name: "fakeForTeam0"},
		{Id: 1, Name: "fakeForTeam1"},
	}
	client.EXPECT().ListWorkspaces(0).Return([]api.Workspace{expectedWorkspaces[0]}, nil)
	client.EXPECT().ListWorkspaces(1).Return([]api.Workspace{expectedWorkspaces[1]}, nil)

	w, err := l.ListWorkspaces(client)
	assert.Equal(t, w, expectedWorkspaces, "should return both workspaces")
	assert.Nil(t, err, "should be nil")
}

func newListWorkspacesCmdWithTeam(teamId int) cmd.ListWorkspacesCmd {
	return cmd.ListWorkspacesCmd{
		Opts: cmd.ListWorkspacesOptions{
			TeamId: &teamId,
		},
	}
}

func newListWorkspacesCmd() cmd.ListWorkspacesCmd {
	return cmd.ListWorkspacesCmd{}
}
