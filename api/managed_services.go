// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	cserrors "github.com/codesphere-cloud/cs-go/api/errors"
	openapi "github.com/codesphere-cloud/cs-go/api/openapi_client"
)

// ListManagedServices lists managed services of a team.
func (c *Client) ListManagedServices(teamId int, includeDeleted bool) ([]ManagedService, error) {
	req := c.api.ManagedServicesAPI.ManagedServicesList(c.ctx).Team(teamId)
	if includeDeleted {
		req = req.IncludeDeleted(true)
	}
	services, r, err := req.Execute()
	if err != nil {
		return nil, cserrors.FormatAPIError(r, err)
	}
	return services, nil
}

// GetManagedService returns the details of a single managed service.
func (c *Client) GetManagedService(id string) (map[string]interface{}, error) {
	details, r, err := c.api.ManagedServicesAPI.ManagedServicesGetDetails(c.ctx, id).Execute()
	if err != nil {
		return nil, cserrors.FormatAPIError(r, err)
	}
	return details, nil
}

// ListManagedServiceProviders lists the available managed service providers.
func (c *Client) ListManagedServiceProviders() ([]ManagedServiceProvider, error) {
	providers, r, err := c.api.ManagedServicesAPI.ManagedServicesListProviders(c.ctx).Execute()
	if err != nil {
		return nil, cserrors.FormatAPIError(r, err)
	}
	return providers, nil
}

// CreateManagedService creates a new managed service.
func (c *Client) CreateManagedService(args CreateManagedServiceArgs) (*ManagedService, error) {
	svc, r, err := c.api.ManagedServicesAPI.ManagedServicesCreate(c.ctx).ManagedServicesCreateRequest(args).Execute()
	if err != nil {
		return nil, cserrors.FormatAPIError(r, err)
	}
	return svc, nil
}

// UpdateManagedService updates an existing managed service.
func (c *Client) UpdateManagedService(id string, args UpdateManagedServiceArgs) (*ManagedService, error) {
	svc, r, err := c.api.ManagedServicesAPI.ManagedServicesUpdate(c.ctx, id).ManagedServicesUpdateRequest(args).Execute()
	if err != nil {
		return nil, cserrors.FormatAPIError(r, err)
	}
	return svc, nil
}

// DeleteManagedService deletes a managed service.
func (c *Client) DeleteManagedService(id string) error {
	r, err := c.api.ManagedServicesAPI.ManagedServicesDelete(c.ctx, id).Execute()
	if err != nil {
		return cserrors.FormatAPIError(r, err)
	}
	return nil
}

// ScheduleManagedServiceBackup schedules a backup for a managed service.
func (c *Client) ScheduleManagedServiceBackup(id string) (*ManagedServiceBackup, error) {
	backup, r, err := c.api.ManagedServicesAPI.ManagedServicesScheduleBackup(c.ctx, id).Execute()
	if err != nil {
		return nil, cserrors.FormatAPIError(r, err)
	}
	return backup, nil
}

// NewManagedServiceProviderRef builds the provider reference for a create request.
func NewManagedServiceProviderRef(name string, version *string) openapi.ManagedServicesCreateRequestProvider {
	return openapi.ManagedServicesCreateRequestProvider{Name: name, Version: version}
}
