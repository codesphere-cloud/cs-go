package api

import (
	openapi "github.com/codesphere-cloud/cs-go/pkg/api/openapi_client"
)

type DataCenter = openapi.MetadataGetDatacenters200ResponseInner
type Team = openapi.TeamsListTeams200ResponseInner
type Domain = openapi.DomainsGetDomain200Response
type DomainVerificationStatus = openapi.DomainsGetDomain200ResponseDomainVerificationStatus
type UpdateDomainArgs = openapi.DomainsGetDomain200ResponseCustomConfig
type PathToWorkspaces = map[string][]*Workspace
type Workspace = openapi.WorkspacesGetWorkspace200Response
type WorkspacePlan = openapi.MetadataGetWorkspacePlans200ResponseInner

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
