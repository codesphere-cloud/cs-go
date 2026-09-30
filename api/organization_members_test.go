// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"errors"
	"github.com/codesphere-cloud/cs-go/api"
	openapi "github.com/codesphere-cloud/cs-go/api/openapi_client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"net/http"
	"reflect"
)

var _ = Describe("Organization members API", func() {
	const orgId = "8316ee2f-87f8-424e-b925-7382dc50d662"
	var service *openapi.MockOrganizationsAPI
	var client *api.Client
	BeforeEach(func(ctx SpecContext) {
		service = openapi.NewMockOrganizationsAPI(GinkgoT())
		client = api.NewClientWithCustomDeps(ctx, api.Configuration{}, &openapi.APIClient{OrganizationsAPI: service}, mockTime())
	})
	Context("AddOrganizationMember", func() {
		BeforeEach(func() {
			service.EXPECT().OrganizationsAddOrgMember(mock.Anything, orgId).Return(openapi.ApiOrganizationsAddOrgMemberRequest{ApiService: service}).Once()
		})
		It("succeeds", func() {
			service.EXPECT().OrganizationsAddOrgMemberExecute(mock.MatchedBy(func(r openapi.ApiOrganizationsAddOrgMemberRequest) bool {
				payload := reflect.ValueOf(r).FieldByName("organizationsAddOrgMemberRequest")
				if payload.IsNil() {
					return false
				}
				payload = payload.Elem()
				return payload.FieldByName("Email").String() == "user@example.com" && payload.FieldByName("Role").String() == "admin"
			})).Return(nil, nil).Once()
			Expect(client.AddOrganizationMember(orgId, "user@example.com", "admin")).To(Succeed())
		})
		It("reports API failures", func() {
			service.EXPECT().OrganizationsAddOrgMemberExecute(mock.Anything).Return(&http.Response{StatusCode: 403}, errors.New("forbidden")).Once()
			Expect(client.AddOrganizationMember(orgId, "user@example.com", "admin")).To(MatchError(ContainSubstring("forbidden")))
		})
	})
	Context("RemoveOrganizationMember", func() {
		BeforeEach(func() {
			service.EXPECT().OrganizationsRemoveOrgMember(mock.Anything, orgId, 0).Return(openapi.ApiOrganizationsRemoveOrgMemberRequest{ApiService: service}).Once()
		})
		It("succeeds", func() {
			service.EXPECT().OrganizationsRemoveOrgMemberExecute(mock.Anything).Return(nil, nil).Once()
			Expect(client.RemoveOrganizationMember(orgId, 0)).To(Succeed())
		})
		It("reports API failures", func() {
			service.EXPECT().OrganizationsRemoveOrgMemberExecute(mock.Anything).Return(&http.Response{StatusCode: 403}, errors.New("forbidden")).Once()
			Expect(client.RemoveOrganizationMember(orgId, 0)).To(MatchError(ContainSubstring("forbidden")))
		})
	})
	Context("ChangeOrganizationMemberRole", func() {
		BeforeEach(func() {
			service.EXPECT().OrganizationsChangeOrgRole(mock.Anything, orgId, 0).Return(openapi.ApiOrganizationsChangeOrgRoleRequest{ApiService: service}).Once()
		})
		It("succeeds", func() {
			service.EXPECT().OrganizationsChangeOrgRoleExecute(mock.MatchedBy(func(r openapi.ApiOrganizationsChangeOrgRoleRequest) bool {
				payload := reflect.ValueOf(r).FieldByName("organizationsChangeOrgRoleRequest")
				if payload.IsNil() {
					return false
				}
				payload = payload.Elem()
				return payload.FieldByName("Role").String() == "member"
			})).Return(nil, nil).Once()
			Expect(client.ChangeOrganizationMemberRole(orgId, 0, "member")).To(Succeed())
		})
		It("reports API failures", func() {
			service.EXPECT().OrganizationsChangeOrgRoleExecute(mock.Anything).Return(&http.Response{StatusCode: 403}, errors.New("forbidden")).Once()
			Expect(client.ChangeOrganizationMemberRole(orgId, 0, "member")).To(MatchError(ContainSubstring("forbidden")))
		})
	})
	Context("ListOrganizationMembers", func() {
		BeforeEach(func() {
			service.EXPECT().OrganizationsListOrgMembers(mock.Anything, orgId).Return(openapi.ApiOrganizationsListOrgMembersRequest{ApiService: service}).Once()
		})
		It("returns organization members", func() {
			members := []api.OrganizationMember{{UserId: 0, OrganizationId: orgId, Role: "admin"}}
			service.EXPECT().OrganizationsListOrgMembersExecute(mock.Anything).Return(members, nil, nil).Once()
			result, err := client.ListOrganizationMembers(orgId)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(members))
		})
		It("reports API failures", func() {
			service.EXPECT().OrganizationsListOrgMembersExecute(mock.Anything).Return(nil, &http.Response{StatusCode: 403}, errors.New("forbidden")).Once()
			result, err := client.ListOrganizationMembers(orgId)
			Expect(err).To(MatchError(ContainSubstring("forbidden")))
			Expect(result).To(BeNil())
		})
	})
})
