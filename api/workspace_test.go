// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

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
	wsApiMock.EXPECT().WorkspacesGetWorkspaceStatus(mock.Anything, workspaceId).
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

			wsApiMock.EXPECT().WorkspacesListWorkspaces(mock.Anything, teamId).
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
			wsApiMock.EXPECT().WorkspacesGitPull(mock.Anything, ws.Id).
				Return(openapi_client.ApiWorkspacesGitPullRequest{
					ApiService: wsApiMock,
				})
			wsApiMock.EXPECT().WorkspacesGitPullExecute(mock.Anything).Return(nil, nil)

			err := client.GitPull(ws.Id, "", "")

			Expect(err).NotTo(HaveOccurred())
		})

		It("sends request to pull with remote and origin when specified", func() {
			wsApiMock.EXPECT().WorkspacesGitPull2(mock.Anything, ws.Id, "origin", "my-branch").
				Return(openapi_client.ApiWorkspacesGitPull2Request{
					ApiService: wsApiMock,
				})
			wsApiMock.EXPECT().WorkspacesGitPull2Execute(mock.Anything).Return(nil, nil)

			err := client.GitPull(ws.Id, "origin", "my-branch")

			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("DeployWorkspace", func() {
		Context("when CreateWorkspace succeeds immediately", func() {
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
				wsApiMock.EXPECT().WorkspacesSetEnvVar(mock.Anything, 0).
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

			It("Does not call SetEnvVar when env vars are empty", func() {
				newWs, err := client.DeployWorkspace(
					api.DeployWorkspaceArgs{
						Timeout: 1 * time.Millisecond,
						EnvVars: map[string]string{},
					},
				)

				Expect(err).NotTo(HaveOccurred())
				Expect(newWs.Name).To(Equal(ws.Name))
			})
		})

		Context("retry logic", func() {
			retryableErr := fmt.Errorf("codesphere API returned error 502 (Bad Gateway): temporarily unavailable")
			nonRetryableErr := fmt.Errorf("codesphere API returned error 400 (Bad Request): invalid input")

			It("Retries on retryable error and succeeds", func() {
				mockWorkspaceStatus(wsApiMock, ws.Id, true)
				wsApiMock.EXPECT().WorkspacesCreateWorkspace(mock.Anything).
					Return(openapi_client.ApiWorkspacesCreateWorkspaceRequest{ApiService: wsApiMock})
				wsApiMock.EXPECT().WorkspacesCreateWorkspaceExecute(mock.Anything).Return(nil, nil, retryableErr).Once()
				wsApiMock.EXPECT().WorkspacesCreateWorkspaceExecute(mock.Anything).Return(&ws, nil, nil).Once()

				newWs, err := client.DeployWorkspace(
					api.DeployWorkspaceArgs{Timeout: 1 * time.Millisecond},
				)

				Expect(err).NotTo(HaveOccurred())
				Expect(newWs.Name).To(Equal(ws.Name))
			})

			It("Fails immediately on non-retryable error", func() {
				wsApiMock.EXPECT().WorkspacesCreateWorkspace(mock.Anything).
					Return(openapi_client.ApiWorkspacesCreateWorkspaceRequest{ApiService: wsApiMock})
				wsApiMock.EXPECT().WorkspacesCreateWorkspaceExecute(mock.Anything).Return(nil, nil, nonRetryableErr).Once()

				newWs, err := client.DeployWorkspace(
					api.DeployWorkspaceArgs{Timeout: 1 * time.Millisecond},
				)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("error 400")))
				Expect(newWs).To(BeNil())
			})

			It("Returns error after all retries exhausted", func() {
				wsApiMock.EXPECT().WorkspacesCreateWorkspace(mock.Anything).
					Return(openapi_client.ApiWorkspacesCreateWorkspaceRequest{ApiService: wsApiMock})
				wsApiMock.EXPECT().WorkspacesCreateWorkspaceExecute(mock.Anything).Return(nil, nil, retryableErr).Times(5)

				newWs, err := client.DeployWorkspace(
					api.DeployWorkspaceArgs{Timeout: 1 * time.Millisecond},
				)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("error 502")))
				Expect(newWs).To(BeNil())
			})
		})

		Context("when WaitForWorkspaceRunning times out", func() {
			It("Returns workspace and timeout error", func() {
				mockWorkspaceStatus(wsApiMock, ws.Id, false, false)
				wsApiMock.EXPECT().WorkspacesCreateWorkspace(mock.Anything).
					Return(openapi_client.ApiWorkspacesCreateWorkspaceRequest{ApiService: wsApiMock})
				wsApiMock.EXPECT().WorkspacesCreateWorkspaceExecute(mock.Anything).Return(&ws, nil, nil)

				newWs, err := client.DeployWorkspace(
					api.DeployWorkspaceArgs{Timeout: 1 * time.Second},
				)

				Expect(err).To(BeAssignableToTypeOf(&errors.TimedOutError{}))
				Expect(newWs).NotTo(BeNil())
				Expect(newWs.Name).To(Equal(ws.Name))
			})
		})

		Context("when SetEnvVarOnWorkspace fails", func() {
			It("Returns workspace and error", func() {
				mockWorkspaceStatus(wsApiMock, ws.Id, true)
				wsApiMock.EXPECT().WorkspacesCreateWorkspace(mock.Anything).
					Return(openapi_client.ApiWorkspacesCreateWorkspaceRequest{ApiService: wsApiMock})
				wsApiMock.EXPECT().WorkspacesCreateWorkspaceExecute(mock.Anything).Return(&ws, nil, nil)

				envErr := fmt.Errorf("codesphere API returned error 500 (Internal Server Error): failed to set env")
				wsApiMock.EXPECT().WorkspacesSetEnvVar(mock.Anything, 0).
					Return(openapi_client.ApiWorkspacesSetEnvVarRequest{ApiService: wsApiMock})
				wsApiMock.EXPECT().WorkspacesSetEnvVarExecute(mock.Anything).Return(nil, envErr).Once()

				newWs, err := client.DeployWorkspace(
					api.DeployWorkspaceArgs{
						Timeout: 1 * time.Millisecond,
						EnvVars: map[string]string{"key": "val"},
					},
				)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("error 500")))
				Expect(newWs).NotTo(BeNil())
				Expect(newWs.Name).To(Equal(ws.Name))
			})
		})
	})
})

var _ = Describe("UpdateWorkspace", func() {
	var (
		server    *httptest.Server
		client    *api.Client
		method    string
		path      string
		body      map[string]any
		status    int
		replyBody string
	)

	BeforeEach(func() {
		status = http.StatusOK
		replyBody = ""
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			method = r.Method
			path = r.URL.Path
			body = map[string]any{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_, _ = w.Write([]byte(replyBody))
		}))
		baseUrl, err := url.Parse(server.URL + "/api")
		Expect(err).NotTo(HaveOccurred())
		client = api.NewClient(context.TODO(), api.Configuration{BaseUrl: baseUrl, Token: "token"})
	})

	AfterEach(func() {
		server.Close()
	})

	It("sends a PATCH with only the restricted flag", func() {
		restricted := false
		err := client.UpdateWorkspace(42, api.UpdateWorkspaceArgs{Restricted: &restricted})

		Expect(err).NotTo(HaveOccurred())
		Expect(method).To(Equal(http.MethodPatch))
		Expect(path).To(Equal("/api/workspaces/42"))
		Expect(body).To(Equal(map[string]any{"restricted": false}))
	})

	It("sends name and plan changes", func() {
		name := "docs"
		plan := 8
		err := client.UpdateWorkspace(42, api.UpdateWorkspaceArgs{Name: &name, PlanId: &plan})

		Expect(err).NotTo(HaveOccurred())
		Expect(body).To(Equal(map[string]any{"name": "docs", "planId": float64(8)}))
	})

	It("returns the formatted API error", func() {
		status = http.StatusNotFound
		replyBody = `{"status":404,"title":"Workspace is not found.","traceId":"abc"}`
		restricted := true
		err := client.UpdateWorkspace(42, api.UpdateWorkspaceArgs{Restricted: &restricted})

		Expect(err).To(MatchError(ContainSubstring("error 404 (Workspace is not found.)")))
	})
})
