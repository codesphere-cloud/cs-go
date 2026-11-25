// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package shared

import (
	"context"
	"time"

	"github.com/codesphere-cloud/cs-go/api"
)

// Client is the common API contract consumed by all CLI commands.
type Client interface {
	UploadSSHPublicKey(ctx context.Context, publicKey string) error
	ListSSHPublicKeys(ctx context.Context) ([]api.SSHPublicKey, error)
	DeleteSSHPublicKey(ctx context.Context, fingerprint string) error
	ChangeOrganizationMemberRole(ctx context.Context, orgId string, userId int, role string) error
	ListOrganizationMembers(ctx context.Context, orgId string) ([]api.OrganizationMember, error)
	RemoveOrganizationMember(ctx context.Context, orgId string, userId int) error
	AddOrganizationMember(ctx context.Context, orgId, email, role string) error
	ListTeams(ctx context.Context, orgId string) ([]api.Team, error)
	ListWorkspaces(ctx context.Context, teamId int) ([]api.Workspace, error)
	ListBaseimages(ctx context.Context) ([]api.Baseimage, error)
	ListOrganizations(ctx context.Context) ([]api.Organization, error)
	CreateOrganization(ctx context.Context, name string, adminEmail string) (*api.Organization, error)
	GetWorkspace(ctx context.Context, workspaceId int) (api.Workspace, error)
	WorkspaceStatus(ctx context.Context, workspaceId int) (*api.WorkspaceStatus, error)
	WaitForWorkspaceRunning(ctx context.Context, workspace *api.Workspace, timeout time.Duration) error
	ScaleWorkspace(ctx context.Context, wsId int, replicas int) error
	ScaleLandscapeServices(ctx context.Context, wsId int, services map[string]int) error
	SetEnvVarOnWorkspace(ctx context.Context, workspaceId int, vars map[string]string) error
	ExecCommand(ctx context.Context, workspaceId int, command string, workdir string, env map[string]string) (string, string, error)
	ListWorkspacePlans(ctx context.Context) ([]api.WorkspacePlan, error)
	DeployWorkspace(ctx context.Context, args api.DeployWorkspaceArgs) (*api.Workspace, error)
	DeleteWorkspace(ctx context.Context, wsId int) error
	StartPipelineStage(ctx context.Context, wsId int, profile string, stage string) error
	StopPipelineStage(ctx context.Context, wsId int, stage string) error
	GetPipelineState(ctx context.Context, wsId int, stage string) ([]api.PipelineStatus, error)
	GitPull(ctx context.Context, wsId int, remote string, branch string) error
	DeployLandscape(ctx context.Context, wsId int, profile string) error
	CreateTeam(ctx context.Context, orgId string, name string, dcId int) (*api.Team, error)
	DeleteTeam(ctx context.Context, teamId int) error
	AddTeamMember(ctx context.Context, teamId int, email string, role int) error
	RemoveTeamMember(ctx context.Context, teamId int, userId int) error
	ListTeamMembers(ctx context.Context, teamId int) ([]api.TeamMember, error)
}
