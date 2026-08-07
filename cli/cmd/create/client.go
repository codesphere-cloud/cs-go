// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import "github.com/codesphere-cloud/cs-go/api"

type Client interface {
	DeployWorkspace(args api.DeployWorkspaceArgs) (*api.Workspace, error)
	ListBaseimages() ([]api.Baseimage, error)
	SetEnvVarOnWorkspace(workspaceId int, vars map[string]string) error
	CreateOrganization(name string, adminEmail string) (*api.Organization, error)
	CreateTeam(orgId string, teamName string, dcId int) (*api.Team, error)
}
