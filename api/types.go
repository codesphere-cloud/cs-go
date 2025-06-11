// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	openapi "github.com/codesphere-cloud/cs-go/api/openapi_client"
	"time"
)

type DataCenter = openapi.MetadataGetDatacenters200ResponseInner
type Team = openapi.TeamsListTeams200ResponseInner
type Domain = openapi.DomainsGetDomain200Response
type DomainVerificationStatus = openapi.DomainsGetDomain200ResponseDomainVerificationStatus
type UpdateDomainArgs = openapi.DomainsGetDomain200ResponseCustomConfig
type PathToWorkspaces = map[string][]*Workspace

type Workspace = openapi.WorkspacesGetWorkspace200Response
type WorkspaceStatus = openapi.WorkspacesGetWorkspaceStatus200Response
type CreateWorkspaceArgs = openapi.WorkspacesCreateWorkspaceRequest
type WorkspacePlan = openapi.MetadataGetWorkspacePlans200ResponseInner

type PipelineStatus = openapi.WorkspacesPipelineStatus200ResponseInner

// TODO: remove the conversion once the api is fixed
func ConvertToTeam(t *openapi.TeamsGetTeam200Response) *Team {
	return &Team{
		Id:                  t.Id,
		DefaultDataCenterId: t.DefaultDataCenterId,
		Name:                t.Name,
		Description:         t.Description,
		AvatarId:            t.AvatarId,
		AvatarUrl:           t.AvatarUrl,
		IsFirst:             t.IsFirst,

		Role: 0,
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
