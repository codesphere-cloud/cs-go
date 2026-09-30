// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	cserrors "github.com/codesphere-cloud/cs-go/api/errors"
	openapi "github.com/codesphere-cloud/cs-go/api/openapi_client"
)

func (c *Client) ListOrganizations() ([]Organization, error) {
	orgs, r, err := c.api.OrganizationsAPI.OrganizationsListOrganizations(c.ctx).Execute()
	if err != nil {
		return nil, cserrors.FormatAPIError(r, err)
	}

	res := make([]Organization, len(orgs))
	copy(res, orgs)
	return res, nil
}

func (c *Client) CreateOrganization(name string, adminEmail string) (*Organization, error) {
	req := openapi.NewClustersCreateOrganizationRequest(name, adminEmail)
	org, r, err := c.api.ClustersAPI.ClustersCreateOrganization(c.ctx).ClustersCreateOrganizationRequest(*req).Execute()
	if err != nil {
		return nil, cserrors.FormatAPIError(r, err)
	}

	return org, nil
}

func (c *Client) ListOrganizationMembers(orgId string) ([]OrganizationMember, error) {
	members, r, err := c.api.OrganizationsAPI.OrganizationsListOrgMembers(c.ctx, orgId).Execute()
	return members, cserrors.FormatAPIError(r, err)
}

func (c *Client) AddOrganizationMember(orgId, email, role string) error {
	req := openapi.NewOrganizationsAddOrgMemberRequest(email, role)
	r, err := c.api.OrganizationsAPI.OrganizationsAddOrgMember(c.ctx, orgId).OrganizationsAddOrgMemberRequest(*req).Execute()
	return cserrors.FormatAPIError(r, err)
}

func (c *Client) RemoveOrganizationMember(orgId string, userId int) error {
	r, err := c.api.OrganizationsAPI.OrganizationsRemoveOrgMember(c.ctx, orgId, userId).Execute()
	return cserrors.FormatAPIError(r, err)
}

func (c *Client) ChangeOrganizationMemberRole(orgId string, userId int, role string) error {
	req := openapi.NewOrganizationsChangeOrgRoleRequest(role)
	r, err := c.api.OrganizationsAPI.OrganizationsChangeOrgRole(c.ctx, orgId, userId).OrganizationsChangeOrgRoleRequest(*req).Execute()
	return cserrors.FormatAPIError(r, err)
}
