// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package shared

import "fmt"

// RequireOrganizationID resolves the flag/environment value and requires an organization scope.
func RequireOrganizationID(opts RootOptions) (string, error) {
	id, err := opts.GetOrgId()
	if err != nil {
		return "", fmt.Errorf("failed to get organization ID: %w", err)
	}
	if id == "" {
		return "", fmt.Errorf("organization ID not set, use --org or CS_ORG_ID to set it")
	}
	return id, nil
}

func ValidateOrganizationRole(role string) error {
	if role != "admin" && role != "member" {
		return fmt.Errorf("invalid organization role %q: must be admin or member", role)
	}
	return nil
}
