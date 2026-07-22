// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"context"

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

	BeforeEach(func() {
		clustersApiMock = openapi_client.NewMockClustersAPI(GinkgoT())
		organizationsApiMock = openapi_client.NewMockOrganizationsAPI(GinkgoT())
		apis := openapi_client.APIClient{
			ClustersAPI:      clustersApiMock,
			OrganizationsAPI: organizationsApiMock,
		}
		client = api.NewClientWithCustomDeps(context.TODO(), api.Configuration{}, &apis, mockTime())
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
			clustersApiMock.EXPECT().ClustersCreateOrganizationExecute(mock.Anything).Return(&expectedOrg, nil, nil)

			org, err := client.CreateOrganization(orgName, adminEmail)

			Expect(err).NotTo(HaveOccurred())
			Expect(org.Id).To(Equal(expectedOrg.Id))
			Expect(org.Name).To(Equal(expectedOrg.Name))
		})
	})
})
