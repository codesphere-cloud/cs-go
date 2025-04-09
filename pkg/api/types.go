package api

import (
	openapi "github.com/codesphere-cloud/cs-go/pkg/api/openapi_client"
)

type DataCenter = openapi.MetadataGetDatacenters200ResponseInner
type Team = openapi.TeamsListTeams200ResponseInner
type Workspace = openapi.WorkspacesGetWorkspace200Response
type WorkspacePlan = openapi.MetadataGetWorkspacePlans200ResponseInner
