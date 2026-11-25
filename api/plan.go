// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"fmt"

	cserrors "github.com/codesphere-cloud/cs-go/api/errors"
)

// Fetches the workspace plan for a given name.
//
// Returns [NotFound] if no plan with the given Id could be found
func (client *Client) PlanByName(ctx context.Context, name string) (WorkspacePlan, error) {
	plans, err := client.ListWorkspacePlans(ctx)
	if err != nil {
		return WorkspacePlan{}, err
	}

	for _, p := range plans {
		if p.Title == name {
			return p, nil
		}
	}
	return WorkspacePlan{}, cserrors.NotFound(fmt.Sprintf("no team with name %s found", name))
}

func (c *Client) ListWorkspacePlans(ctx context.Context) ([]WorkspacePlan, error) {
	plans, r, err := c.api.MetadataAPI.MetadataGetWorkspacePlans(c.withAuth(ctx)).Execute()
	return plans, cserrors.FormatAPIError(r, err)
}
