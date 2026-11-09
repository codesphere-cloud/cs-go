// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"context"
	"errors"

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

	It("lists managed services of a team", func() {
		expected := []api.ManagedService{{Id: "id-1", Name: "db1"}, {Id: "id-2", Name: "db2"}}

		msApiMock.EXPECT().ManagedServicesList(mock.Anything).
			Return(openapi_client.ApiManagedServicesListRequest{ApiService: msApiMock})
		msApiMock.EXPECT().ManagedServicesListExecute(mock.Anything).Return(expected, nil, nil)

		services, err := client.ListManagedServices(42, false)

		Expect(err).NotTo(HaveOccurred())
		Expect(services).To(Equal(expected))
	})

	It("returns an error when listing managed services fails", func() {
		msApiMock.EXPECT().ManagedServicesList(mock.Anything).
			Return(openapi_client.ApiManagedServicesListRequest{ApiService: msApiMock})
		msApiMock.EXPECT().ManagedServicesListExecute(mock.Anything).Return(nil, nil, errors.New("boom"))

		_, err := client.ListManagedServices(42, false)

		Expect(err).To(HaveOccurred())
	})

	It("gets managed service details", func() {
		expected := map[string]interface{}{"id": "id-1"}

		msApiMock.EXPECT().ManagedServicesGetDetails(mock.Anything, "id-1").
			Return(openapi_client.ApiManagedServicesGetDetailsRequest{ApiService: msApiMock})
		msApiMock.EXPECT().ManagedServicesGetDetailsExecute(mock.Anything).Return(expected, nil, nil)

		details, err := client.GetManagedService("id-1")

		Expect(err).NotTo(HaveOccurred())
		Expect(details).To(Equal(expected))
	})

	It("lists managed service providers", func() {
		expected := []api.ManagedServiceProvider{{}}

		msApiMock.EXPECT().ManagedServicesListProviders(mock.Anything).
			Return(openapi_client.ApiManagedServicesListProvidersRequest{ApiService: msApiMock})
		msApiMock.EXPECT().ManagedServicesListProvidersExecute(mock.Anything).Return(expected, nil, nil)

		providers, err := client.ListManagedServiceProviders()

		Expect(err).NotTo(HaveOccurred())
		Expect(providers).To(Equal(expected))
	})

	It("creates a managed service", func() {
		expected := &api.ManagedService{Id: "id-1", Name: "db1"}

		msApiMock.EXPECT().ManagedServicesCreate(mock.Anything).
			Return(openapi_client.ApiManagedServicesCreateRequest{ApiService: msApiMock})
		msApiMock.EXPECT().ManagedServicesCreateExecute(mock.Anything).Return(expected, nil, nil)

		svc, err := client.CreateManagedService(api.CreateManagedServiceArgs{Name: "db1", TeamId: 1})

		Expect(err).NotTo(HaveOccurred())
		Expect(svc).To(Equal(expected))
	})

	It("updates a managed service", func() {
		expected := &api.ManagedService{Id: "id-1", Name: "renamed"}

		msApiMock.EXPECT().ManagedServicesUpdate(mock.Anything, "id-1").
			Return(openapi_client.ApiManagedServicesUpdateRequest{ApiService: msApiMock})
		msApiMock.EXPECT().ManagedServicesUpdateExecute(mock.Anything).Return(expected, nil, nil)

		name := "renamed"
		svc, err := client.UpdateManagedService("id-1", api.UpdateManagedServiceArgs{Name: &name})

		Expect(err).NotTo(HaveOccurred())
		Expect(svc).To(Equal(expected))
	})

	It("deletes a managed service", func() {
		msApiMock.EXPECT().ManagedServicesDelete(mock.Anything, "id-1").
			Return(openapi_client.ApiManagedServicesDeleteRequest{ApiService: msApiMock})
		msApiMock.EXPECT().ManagedServicesDeleteExecute(mock.Anything).Return(nil, nil)

		Expect(client.DeleteManagedService("id-1")).To(Succeed())
	})

	It("returns an error when deleting a managed service fails", func() {
		msApiMock.EXPECT().ManagedServicesDelete(mock.Anything, "id-1").
			Return(openapi_client.ApiManagedServicesDeleteRequest{ApiService: msApiMock})
		msApiMock.EXPECT().ManagedServicesDeleteExecute(mock.Anything).Return(nil, errors.New("boom"))

		Expect(client.DeleteManagedService("id-1")).To(HaveOccurred())
	})

	It("schedules a backup", func() {
		expected := &api.ManagedServiceBackup{Id: "backup-1"}

		msApiMock.EXPECT().ManagedServicesScheduleBackup(mock.Anything, "id-1").
			Return(openapi_client.ApiManagedServicesScheduleBackupRequest{ApiService: msApiMock})
		msApiMock.EXPECT().ManagedServicesScheduleBackupExecute(mock.Anything).Return(expected, nil, nil)

		backup, err := client.ScheduleManagedServiceBackup("id-1")

		Expect(err).NotTo(HaveOccurred())
		Expect(backup).To(Equal(expected))
	})
})
