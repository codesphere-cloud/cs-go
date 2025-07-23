// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"time"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/api/errors"
	"github.com/codesphere-cloud/cs-go/api/openapi_client"
)

func mockTime() *api.MockTime {
	currentTime := time.Unix(1746190963, 0)
	m := api.NewMockTime(GinkgoT())
	m.EXPECT().Now().RunAndReturn(func() time.Time {
		return currentTime
	}).Maybe()
	m.EXPECT().Sleep(mock.Anything).Run(func(delay time.Duration) {
		currentTime = currentTime.Add(delay)
	}).Maybe()
	return m
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

var _ = Describe("Workspace", func() {
	var (
		ws        api.Workspace
		wsApiMock *openapi_client.MockWorkspacesAPI
		client    *api.Client
	)

	BeforeEach(func() {
		wsApiMock = openapi_client.NewMockWorkspacesAPI(GinkgoT())
		mockTime := mockTime()
		apis := openapi_client.APIClient{
			WorkspacesAPI: wsApiMock,
		}
		client = api.NewClientWithCustomDeps(context.TODO(), api.Configuration{}, &apis, mockTime)
	})

	Context("ListWorkspace", func() {
		It("lists workspaces", func() {
			expectedWorkspaces := []api.Workspace{
				{Id: 0, Name: "fakeForTeam0"},
				{Id: 1, Name: "fakeForTeam1"},
			}
			teamId := 42

			wsApiMock.EXPECT().WorkspacesListWorkspaces(mock.Anything, float32(teamId)).
				Return(openapi_client.ApiWorkspacesListWorkspacesRequest{ApiService: wsApiMock})
			wsApiMock.EXPECT().WorkspacesListWorkspacesExecute(mock.Anything).Return(expectedWorkspaces, nil, nil)
			workspaces, err := client.ListWorkspaces(teamId)

			Expect(err).NotTo(HaveOccurred())
			Expect(workspaces).To(Equal(expectedWorkspaces))
		})
	})

	Context("WaitForWorkspaceRunning", func() {

		BeforeEach(func() {
			ws = api.Workspace{
				Id: 0, Name: "fakeWorkspace",
			}
		})

		It("Success when already running", func() {
			mockWorkspaceStatus(wsApiMock, ws.Id, true)

			err := client.WaitForWorkspaceRunning(&ws, 1*time.Millisecond)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Returns an error on timeout", func() {
			mockWorkspaceStatus(wsApiMock, ws.Id, false, false)
			err := client.WaitForWorkspaceRunning(&ws, 1*time.Second)

			Expect(err).To(BeAssignableToTypeOf(&errors.TimedOutError{}))
		})

		It("Success on retry", func() {
			mockWorkspaceStatus(wsApiMock, ws.Id, false, true)
			err := client.WaitForWorkspaceRunning(&ws, 1*time.Second)

			Expect(err).NotTo(HaveOccurred())
		})
	})
	Context("GitPull", func() {
		It("sends request to pull without remote and origin", func() {
			wsApiMock.EXPECT().WorkspacesGitPull(mock.Anything, float32(ws.Id)).
				Return(openapi_client.ApiWorkspacesGitPullRequest{
					ApiService: wsApiMock,
				})
			wsApiMock.EXPECT().WorkspacesGitPullExecute(mock.Anything).Return(nil, nil)

			err := client.GitPull(ws.Id, "", "")

			Expect(err).NotTo(HaveOccurred())
		})

		It("sends request to pull with remote and origin when specified", func() {
			wsApiMock.EXPECT().WorkspacesGitPull2(mock.Anything, float32(ws.Id), "origin", "my-branch").
				Return(openapi_client.ApiWorkspacesGitPull2Request{
					ApiService: wsApiMock,
				})
			wsApiMock.EXPECT().WorkspacesGitPull2Execute(mock.Anything).Return(nil, nil)

			err := client.GitPull(ws.Id, "origin", "my-branch")

			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("DeployWorkspace", func() {
		BeforeEach(func() {
			mockWorkspaceStatus(wsApiMock, ws.Id, true)
			wsApiMock.EXPECT().WorkspacesCreateWorkspace(mock.Anything).
				Return(openapi_client.ApiWorkspacesCreateWorkspaceRequest{ApiService: wsApiMock})
			wsApiMock.EXPECT().WorkspacesCreateWorkspaceExecute(mock.Anything).Return(&ws, nil, nil)
		})

		It("Returns workspace on success", func() {
			newWs, err := client.DeployWorkspace(
				api.DeployWorkspaceArgs{Timeout: 1 * time.Millisecond},
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(newWs.Name).To(Equal(ws.Name))
		})

		It("Calls SetEnvVar endpoint when env vars are set", func() {
			wsApiMock.EXPECT().WorkspacesSetEnvVar(mock.Anything, float32(0)).
				Return(openapi_client.ApiWorkspacesSetEnvVarRequest{ApiService: wsApiMock})
			wsApiMock.EXPECT().WorkspacesSetEnvVarExecute(mock.Anything).Return(nil, nil).Once()

			newWs, err := client.DeployWorkspace(
				api.DeployWorkspaceArgs{
					Timeout: 1 * time.Millisecond,
					EnvVars: map[string]string{
						"foo":  "bar",
						"some": "thing",
					},
				},
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(newWs.Name).To(Equal(ws.Name))
		})
	})
})
