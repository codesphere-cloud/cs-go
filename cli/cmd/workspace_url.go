// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/codesphere-cloud/cs-go/api"
)

// ConstructWorkspaceServiceURL constructs a URL for accessing a workspace service
// Format: https://${WORKSPACE_ID}-${PORT}.${DEV_DOMAIN}${PATH}
func ConstructWorkspaceServiceURL(workspace api.Workspace, port int, path string) (string, error) {
	if workspace.DevDomain == nil {
		return "", fmt.Errorf("workspace %d does not have a development domain configured", workspace.Id)
	}

	devDomain := *workspace.DevDomain

	// extract just the base domain (e.g., "dev.codesphere.com")
	if strings.Contains(devDomain, ".") {
		parts := strings.SplitN(devDomain, ".", 2)
		if len(parts) == 2 {
			// Check if the first part starts with workspace ID followed by a hyphen
			prefix := parts[0]
			wsIdStr := strconv.Itoa(workspace.Id)
			if strings.HasPrefix(prefix, wsIdStr+"-") {
				// Strip the workspace-port prefix and use the base domain
				devDomain = parts[1]
			}
		}
	}

	url := fmt.Sprintf("https://%d-%d.%s%s", workspace.Id, port, devDomain, path)
	return url, nil
}
