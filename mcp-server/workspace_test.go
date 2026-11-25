// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"context"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/api/openapi_client"
	mcpserver "github.com/codesphere-cloud/cs-go/mcp-server"
)

var _ = Describe("Workspace tools", func() {
	var (
		wsApiMock *openapi_client.MockWorkspacesAPI
		session   *mcp.ClientSession
	)

	BeforeEach(func() {
		wsApiMock = openapi_client.NewMockWorkspacesAPI(GinkgoT())
		client := api.NewClientWithCustomDeps(
			api.Configuration{Token: "secret-token"},
			&openapi_client.APIClient{WorkspacesAPI: wsApiMock},
			api.NewMockTime(GinkgoT()),
		)

		server := mcp.NewServer(&mcp.Implementation{Name: "codesphere-mcp-test", Version: "v0.0.0"}, nil)
		mcpserver.RegisterWorkspaceTools(server, client)

		serverTransport, clientTransport := mcp.NewInMemoryTransports()
		serverSession, err := server.Connect(context.Background(), serverTransport, nil)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(serverSession.Close)

		mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.0"}, nil)
		session, err = mcpClient.Connect(context.Background(), clientTransport, nil)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(session.Close)
	})

	It("calls the API with the authenticated handler context", func() {
		wsApiMock.EXPECT().WorkspacesDeleteWorkspace(mock.MatchedBy(func(ctx context.Context) bool {
			return ctx.Value(openapi_client.ContextAccessToken) == "secret-token"
		}), 42).Return(openapi_client.ApiWorkspacesDeleteWorkspaceRequest{ApiService: wsApiMock})
		wsApiMock.EXPECT().WorkspacesDeleteWorkspaceExecute(mock.Anything).Return(nil, nil)

		res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
			Name:      "delete_workspace",
			Arguments: map[string]any{"workspaceId": 42},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(res.IsError).To(BeFalse())
	})

	It("cancels the in-flight API request when the MCP client cancels the tool call", func() {
		var apiCtx context.Context
		requestStarted := make(chan struct{})
		requestEnded := make(chan error, 1)
		// Unblocks the fake request at cleanup, so a missing cancellation fails
		// the assertion below instead of hanging the session shutdown.
		release := make(chan struct{})
		DeferCleanup(func() { close(release) })

		wsApiMock.EXPECT().WorkspacesDeleteWorkspace(mock.Anything, 42).
			RunAndReturn(func(ctx context.Context, _ int) openapi_client.ApiWorkspacesDeleteWorkspaceRequest {
				apiCtx = ctx
				return openapi_client.ApiWorkspacesDeleteWorkspaceRequest{ApiService: wsApiMock}
			})
		wsApiMock.EXPECT().WorkspacesDeleteWorkspaceExecute(mock.Anything).
			RunAndReturn(func(openapi_client.ApiWorkspacesDeleteWorkspaceRequest) (*http.Response, error) {
				// Simulate a slow HTTP request that only returns once its context is done.
				close(requestStarted)
				select {
				case <-apiCtx.Done():
				case <-release:
				}
				requestEnded <- apiCtx.Err()
				return nil, apiCtx.Err()
			})

		callCtx, cancelCall := context.WithCancel(context.Background())
		DeferCleanup(cancelCall)
		callResult := make(chan error, 1)
		go func() {
			defer GinkgoRecover()
			_, err := session.CallTool(callCtx, &mcp.CallToolParams{
				Name:      "delete_workspace",
				Arguments: map[string]any{"workspaceId": 42},
			})
			callResult <- err
		}()

		Eventually(requestStarted).Should(BeClosed())
		cancelCall()

		Eventually(callResult).Should(Receive(MatchError(context.Canceled)))
		Eventually(requestEnded).Should(Receive(MatchError(context.Canceled)))
	})
})
