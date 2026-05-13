// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	cserrors "github.com/codesphere-cloud/cs-go/api/errors"
)

func (c *Client) ListOrganizations() ([]Organization, error) {
	organizations, r, err := c.api.OrganizationsAPI.OrganizationsListOrganizations(c.ctx).Execute()
	return organizations, cserrors.FormatAPIError(r, err)
}
