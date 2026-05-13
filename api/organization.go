package api

import (
	cserrors "github.com/codesphere-cloud/cs-go/api/errors"
)

func (c *Client) ListOrganizations() ([]Organization, error) {
	organizations, r, err := c.api.OrganizationsAPI.OrganizationsListOrganizations(c.ctx).Execute()
	return organizations, cserrors.FormatAPIError(r, err)
}
