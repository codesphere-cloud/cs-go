package api_test

import (
	"testing"

	"context"
	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/api/openapi_client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListWorkspaces(t *testing.T) {
	wsApiMock := openapi_client.NewMockWorkspacesAPI(t)
	apis := openapi_client.APIClient{
		WorkspacesAPI: wsApiMock,
	}
	client := api.NewClientWithCustomApi(context.TODO(), api.Configuration{}, &apis)

	workspaces := []api.Workspace{
		{Id: 0, Name: "fakeForTeam0"},
		{Id: 1, Name: "fakeForTeam1"},
	}
	teamId := 42

	wsApiMock.EXPECT().WorkspacesListWorkspaces(mock.Anything, float32(teamId)).Return(openapi_client.ApiWorkspacesListWorkspacesRequest{})
	wsApiMock.EXPECT().WorkspacesListWorkspacesExecute(mock.Anything).Return(workspaces, nil, nil)
	workspaces, err := client.ListWorkspaces(teamId)
	assert.Nil(t, err, "should be nil")
	assert.Equal(t, workspaces, workspaces)
}
