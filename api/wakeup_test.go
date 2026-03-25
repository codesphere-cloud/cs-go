// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/api/openapi_client"
	"github.com/codesphere-cloud/cs-go/pkg/testutil"
)

var _ = Describe("WakeUp", func() {
	var (
		wsId      int
		teamId    int
		client    *api.RealClient
		wsApiMock *openapi_client.MockWorkspacesAPI
	)
	BeforeEach(func() {
		wsId = 42
		teamId = 21
		wsApiMock = openapi_client.NewMockWorkspacesAPI(GinkgoT())
	})

	JustBeforeEach(func() {
		mockTime := testutil.MockTime()
		apis := openapi_client.APIClient{
			WorkspacesAPI: wsApiMock,
		}
		client = api.NewClientWithCustomDeps(context.TODO(), api.Configuration{}, &apis, mockTime)
	})

	Context("WakeUpWorkspace", func() {
		It("should return error if GetWorkspace fails", func() {
			wsApiMock.EXPECT().WorkspacesGetWorkspace(mock.Anything, float32(wsId)).
				Return(openapi_client.ApiWorkspacesGetWorkspaceRequest{ApiService: wsApiMock})
			wsApiMock.EXPECT().WorkspacesGetWorkspaceExecute(mock.Anything).Return(nil, nil, fmt.Errorf("api error"))

			err := client.WakeUpWorkspace(wsId, "", "", 0)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get workspace"))
		})

		Context("when GetWorkspace succeeds", func() {
			var workspace api.Workspace

			BeforeEach(func() {
				workspace = api.Workspace{
					Id:     wsId,
					TeamId: teamId,
					Name:   "test-workspace",
				}

				wsApiMock.EXPECT().WorkspacesGetWorkspace(mock.Anything, float32(wsId)).
					Return(openapi_client.ApiWorkspacesGetWorkspaceRequest{ApiService: wsApiMock})
				wsApiMock.EXPECT().WorkspacesGetWorkspaceExecute(mock.Anything).Return(&workspace, nil, nil)
			})

			Context("when workspace is not running", func() {
				BeforeEach(func() {
					wsApiMock.EXPECT().WorkspacesGetWorkspaceStatus(mock.Anything, float32(wsId)).
						Return(openapi_client.ApiWorkspacesGetWorkspaceStatusRequest{ApiService: wsApiMock}).Once()
					wsApiMock.EXPECT().WorkspacesGetWorkspaceStatusExecute(mock.Anything).
						Return(&api.WorkspaceStatus{IsRunning: false}, nil, nil).Once()
				})

				Context("when scaling succeeds", func() {
					BeforeEach(func() {
						wsApiMock.EXPECT().WorkspacesUpdateWorkspace(mock.Anything, float32(wsId)).
							Return(openapi_client.ApiWorkspacesUpdateWorkspaceRequest{ApiService: wsApiMock})
						wsApiMock.EXPECT().WorkspacesUpdateWorkspaceExecute(mock.Anything).Return(nil, nil)
						wsApiMock.EXPECT().WorkspacesGetWorkspaceStatus(mock.Anything, float32(wsId)).
							Return(openapi_client.ApiWorkspacesGetWorkspaceStatusRequest{ApiService: wsApiMock}).Once()
						wsApiMock.EXPECT().WorkspacesGetWorkspaceStatusExecute(mock.Anything).
							Return(&api.WorkspaceStatus{IsRunning: true}, nil, nil).Once()
					})

					Context("when deploying the default landscape", func() {
						BeforeEach(func() {
							wsApiMock.EXPECT().WorkspacesDeployLandscape(mock.Anything, float32(wsId)).
								Return(openapi_client.ApiWorkspacesDeployLandscapeRequest{ApiService: wsApiMock})
						})

						It("should wake up the workspace by scaling to 1 replica", func() {
							wsApiMock.EXPECT().WorkspacesDeployLandscapeExecute(mock.Anything).Return(nil, nil)

							err := client.WakeUpWorkspace(wsId, "", "", 0)

							Expect(err).ToNot(HaveOccurred())
						})

						It("should sync landscape when SyncLandscape flag is set", func() {
							wsApiMock.EXPECT().WorkspacesDeployLandscapeExecute(mock.Anything).Return(nil, nil)

							err := client.WakeUpWorkspace(wsId, "", "", 10)

							Expect(err).ToNot(HaveOccurred())
						})

						It("should return error if DeployLandscape fails", func() {
							wsApiMock.EXPECT().WorkspacesDeployLandscapeExecute(mock.Anything).Return(nil, fmt.Errorf("deploy error"))

							err := client.WakeUpWorkspace(wsId, "", "", 10)

							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("failed to deploy landscape"))
						})
					})

					Context("when deploying a custom profile", func() {
						BeforeEach(func() {
							wsApiMock.EXPECT().WorkspacesDeployLandscape1(mock.Anything, float32(wsId), "prod").
								Return(openapi_client.ApiWorkspacesDeployLandscape1Request{ApiService: wsApiMock})
							wsApiMock.EXPECT().WorkspacesDeployLandscape1Execute(mock.Anything).Return(nil, nil)
						})

						It("should sync landscape with custom profile", func() {
							err := client.WakeUpWorkspace(wsId, "", "prod", 10)

							Expect(err).ToNot(HaveOccurred())
						})
					})
				})

				It("should return error if ScaleWorkspace fails", func() {
					wsApiMock.EXPECT().WorkspacesUpdateWorkspace(mock.Anything, float32(wsId)).
						Return(openapi_client.ApiWorkspacesUpdateWorkspaceRequest{ApiService: wsApiMock})
					wsApiMock.EXPECT().WorkspacesUpdateWorkspaceExecute(mock.Anything).Return(nil, fmt.Errorf("scale error"))

					err := client.WakeUpWorkspace(wsId, "", "", 0)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to scale workspace"))
				})
			})

			Context("when workspace is already running", func() {
				BeforeEach(func() {
					wsApiMock.EXPECT().WorkspacesGetWorkspaceStatus(mock.Anything, float32(wsId)).
						Return(openapi_client.ApiWorkspacesGetWorkspaceStatusRequest{ApiService: wsApiMock})
					wsApiMock.EXPECT().WorkspacesGetWorkspaceStatusExecute(mock.Anything).
						Return(&api.WorkspaceStatus{IsRunning: true}, nil, nil)
					wsApiMock.EXPECT().WorkspacesDeployLandscape(mock.Anything, float32(wsId)).
						Return(openapi_client.ApiWorkspacesDeployLandscapeRequest{ApiService: wsApiMock})
					wsApiMock.EXPECT().WorkspacesDeployLandscapeExecute(mock.Anything).Return(nil, nil)
				})

				It("should return early if workspace is already running", func() {
					err := client.WakeUpWorkspace(wsId, "", "", 0)

					Expect(err).ToNot(HaveOccurred())
				})

				It("should sync landscape even when workspace is already running", func() {
					err := client.WakeUpWorkspace(wsId, "", "", 10)

					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
	})
})
