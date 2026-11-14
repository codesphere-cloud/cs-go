// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"context"
	"errors"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/api/openapi_client"
)

var _ = Describe("ManagedServices", func() {
	var (
		msApiMock *openapi_client.MockManagedServicesAPI
		client    *api.Client
	)

	BeforeEach(func() {
		msApiMock = openapi_client.NewMockManagedServicesAPI(GinkgoT())
		apis := openapi_client.APIClient{ManagedServicesAPI: msApiMock}
		client = api.NewClientWithCustomDeps(context.TODO(), api.Configuration{}, &apis, mockTime())
	})

	Context("ListManagedServices", func() {
		var req openapi_client.ApiManagedServicesListRequest

		BeforeEach(func() {
			msApiMock.EXPECT().ManagedServicesList(mock.Anything).
				Return(openapi_client.ApiManagedServicesListRequest{ApiService: msApiMock})
		})

		It("lists managed services of a team", func() {
			expected := []api.ManagedService{{Id: "id-1", Name: "db1"}, {Id: "id-2", Name: "db2"}}
			msApiMock.EXPECT().ManagedServicesListExecute(mock.Anything).
				Run(func(r openapi_client.ApiManagedServicesListRequest) { req = r }).
				Return(expected, nil, nil)

			services, err := client.ListManagedServices(42, false)

			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(Equal(expected))
			Expect(requestField(req, "team").Elem().Int()).To(BeEquivalentTo(42))
		})

		// The generated client rejects list requests without includeDeleted before sending them.
		It("sends includeDeleted=false when deleted services are not requested", func() {
			msApiMock.EXPECT().ManagedServicesListExecute(mock.Anything).
				Run(func(r openapi_client.ApiManagedServicesListRequest) { req = r }).
				Return(nil, nil, nil)

			_, err := client.ListManagedServices(42, false)

			Expect(err).NotTo(HaveOccurred())
			includeDeleted := requestField(req, "includeDeleted")
			Expect(includeDeleted.IsNil()).To(BeFalse(), "includeDeleted must always be set")
			Expect(includeDeleted.Elem().Bool()).To(BeFalse())
		})

		It("sends includeDeleted=true when deleted services are requested", func() {
			msApiMock.EXPECT().ManagedServicesListExecute(mock.Anything).
				Run(func(r openapi_client.ApiManagedServicesListRequest) { req = r }).
				Return(nil, nil, nil)

			_, err := client.ListManagedServices(42, true)

			Expect(err).NotTo(HaveOccurred())
			Expect(requestField(req, "includeDeleted").Elem().Bool()).To(BeTrue())
		})

		It("returns an error when listing managed services fails", func() {
			msApiMock.EXPECT().ManagedServicesListExecute(mock.Anything).Return(nil, nil, errors.New("boom"))

			_, err := client.ListManagedServices(42, false)

			Expect(err).To(MatchError(ContainSubstring("boom")))
		})
	})

	Context("GetManagedService", func() {
		BeforeEach(func() {
			msApiMock.EXPECT().ManagedServicesGetDetails(mock.Anything, "id-1").
				Return(openapi_client.ApiManagedServicesGetDetailsRequest{ApiService: msApiMock})
		})

		It("gets managed service details", func() {
			expected := map[string]interface{}{"id": "id-1"}
			msApiMock.EXPECT().ManagedServicesGetDetailsExecute(mock.Anything).Return(expected, nil, nil)

			details, err := client.GetManagedService("id-1")

			Expect(err).NotTo(HaveOccurred())
			Expect(details).To(Equal(expected))
		})

		It("returns an error when getting managed service details fails", func() {
			msApiMock.EXPECT().ManagedServicesGetDetailsExecute(mock.Anything).Return(nil, nil, errors.New("boom"))

			_, err := client.GetManagedService("id-1")

			Expect(err).To(MatchError(ContainSubstring("boom")))
		})
	})

	Context("ListManagedServiceProviders", func() {
		BeforeEach(func() {
			msApiMock.EXPECT().ManagedServicesListProviders(mock.Anything).
				Return(openapi_client.ApiManagedServicesListProvidersRequest{ApiService: msApiMock})
		})

		It("lists managed service providers", func() {
			expected := []api.ManagedServiceProvider{{}}
			msApiMock.EXPECT().ManagedServicesListProvidersExecute(mock.Anything).Return(expected, nil, nil)

			providers, err := client.ListManagedServiceProviders()

			Expect(err).NotTo(HaveOccurred())
			Expect(providers).To(Equal(expected))
		})

		It("returns an error when listing providers fails", func() {
			msApiMock.EXPECT().ManagedServicesListProvidersExecute(mock.Anything).Return(nil, nil, errors.New("boom"))

			_, err := client.ListManagedServiceProviders()

			Expect(err).To(MatchError(ContainSubstring("boom")))
		})
	})

	Context("CreateManagedService", func() {
		BeforeEach(func() {
			msApiMock.EXPECT().ManagedServicesCreate(mock.Anything).
				Return(openapi_client.ApiManagedServicesCreateRequest{ApiService: msApiMock})
		})

		It("creates a managed service from the given arguments", func() {
			expected := &api.ManagedService{Id: "id-1", Name: "db1"}
			var req openapi_client.ApiManagedServicesCreateRequest
			msApiMock.EXPECT().ManagedServicesCreateExecute(mock.Anything).
				Run(func(r openapi_client.ApiManagedServicesCreateRequest) { req = r }).
				Return(expected, nil, nil)

			svc, err := client.CreateManagedService(api.CreateManagedServiceArgs{Name: "db1", TeamId: 1})

			Expect(err).NotTo(HaveOccurred())
			Expect(svc).To(Equal(expected))
			Expect(requestField(req, "managedServicesCreateRequest", "Name").String()).To(Equal("db1"))
			Expect(requestField(req, "managedServicesCreateRequest", "TeamId").Int()).To(BeEquivalentTo(1))
		})

		It("returns an error when creating a managed service fails", func() {
			msApiMock.EXPECT().ManagedServicesCreateExecute(mock.Anything).Return(nil, nil, errors.New("boom"))

			_, err := client.CreateManagedService(api.CreateManagedServiceArgs{Name: "db1", TeamId: 1})

			Expect(err).To(MatchError(ContainSubstring("boom")))
		})
	})

	Context("UpdateManagedService", func() {
		BeforeEach(func() {
			msApiMock.EXPECT().ManagedServicesUpdate(mock.Anything, "id-1").
				Return(openapi_client.ApiManagedServicesUpdateRequest{ApiService: msApiMock})
		})

		It("updates a managed service with the given arguments", func() {
			expected := &api.ManagedService{Id: "id-1", Name: "renamed"}
			var req openapi_client.ApiManagedServicesUpdateRequest
			msApiMock.EXPECT().ManagedServicesUpdateExecute(mock.Anything).
				Run(func(r openapi_client.ApiManagedServicesUpdateRequest) { req = r }).
				Return(expected, nil, nil)

			name := "renamed"
			svc, err := client.UpdateManagedService("id-1", api.UpdateManagedServiceArgs{Name: &name})

			Expect(err).NotTo(HaveOccurred())
			Expect(svc).To(Equal(expected))
			Expect(requestField(req, "managedServicesUpdateRequest", "Name").Elem().String()).To(Equal("renamed"))
		})

		It("returns an error when updating a managed service fails", func() {
			msApiMock.EXPECT().ManagedServicesUpdateExecute(mock.Anything).Return(nil, nil, errors.New("boom"))

			_, err := client.UpdateManagedService("id-1", api.UpdateManagedServiceArgs{})

			Expect(err).To(MatchError(ContainSubstring("boom")))
		})
	})

	Context("DeleteManagedService", func() {
		BeforeEach(func() {
			msApiMock.EXPECT().ManagedServicesDelete(mock.Anything, "id-1").
				Return(openapi_client.ApiManagedServicesDeleteRequest{ApiService: msApiMock})
		})

		It("deletes a managed service", func() {
			msApiMock.EXPECT().ManagedServicesDeleteExecute(mock.Anything).Return(nil, nil)

			Expect(client.DeleteManagedService("id-1")).To(Succeed())
		})

		It("returns an error when deleting a managed service fails", func() {
			msApiMock.EXPECT().ManagedServicesDeleteExecute(mock.Anything).Return(nil, errors.New("boom"))

			Expect(client.DeleteManagedService("id-1")).To(MatchError(ContainSubstring("boom")))
		})
	})

	Context("ScheduleManagedServiceBackup", func() {
		BeforeEach(func() {
			msApiMock.EXPECT().ManagedServicesScheduleBackup(mock.Anything, "id-1").
				Return(openapi_client.ApiManagedServicesScheduleBackupRequest{ApiService: msApiMock})
		})

		It("schedules a backup", func() {
			expected := &api.ManagedServiceBackup{Id: "backup-1"}
			msApiMock.EXPECT().ManagedServicesScheduleBackupExecute(mock.Anything).Return(expected, nil, nil)

			backup, err := client.ScheduleManagedServiceBackup("id-1")

			Expect(err).NotTo(HaveOccurred())
			Expect(backup).To(Equal(expected))
		})

		It("returns an error when scheduling a backup fails", func() {
			msApiMock.EXPECT().ManagedServicesScheduleBackupExecute(mock.Anything).Return(nil, nil, errors.New("boom"))

			_, err := client.ScheduleManagedServiceBackup("id-1")

			Expect(err).To(MatchError(ContainSubstring("boom")))
		})
	})

	Context("NewManagedServiceProviderRef", func() {
		It("builds a provider reference from name and version", func() {
			version := "v1"

			ref := api.NewManagedServiceProviderRef("postgres", &version)

			Expect(ref.Name).To(Equal("postgres"))
			Expect(ref.Version).To(HaveValue(Equal("v1")))
		})

		It("leaves the version unset when none is given", func() {
			ref := api.NewManagedServiceProviderRef("postgres", nil)

			Expect(ref.Version).To(BeNil())
		})
	})
})

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
