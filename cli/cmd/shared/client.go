// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package shared

import (
	"github.com/codesphere-cloud/cs-go/api"
	"time"
)

// Client is the common API contract consumed by all CLI commands.
type Client interface {
	ListTeams(orgId string) ([]api.Team, error)
	ListWorkspaces(teamId int) ([]api.Workspace, error)
	ListBaseimages() ([]api.Baseimage, error)
	ListOrganizations() ([]api.Organization, error)
	CreateOrganization(name string, adminEmail string) (*api.Organization, error)
	GetWorkspace(workspaceId int) (api.Workspace, error)
	WorkspaceStatus(workspaceId int) (*api.WorkspaceStatus, error)
	WaitForWorkspaceRunning(workspace *api.Workspace, timeout time.Duration) error
	ScaleWorkspace(wsId int, replicas int) error
	ScaleLandscapeServices(wsId int, services map[string]int) error
	SetEnvVarOnWorkspace(workspaceId int, vars map[string]string) error
	ExecCommand(workspaceId int, command string, workdir string, env map[string]string) (string, string, error)
	ListWorkspacePlans() ([]api.WorkspacePlan, error)
	DeployWorkspace(args api.DeployWorkspaceArgs) (*api.Workspace, error)
	DeleteWorkspace(wsId int) error
	StartPipelineStage(wsId int, profile string, stage string) error
	StopPipelineStage(wsId int, stage string) error
	GetPipelineState(wsId int, stage string) ([]api.PipelineStatus, error)
	GitPull(wsId int, remote string, branch string) error
	DeployLandscape(wsId int, profile string) error
	CreateTeam(orgId string, name string, dcId int) (*api.Team, error)
	DeleteTeam(teamId int) error
	AddTeamMember(teamId int, email string, role int) error
	RemoveTeamMember(teamId int, userId int) error
	ListTeamMembers(teamId int) ([]api.TeamMember, error)
}
