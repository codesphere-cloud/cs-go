// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/codesphere-cloud/cs-go/api"
)

// ConstructWorkspaceServiceURL constructs a URL for accessing a workspace service
// Format: https://${WORKSPACE_ID}-${PORT}.${DEV_DOMAIN}${PATH}
func ConstructWorkspaceServiceURL(workspace api.Workspace, port int, path string) (string, error) {
	if workspace.DevDomain == nil {
		return "", fmt.Errorf("workspace %d does not have a development domain configured", workspace.Id)
	}

	url := fmt.Sprintf("https://%d-%d.%s%s", workspace.Id, port, *workspace.DevDomain, path)
	return url, nil
}
