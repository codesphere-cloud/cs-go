// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package delete

import "github.com/codesphere-cloud/cs-go/api"

type Client interface {
	GetWorkspace(workspaceId int) (api.Workspace, error)
	DeleteWorkspace(wsId int) error
	DeleteTeam(teamId int) error
	RemoveTeamMember(teamId int, userId int) error
}
