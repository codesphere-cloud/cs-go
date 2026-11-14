// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"context"
	"errors"
	"reflect"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/api/openapi_client"
	mcpserver "github.com/codesphere-cloud/cs-go/mcp-server"
)

var _ = Describe("ManagedServiceTools", func() {
	var (
		msApiMock *openapi_client.MockManagedServicesAPI
		session   *mcp.ClientSession
	)

	BeforeEach(func() {
		msApiMock = openapi_client.NewMockManagedServicesAPI(GinkgoT())
		apis := openapi_client.APIClient{ManagedServicesAPI: msApiMock}
		client := api.NewClientWithCustomDeps(context.Background(), api.Configuration{}, &apis, api.NewMockTime(GinkgoT()))

		server := mcp.NewServer(&mcp.Implementation{Name: "codesphere-mcp-test", Version: "v0.0.0"}, nil)
		mcpserver.RegisterManagedServiceTools(server, client)

		clientTransport, serverTransport := mcp.NewInMemoryTransports()
		_, err := server.Connect(context.Background(), serverTransport, nil)
		Expect(err).NotTo(HaveOccurred())

		mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.0"}, nil)
		session, err = mcpClient.Connect(context.Background(), clientTransport, nil)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(session.Close)
	})

	callTool := func(name string, args map[string]any) *mcp.CallToolResult {
		res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
		Expect(err).NotTo(HaveOccurred())
		return res
	}

	Context("list_managed_services", func() {
		var req openapi_client.ApiManagedServicesListRequest

		BeforeEach(func() {
			msApiMock.EXPECT().ManagedServicesList(mock.Anything).
				Return(openapi_client.ApiManagedServicesListRequest{ApiService: msApiMock})
		})

		It("lists the managed services of a team without deleted ones by default", func() {
			services := []api.ManagedService{*managedService("id-1", "db1"), *managedService("id-2", "db2")}
			msApiMock.EXPECT().ManagedServicesListExecute(mock.Anything).
				Run(func(r openapi_client.ApiManagedServicesListRequest) { req = r }).
				Return(services, nil, nil)

			res := callTool("list_managed_services", map[string]any{"teamId": 42})

			Expect(res.IsError).To(BeFalse())
			Expect(res.StructuredContent).To(HaveKeyWithValue("items", ConsistOf(
				HaveKeyWithValue("id", "id-1"),
				HaveKeyWithValue("id", "id-2"),
			)))
			Expect(requestField(req, "team").Elem().Int()).To(BeEquivalentTo(42))
			Expect(requestField(req, "includeDeleted").IsNil()).To(BeFalse(), "includeDeleted must always be set")
			Expect(requestField(req, "includeDeleted").Elem().Bool()).To(BeFalse())
		})

		It("includes deleted managed services when asked to", func() {
			msApiMock.EXPECT().ManagedServicesListExecute(mock.Anything).
				Run(func(r openapi_client.ApiManagedServicesListRequest) { req = r }).
				Return(nil, nil, nil)

			res := callTool("list_managed_services", map[string]any{"teamId": 42, "includeDeleted": true})

			Expect(res.IsError).To(BeFalse())
			Expect(requestField(req, "includeDeleted").Elem().Bool()).To(BeTrue())
		})

		It("returns an empty items list when the team has no managed services", func() {
			msApiMock.EXPECT().ManagedServicesListExecute(mock.Anything).Return(nil, nil, nil)

			res := callTool("list_managed_services", map[string]any{"teamId": 42})

			Expect(res.IsError).To(BeFalse())
			Expect(res.StructuredContent).To(HaveKeyWithValue("items", Equal([]any{})))
		})

		It("reports API errors as tool errors", func() {
			msApiMock.EXPECT().ManagedServicesListExecute(mock.Anything).Return(nil, nil, errors.New("boom"))

			res := callTool("list_managed_services", map[string]any{"teamId": 42})

			Expect(res.IsError).To(BeTrue())
			Expect(res.Content).To(ContainElement(HaveField("Text", ContainSubstring("boom"))))
		})
	})

	Context("get_managed_service", func() {
		It("returns the details of the managed service", func() {
			msApiMock.EXPECT().ManagedServicesGetDetails(mock.Anything, "id-1").
				Return(openapi_client.ApiManagedServicesGetDetailsRequest{ApiService: msApiMock})
			msApiMock.EXPECT().ManagedServicesGetDetailsExecute(mock.Anything).
				Return(map[string]interface{}{"id": "id-1", "hostname": "db1.internal"}, nil, nil)

			res := callTool("get_managed_service", map[string]any{"serviceId": "id-1"})

			Expect(res.IsError).To(BeFalse())
			Expect(res.StructuredContent).To(HaveKeyWithValue("hostname", "db1.internal"))
		})
	})

	Context("list_managed_service_providers", func() {
		It("lists the available providers", func() {
			providers := []api.ManagedServiceProvider{{
				ManagedServicesListProviders200ResponseInnerAnyOf: &openapi_client.ManagedServicesListProviders200ResponseInnerAnyOf{
					Name:    "postgres",
					Version: "v1",
				},
			}}
			msApiMock.EXPECT().ManagedServicesListProviders(mock.Anything).
				Return(openapi_client.ApiManagedServicesListProvidersRequest{ApiService: msApiMock})
			msApiMock.EXPECT().ManagedServicesListProvidersExecute(mock.Anything).Return(providers, nil, nil)

			res := callTool("list_managed_service_providers", map[string]any{})

			Expect(res.IsError).To(BeFalse())
			Expect(res.StructuredContent).To(HaveKeyWithValue("items", ConsistOf(HaveKeyWithValue("name", "postgres"))))
		})
	})

	Context("create_managed_service", func() {
		var req openapi_client.ApiManagedServicesCreateRequest

		BeforeEach(func() {
			msApiMock.EXPECT().ManagedServicesCreate(mock.Anything).
				Return(openapi_client.ApiManagedServicesCreateRequest{ApiService: msApiMock}).Maybe()
		})

		It("creates the managed service from the tool arguments", func() {
			msApiMock.EXPECT().ManagedServicesCreateExecute(mock.Anything).
				Run(func(r openapi_client.ApiManagedServicesCreateRequest) { req = r }).
				Return(managedService("id-1", "db1"), nil, nil)

			res := callTool("create_managed_service", map[string]any{
				"teamId":          7,
				"name":            "db1",
				"providerName":    "postgres",
				"providerVersion": "v1",
				"plan":            map[string]any{"id": 2, "parameters": map[string]any{"storage": 10}},
				"config":          map[string]any{"database": "app"},
				"secrets":         map[string]any{"password": "s3cret"},
			})

			Expect(res.IsError).To(BeFalse())
			Expect(res.StructuredContent).To(HaveKeyWithValue("id", "id-1"))
			body := requestField(req, "managedServicesCreateRequest")
			Expect(requestField(body, "TeamId").Int()).To(BeEquivalentTo(7))
			Expect(requestField(body, "Name").String()).To(Equal("db1"))
			Expect(requestField(body, "Provider", "Name").String()).To(Equal("postgres"))
			Expect(requestField(body, "Provider", "Version").Elem().String()).To(Equal("v1"))
			Expect(requestField(body, "Plan", "Id").Int()).To(BeEquivalentTo(2))
			Expect(requestField(body, "Plan", "Parameters").MapIndex(reflect.ValueOf("storage")).Int()).To(BeEquivalentTo(10))
			Expect(requestField(body, "Config").MapIndex(reflect.ValueOf("database")).Elem().String()).To(Equal("app"))
			Expect(requestField(body, "Secrets").MapIndex(reflect.ValueOf("password")).Elem().String()).To(Equal("s3cret"))
		})

		It("leaves the provider version unset when none is given", func() {
			msApiMock.EXPECT().ManagedServicesCreateExecute(mock.Anything).
				Run(func(r openapi_client.ApiManagedServicesCreateRequest) { req = r }).
				Return(managedService("id-1", "db1"), nil, nil)

			res := callTool("create_managed_service", map[string]any{
				"teamId":       7,
				"name":         "db1",
				"providerName": "postgres",
				"plan":         map[string]any{"id": 2, "parameters": map[string]any{}},
				"config":       map[string]any{},
				"secrets":      map[string]any{},
			})

			Expect(res.IsError).To(BeFalse())
			Expect(requestField(req, "managedServicesCreateRequest", "Provider", "Version").IsNil()).To(BeTrue())
		})

		It("rejects calls without the required arguments", func() {
			res := callTool("create_managed_service", map[string]any{"teamId": 7})

			Expect(res.IsError).To(BeTrue())
			Expect(res.Content).To(ContainElement(HaveField("Text", ContainSubstring("validating"))))
		})
	})

	Context("update_managed_service", func() {
		var req openapi_client.ApiManagedServicesUpdateRequest

		BeforeEach(func() {
			msApiMock.EXPECT().ManagedServicesUpdate(mock.Anything, "id-1").
				Return(openapi_client.ApiManagedServicesUpdateRequest{ApiService: msApiMock})
			msApiMock.EXPECT().ManagedServicesUpdateExecute(mock.Anything).
				Run(func(r openapi_client.ApiManagedServicesUpdateRequest) { req = r }).
				Return(managedService("id-1", "renamed"), nil, nil)
		})

		It("sends only the given fields", func() {
			res := callTool("update_managed_service", map[string]any{"serviceId": "id-1", "name": "renamed", "pause": true})

			Expect(res.IsError).To(BeFalse())
			Expect(res.StructuredContent).To(HaveKeyWithValue("name", "renamed"))
			body := requestField(req, "managedServicesUpdateRequest")
			Expect(requestField(body, "Name").Elem().String()).To(Equal("renamed"))
			Expect(requestField(body, "Pause").Elem().Bool()).To(BeTrue())
			Expect(requestField(body, "Plan").IsNil()).To(BeTrue())
			Expect(requestField(body, "Config").IsNil()).To(BeTrue())
			Expect(requestField(body, "Secrets").IsNil()).To(BeTrue())
			Expect(requestField(body, "Version").IsNil()).To(BeTrue())
		})

		It("sends a new plan when one is given", func() {
			res := callTool("update_managed_service", map[string]any{
				"serviceId": "id-1",
				"plan":      map[string]any{"id": 3, "parameters": map[string]any{"storage": 20}},
			})

			Expect(res.IsError).To(BeFalse())
			plan := requestField(req, "managedServicesUpdateRequest", "Plan")
			Expect(plan.IsNil()).To(BeFalse())
			Expect(requestField(plan, "Id").Int()).To(BeEquivalentTo(3))
			Expect(requestField(plan, "Parameters").MapIndex(reflect.ValueOf("storage")).Int()).To(BeEquivalentTo(20))
		})
	})

	Context("delete_managed_service", func() {
		BeforeEach(func() {
			msApiMock.EXPECT().ManagedServicesDelete(mock.Anything, "id-1").
				Return(openapi_client.ApiManagedServicesDeleteRequest{ApiService: msApiMock})
		})

		It("deletes the managed service", func() {
			msApiMock.EXPECT().ManagedServicesDeleteExecute(mock.Anything).Return(nil, nil)

			res := callTool("delete_managed_service", map[string]any{"serviceId": "id-1"})

			Expect(res.IsError).To(BeFalse())
			Expect(res.StructuredContent).To(HaveKeyWithValue("status", "deleted"))
		})

		It("reports API errors as tool errors", func() {
			msApiMock.EXPECT().ManagedServicesDeleteExecute(mock.Anything).Return(nil, errors.New("boom"))

			res := callTool("delete_managed_service", map[string]any{"serviceId": "id-1"})

			Expect(res.IsError).To(BeTrue())
			Expect(res.Content).To(ContainElement(HaveField("Text", ContainSubstring("boom"))))
		})
	})

	Context("schedule_managed_service_backup", func() {
		It("schedules a backup of the managed service", func() {
			msApiMock.EXPECT().ManagedServicesScheduleBackup(mock.Anything, "id-1").
				Return(openapi_client.ApiManagedServicesScheduleBackupRequest{ApiService: msApiMock})
			msApiMock.EXPECT().ManagedServicesScheduleBackupExecute(mock.Anything).
				Return(&api.ManagedServiceBackup{Id: "backup-1"}, nil, nil)

			res := callTool("schedule_managed_service_backup", map[string]any{"serviceId": "id-1"})

			Expect(res.IsError).To(BeFalse())
			Expect(res.StructuredContent).To(HaveKeyWithValue("id", "backup-1"))
		})
	})
})

// managedService builds a managed service the way the API returns it. Status and backups are
// anyOf types that only serialize when one of their variants is set.
func managedService(id, name string) *api.ManagedService {
	return &api.ManagedService{
		Id:       id,
		Name:     name,
		Provider: openapi_client.ManagedServicesList200ResponseInnerProvider{Name: "postgres", Version: "v1", SchemaVersion: "v1"},
		Backups: openapi_client.ManagedServicesList200ResponseInnerBackups{
			ManagedServicesList200ResponseInnerBackupsAnyOf1: &openapi_client.ManagedServicesList200ResponseInnerBackupsAnyOf1{
				Entries: []api.ManagedServiceBackup{},
			},
		},
		Plan: api.ManagedServicePlan{Id: 2, Parameters: map[string]int{"storage": 10}},
		Status: openapi_client.ManagedServicesList200ResponseInnerStatus{
			ManagedServicesList200ResponseInnerStatusAnyOf: &openapi_client.ManagedServicesList200ResponseInnerStatusAnyOf{State: "running"},
		},
	}
}

// requestField reads a field of a generated request, following pointers along the path. The
// generated client keeps query parameters and request bodies in unexported fields, so tests can
// only reach them through reflection.
func requestField(r any, path ...string) reflect.Value {
	v, ok := r.(reflect.Value)
	if !ok {
		v = reflect.ValueOf(r)
	}
	for _, name := range path {
		v = reflect.Indirect(v).FieldByName(name)
	}
	return v
}
