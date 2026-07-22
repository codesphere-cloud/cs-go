// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"context"
	"errors"
	"net/http"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/api/openapi_client"
)

var _ = Describe("Organizations", func() {
	var (
		clustersApiMock      *openapi_client.MockClustersAPI
		organizationsApiMock *openapi_client.MockOrganizationsAPI
		client               *api.Client
	)

	BeforeEach(func(ctx SpecContext) {
		clustersApiMock = openapi_client.NewMockClustersAPI(GinkgoT())
		organizationsApiMock = openapi_client.NewMockOrganizationsAPI(GinkgoT())
		apis := openapi_client.APIClient{
			ClustersAPI:      clustersApiMock,
			OrganizationsAPI: organizationsApiMock,
		}
		client = api.NewClientWithCustomDeps(ctx, api.Configuration{}, &apis, mockTime())
	})

	Context("ListOrganizations", func() {
		It("lists organizations", func() {
			expectedOrgs := []api.Organization{
				{Id: "org-1", Name: "fakeOrg1"},
				{Id: "org-2", Name: "fakeOrg2"},
			}

			organizationsApiMock.EXPECT().OrganizationsListOrganizations(mock.Anything).
				Return(openapi_client.ApiOrganizationsListOrganizationsRequest{ApiService: organizationsApiMock})
			organizationsApiMock.EXPECT().OrganizationsListOrganizationsExecute(mock.Anything).Return(expectedOrgs, nil, nil)
			orgs, err := client.ListOrganizations()

			Expect(err).NotTo(HaveOccurred())
			Expect(orgs).To(Equal(expectedOrgs))
		})

		It("handles HTTP 500 errors when the API fails", func() {
			apiErr := errors.New("internal server error")
			organizationsApiMock.EXPECT().OrganizationsListOrganizations(mock.Anything).
				Return(openapi_client.ApiOrganizationsListOrganizationsRequest{ApiService: organizationsApiMock})
			organizationsApiMock.EXPECT().OrganizationsListOrganizationsExecute(mock.Anything).
				Return(nil, &http.Response{StatusCode: 500}, apiErr)

			orgs, err := client.ListOrganizations()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("internal server error"))
			Expect(orgs).To(BeNil())
		})

		It("handles HTTP 403 forbidden errors", func() {
			apiErr := errors.New("forbidden")
			organizationsApiMock.EXPECT().OrganizationsListOrganizations(mock.Anything).
				Return(openapi_client.ApiOrganizationsListOrganizationsRequest{ApiService: organizationsApiMock})
			organizationsApiMock.EXPECT().OrganizationsListOrganizationsExecute(mock.Anything).
				Return(nil, &http.Response{StatusCode: 403}, apiErr)

			orgs, err := client.ListOrganizations()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("forbidden"))
			Expect(orgs).To(BeNil())
		})
	})

	Context("CreateOrganization", func() {
		It("creates an organization", func() {
			expectedOrg := api.Organization{
				Id:   "new-org-id",
				Name: "newOrg",
			}
			orgName := "newOrg"
			adminEmail := "admin@example.com"

			clustersApiMock.EXPECT().ClustersCreateOrganization(mock.Anything).
				Return(openapi_client.ApiClustersCreateOrganizationRequest{ApiService: clustersApiMock})
			clustersApiMock.EXPECT().ClustersCreateOrganizationExecute(mock.MatchedBy(func(r openapi_client.ApiClustersCreateOrganizationRequest) bool {
				val := reflect.ValueOf(r)
				payloadField := val.FieldByName("clustersCreateOrganizationRequest")
				if !payloadField.IsValid() || payloadField.IsNil() {
					return false
				}
				elem := payloadField.Elem()
				return elem.FieldByName("Name").String() == orgName &&
					elem.FieldByName("AdminEmail").String() == adminEmail
			})).Return(&expectedOrg, nil, nil)

			org, err := client.CreateOrganization(orgName, adminEmail)

			Expect(err).NotTo(HaveOccurred())
			Expect(org.Id).To(Equal(expectedOrg.Id))
			Expect(org.Name).To(Equal(expectedOrg.Name))
		})

		It("handles empty strings and invalid email formats from API errors", func() {
			orgName := ""
			adminEmail := "invalid-email"
			apiErr := errors.New("bad request: invalid payload")

			clustersApiMock.EXPECT().ClustersCreateOrganization(mock.Anything).
				Return(openapi_client.ApiClustersCreateOrganizationRequest{ApiService: clustersApiMock})
			clustersApiMock.EXPECT().ClustersCreateOrganizationExecute(mock.MatchedBy(func(r openapi_client.ApiClustersCreateOrganizationRequest) bool {
				val := reflect.ValueOf(r)
				payloadField := val.FieldByName("clustersCreateOrganizationRequest")
				if !payloadField.IsValid() || payloadField.IsNil() {
					return false
				}
				elem := payloadField.Elem()
				return elem.FieldByName("Name").String() == orgName &&
					elem.FieldByName("AdminEmail").String() == adminEmail
			})).Return(nil, &http.Response{StatusCode: 400}, apiErr)

			org, err := client.CreateOrganization(orgName, adminEmail)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid payload"))
			Expect(org).To(BeNil())
		})

		It("handles context timeouts", func() {
			orgName := "timeoutOrg"
			adminEmail := "timeout@example.com"
			apiErr := context.DeadlineExceeded

			clustersApiMock.EXPECT().ClustersCreateOrganization(mock.Anything).
				Return(openapi_client.ApiClustersCreateOrganizationRequest{ApiService: clustersApiMock})
			clustersApiMock.EXPECT().ClustersCreateOrganizationExecute(mock.Anything).
				Return(nil, nil, apiErr)

			org, err := client.CreateOrganization(orgName, adminEmail)

			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, context.DeadlineExceeded)).To(BeTrue())
			Expect(org).To(BeNil())
		})
	})
})
