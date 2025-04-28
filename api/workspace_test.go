package api_test

import (
	"testing"
	"time"

	"context"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/api/errors"
	"github.com/codesphere-cloud/cs-go/api/openapi_client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getTestingClient(t *testing.T) (*api.Client, *openapi_client.MockWorkspacesAPI) {
	wsApiMock := openapi_client.NewMockWorkspacesAPI(t)
	apis := openapi_client.APIClient{
		WorkspacesAPI: wsApiMock,
	}
	return api.NewClientWithCustomApi(context.TODO(), api.Configuration{}, &apis), wsApiMock
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

	wsApiMock.EXPECT().WorkspacesGetWorkspaceStatus(mock.Anything, float32(0)).
		Return(openapi_client.ApiWorkspacesGetWorkspaceStatusRequest{ApiService: wsApiMock})
	wsApiMock.EXPECT().WorkspacesGetWorkspaceStatusExecute(mock.Anything).Return(&api.WorkspaceStatus{
		IsRunning: true,
	}, nil, nil)

	err := client.WaitForWorkspaceRunning(
		&ws,
		api.WaitForWorkspaceRunningOptions{Timeout: 1 * time.Millisecond, Delay: 1 * time.Millisecond},
	)

	assert.Nil(t, err, "should be nil")
}

func TestWaitForWorkspaceRunningTimeout(t *testing.T) {
	client, wsApiMock := getTestingClient(t)

	ws := api.Workspace{
		Id: 0, Name: "fakeWorkspace",
	}

	wsApiMock.EXPECT().WorkspacesGetWorkspaceStatus(mock.Anything, float32(0)).
		Return(openapi_client.ApiWorkspacesGetWorkspaceStatusRequest{ApiService: wsApiMock})
	wsApiMock.EXPECT().WorkspacesGetWorkspaceStatusExecute(mock.Anything).Return(&api.WorkspaceStatus{
		IsRunning: false,
	}, nil, nil)

	err := client.WaitForWorkspaceRunning(
		&ws,
		api.WaitForWorkspaceRunningOptions{Timeout: 1 * time.Millisecond, Delay: 1 * time.Millisecond},
	)

	assert.IsType(t, err, &errors.TimedOutError{}, "expected timeout error")
}

func TestWaitForWorkspaceRunningOnRetry(t *testing.T) {
	client, wsApiMock := getTestingClient(t)

	ws := api.Workspace{
		Id: 42, Name: "fakeWorkspace",
	}

	wsApiMock.EXPECT().WorkspacesGetWorkspaceStatus(mock.Anything, float32(42)).
		Return(openapi_client.ApiWorkspacesGetWorkspaceStatusRequest{ApiService: wsApiMock})
	wsApiMock.EXPECT().WorkspacesGetWorkspaceStatusExecute(mock.Anything).Return(&api.WorkspaceStatus{
		IsRunning: false,
	}, nil, nil).Once()
	wsApiMock.EXPECT().WorkspacesGetWorkspaceStatusExecute(mock.Anything).Return(&api.WorkspaceStatus{
		IsRunning: true,
	}, nil, nil).Once()

	err := client.WaitForWorkspaceRunning(
		&ws,
		api.WaitForWorkspaceRunningOptions{Timeout: 10 * time.Millisecond, Delay: 1 * time.Millisecond},
	)

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

	wsApiMock.EXPECT().WorkspacesGetWorkspaceStatus(mock.Anything, float32(42)).
		Return(openapi_client.ApiWorkspacesGetWorkspaceStatusRequest{ApiService: wsApiMock})
	wsApiMock.EXPECT().WorkspacesGetWorkspaceStatusExecute(mock.Anything).Return(&api.WorkspaceStatus{
		IsRunning: true,
	}, nil, nil)

	err := client.DeployWorkspace(
		api.DeployWorkspaceArgs{Timeout: 1 * time.Millisecond},
	)

	assert.Nil(t, err, "should be nil")
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

	wsApiMock.EXPECT().WorkspacesCreateWorkspace(mock.Anything).
		Return(openapi_client.ApiWorkspacesCreateWorkspaceRequest{ApiService: wsApiMock})
	wsApiMock.EXPECT().WorkspacesCreateWorkspaceExecute(mock.Anything).Return(&ws, nil, nil)

	wsApiMock.EXPECT().WorkspacesGetWorkspaceStatus(mock.Anything, float32(42)).
		Return(openapi_client.ApiWorkspacesGetWorkspaceStatusRequest{ApiService: wsApiMock})
	wsApiMock.EXPECT().WorkspacesGetWorkspaceStatusExecute(mock.Anything).Return(&api.WorkspaceStatus{
		IsRunning: true,
	}, nil, nil)

	wsApiMock.EXPECT().WorkspacesSetEnvVar(mock.Anything, float32(42)).
		Return(openapi_client.ApiWorkspacesSetEnvVarRequest{ApiService: wsApiMock})
	wsApiMock.EXPECT().WorkspacesSetEnvVarExecute(mock.Anything).Return(nil, nil).Once()

	err := client.DeployWorkspace(args)

	assert.Nil(t, err, "should be nil")
}
