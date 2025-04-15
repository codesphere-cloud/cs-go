package api

import (
	openapi "github.com/codesphere-cloud/cs-go/api/openapi_client"
)

type DataCenter = openapi.MetadataGetDatacenters200ResponseInner
type Team = openapi.TeamsGetTeam200Response
type Domain = openapi.DomainsGetDomain200Response
type DomainVerificationStatus = openapi.DomainsGetDomain200ResponseDomainVerificationStatus
type PathToWorkspaces = map[string][]*Workspace
type Workspace = openapi.WorkspacesGetWorkspace200Response
type WorkspacePlan = openapi.MetadataGetWorkspacePlans200ResponseInner
