// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package list

import "github.com/codesphere-cloud/cs-go/api"

type Client interface {
	ListTeams(orgId string) ([]api.Team, error)
	ListWorkspaces(teamId int) ([]api.Workspace, error)
	ListBaseimages() ([]api.Baseimage, error)
	ListOrganizations() ([]api.Organization, error)
	ListWorkspacePlans() ([]api.WorkspacePlan, error)
	ListTeamMembers(teamId int) ([]api.TeamMember, error)
}
