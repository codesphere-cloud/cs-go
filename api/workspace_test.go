package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/api/errors"
	"github.com/codesphere-cloud/cs-go/api/openapi_client"
	tu "github.com/codesphere-cloud/cs-go/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getTestingClient(t *testing.T) (*api.Client, *openapi_client.MockWorkspacesAPI) {
	wsApiMock := openapi_client.NewMockWorkspacesAPI(t)
	apis := openapi_client.APIClient{
		WorkspacesAPI: wsApiMock,
	}
	tu.PatchTimeFuncs(t)
	return api.NewClientWithCustomApi(context.TODO(), api.Configuration{}, &apis), wsApiMock
}

func mockWorkspaceStatus(wsApiMock *openapi_client.MockWorkspacesAPI, workspaceId int, isRunning ...bool) {
	wsApiMock.EXPECT().WorkspacesGetWorkspaceStatus(mock.Anything, float32(workspaceId)).
		Return(openapi_client.ApiWorkspacesGetWorkspaceStatusRequest{ApiService: wsApiMock})
	for _, running := range isRunning {
		wsApiMock.EXPECT().WorkspacesGetWorkspaceStatusExecute(mock.Anything).Once().Return(&api.WorkspaceStatus{
			IsRunning: running,
		}, nil, nil)
	}
	mock.InOrder(wsApiMock.ExpectedCalls...)
}

func mockCreateWorkspace(wsApiMock *openapi_client.MockWorkspacesAPI, ws api.Workspace) {
	wsApiMock.EXPECT().WorkspacesCreateWorkspace(mock.Anything).
		Return(openapi_client.ApiWorkspacesCreateWorkspaceRequest{ApiService: wsApiMock})
	wsApiMock.EXPECT().WorkspacesCreateWorkspaceExecute(mock.Anything).Return(&ws, nil, nil)
}

func TestListWorkspaces(t *testing.T) {
	client, wsApiMock := getTestingClient(t)

	workspaces := []api.Workspace{
		{Id: 0, Name: "fakeForTeam0"},
		{Id: 1, Name: "fakeForTeam1"},
	}
	teamId := 42

	wsApiMock.EXPECT().WorkspacesListWorkspaces(mock.Anything, float32(teamId)).
		Return(openapi_client.ApiWorkspacesListWorkspacesRequest{ApiService: wsApiMock})
	wsApiMock.EXPECT().WorkspacesListWorkspacesExecute(mock.Anything).Return(workspaces, nil, nil)
	workspaces, err := client.ListWorkspaces(teamId)
	assert.Nil(t, err, "should be nil")
	assert.Equal(t, workspaces, workspaces)
}

func TestWaitForWorkspaceRunningSuccess(t *testing.T) {
	client, wsApiMock := getTestingClient(t)

	ws := api.Workspace{
		Id: 0, Name: "fakeWorkspace",
	}

	mockWorkspaceStatus(wsApiMock, ws.Id, true)
	err := client.WaitForWorkspaceRunning(&ws, 1*time.Millisecond)

	assert.Nil(t, err, "should be nil")
}

func TestWaitForWorkspaceRunningTimeout(t *testing.T) {
	client, wsApiMock := getTestingClient(t)

	ws := api.Workspace{
		Id: 0, Name: "fakeWorkspace",
	}

	mockWorkspaceStatus(wsApiMock, ws.Id, false, false)
	err := client.WaitForWorkspaceRunning(&ws, 1*time.Second)

	assert.IsType(t, err, &errors.TimedOutError{}, "expected timeout error")
}

func TestWaitForWorkspaceRunningOnRetry(t *testing.T) {
	client, wsApiMock := getTestingClient(t)

	ws := api.Workspace{
		Id: 42, Name: "fakeWorkspace",
	}

	mockWorkspaceStatus(wsApiMock, ws.Id, false, true)
	err := client.WaitForWorkspaceRunning(&ws, 1*time.Millisecond)

	assert.Nil(t, err, "should be nil")
}

func TestDeployWorkspace(t *testing.T) {
	client, wsApiMock := getTestingClient(t)

	ws := api.Workspace{
		Id: 42, Name: "fakeWorkspace",
	}

	wsApiMock.EXPECT().WorkspacesCreateWorkspace(mock.Anything).
		Return(openapi_client.ApiWorkspacesCreateWorkspaceRequest{ApiService: wsApiMock})
	wsApiMock.EXPECT().WorkspacesCreateWorkspaceExecute(mock.Anything).Return(&ws, nil, nil)
	mockWorkspaceStatus(wsApiMock, ws.Id, true)

	newWs, err := client.DeployWorkspace(
		api.DeployWorkspaceArgs{Timeout: 1 * time.Millisecond},
	)

	assert.Nil(t, err, "should be nil")
	assert.Equal(t, newWs.Name, ws.Name, "Should have the same name")
}

func TestDeployWorkspaceWithEnvVars(t *testing.T) {
	client, wsApiMock := getTestingClient(t)

	ws := api.Workspace{
		Id: 42, Name: "fakeWorkspace",
	}
	args := api.DeployWorkspaceArgs{
		Timeout: 1 * time.Millisecond,
		EnvVars: map[string]string{
			"foo":  "bar",
			"some": "thing",
		},
	}

	mockCreateWorkspace(wsApiMock, ws)
	mockWorkspaceStatus(wsApiMock, ws.Id, true)
	wsApiMock.EXPECT().WorkspacesSetEnvVar(mock.Anything, float32(42)).
		Return(openapi_client.ApiWorkspacesSetEnvVarRequest{ApiService: wsApiMock})
	wsApiMock.EXPECT().WorkspacesSetEnvVarExecute(mock.Anything).Return(nil, nil).Once()
	newWs, err := client.DeployWorkspace(args)

	assert.Nil(t, err, "should be nil")
	assert.Equal(t, newWs.Name, ws.Name, "Should have the same name")
}
