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
