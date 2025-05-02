package api

import (
	"fmt"

	cserrors "github.com/codesphere-cloud/cs-go/api/errors"
)

// Fetches the workspace plan for a given name.
//
// Returns [NotFound] if no plan with the given Id could be found
func (client *Client) PlanByName(name string) (WorkspacePlan, error) {
	plans, err := client.ListWorkspacePlans()
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

func (c *Client) ListWorkspacePlans() ([]WorkspacePlan, error) {
	plans, _, err := c.api.MetadataAPI.MetadataGetWorkspacePlans(c.ctx).Execute()
	return plans, err
}
