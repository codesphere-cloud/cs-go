// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"time"

	openapi "github.com/codesphere-cloud/cs-go/api/openapi_client"
)

type OpenAPITeam = openapi.TeamsGetTeam200Response
type OpenAPIListTeam = openapi.TeamsListTeams200ResponseInner
type OpenAPIOrgTeam = openapi.OrganizationsListOrgTeams200ResponseInner
type OpenAPITeamMember = openapi.TeamsListMembers200ResponseInner

type DataCenter = openapi.MetadataGetDatacenters200ResponseInner
type Team struct {
	Id                  int     `json:"id"`
	DefaultDataCenterId int     `json:"defaultDataCenterId"`
	Name                string  `json:"name"`
	Description         *string `json:"description,omitempty"`
	AvatarId            *string `json:"avatarId,omitempty"`
	AvatarUrl           *string `json:"avatarUrl,omitempty"`
	IsFirst             *bool   `json:"isFirst,omitempty"`
	OrganizationId      *string `json:"organizationId,omitempty"`
	Role                *int    `json:"role,omitempty"`
}
type Domain = openapi.DomainsGetDomain200Response
type DomainVerificationStatus = openapi.DomainsGetDomain200ResponseDomainVerificationStatus
type UpdateDomainArgs = openapi.DomainsUpdateDomainRequest
type PathToWorkspaces = map[string][]*Workspace
type Organization = openapi.ClustersListAllOrganizations200ResponseInner
type Workspace = openapi.WorkspacesGetWorkspace200Response
type Baseimage = openapi.MetadataGetWorkspaceBaseImages200ResponseInner
type WorkspaceStatus = openapi.WorkspacesGetWorkspaceStatus200Response
type CreateWorkspaceArgs = openapi.WorkspacesCreateWorkspaceRequest
type WorkspacePlan = openapi.MetadataGetWorkspacePlans200ResponseInner

type PipelineStatus = openapi.WorkspacesPipelineStatus200ResponseInner

type TeamMember struct {
	UserId    int        `json:"userId"`
	TeamId    int        `json:"teamId"`
	Role      int        `json:"role"`
	Pending   bool       `json:"pending"`
	CreatedAt time.Time  `json:"createdAt"`
	Name      *string    `json:"name,omitempty"`
	Email     *string    `json:"email,omitempty"`
	AvatarUrl *string    `json:"avatarUrl,omitempty"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

func ConvertToTeam(t *OpenAPITeam) *Team {
	return &Team{
		Id:                  t.Id,
		DefaultDataCenterId: t.DefaultDataCenterId,
		Name:                t.Name,
		Description:         nullableStringToPtr(t.Description),
		AvatarId:            nullableStringToPtr(t.AvatarId),
		AvatarUrl:           nullableStringToPtr(t.AvatarUrl),
		IsFirst:             t.IsFirst,
		OrganizationId:      t.OrganizationId,
		Role:                nil, // GetTeam API does not return role
	}
}

func ConvertOrgTeamToTeam(t OpenAPIOrgTeam, orgId string) *Team {
	return &Team{
		Id:                  t.Id,
		DefaultDataCenterId: t.DefaultDataCenterId,
		Name:                t.Name,
		Description:         t.Description,
		AvatarId:            t.AvatarId,
		AvatarUrl:           t.AvatarUrl,
		IsFirst:             t.IsFirst,
		OrganizationId:      &orgId,
		Role:                nil, // Org teams API does not return role
	}
}

func ConvertFromListTeams(t OpenAPIListTeam) *Team {
	role := t.Role
	return &Team{
		Id:                  t.Id,
		DefaultDataCenterId: t.DefaultDataCenterId,
		Name:                t.Name,
		Description:         nullableStringToPtr(t.Description),
		AvatarId:            nullableStringToPtr(t.AvatarId),
		AvatarUrl:           nullableStringToPtr(t.AvatarUrl),
		IsFirst:             t.IsFirst,
		OrganizationId:      t.OrganizationId,
		Role:                &role,
	}
}

func nullableStringToPtr(ns openapi.NullableString) *string {
	if ns.IsSet() && ns.Get() != nil {
		v := *ns.Get()
		return &v
	}
	return nil
}

func ConvertToTeamMember(t OpenAPITeamMember) TeamMember {
	return TeamMember{
		UserId:    t.UserId,
		TeamId:    t.TeamId,
		Role:      t.Role,
		Pending:   t.Pending,
		CreatedAt: t.CreatedAt,
		Name:      nullableStringToPtr(t.Name),
		Email:     nullableStringToPtr(t.Email),
		AvatarUrl: nullableStringToPtr(t.AvatarUrl),
		DeletedAt: t.DeletedAt,
	}
}

type Time interface {
	Sleep(time.Duration)
	Now() time.Time
}

type RealTime struct{}

func (r *RealTime) Now() time.Time {
	return time.Now()
}

func (r *RealTime) Sleep(t time.Duration) {
	time.Sleep(t)
}
