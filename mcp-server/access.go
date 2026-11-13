// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"errors"
	"fmt"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type UpdateWorkspaceArgs struct {
	WorkspaceId *int    `json:"workspaceId,omitempty" jsonschema:"ID of the workspace to change. Defaults to the current workspace (CS_WORKSPACE_ID)."`
	Restricted  *bool   `json:"restricted,omitempty" jsonschema:"false: anyone can open the workspace's dev domain and every landscape path on it without signing in. true: visitors are redirected to the Codesphere sign-in and only team members get in. Omit to leave access unchanged."`
	Name        *string `json:"name,omitempty" jsonschema:"New workspace name. Omit to keep the current name."`
	PlanId      *int    `json:"planId,omitempty" jsonschema:"New plan ID for the workspace itself (see list_workspace_plans). This is not the plan of landscape services, which is set per service in ci.yml. Omit to keep the current plan."`
}

// UpdateWorkspaceResult is the workspace after the change.
type UpdateWorkspaceResult struct {
	Workspace  api.Workspace `json:"workspace"`
	Restricted bool          `json:"restricted"`
	DevDomain  string        `json:"devDomainUrl,omitempty"`
	Access     string        `json:"access"`
}

const updateWorkspaceDescription = `Change who can reach a Codesphere workspace, and optionally its name or plan.

This changes who can reach the workspace: with restricted=false, anyone on the internet can open the workspace's dev domain (https://<devDomain>/) and every public landscape path on it without signing in. With restricted=true, visitors are redirected to the Codesphere sign-in.

The user is asked to approve every call. Before calling, tell the user what you are about to change and why. This is the only sanctioned way for an agent to change workspace access: raw API calls (e.g. curl -X PATCH .../workspaces/<id>) are refused.

When to use: get_landscape_status warns that a public service sits in a restricted workspace and the user agreed to make it public, or the user asked to restrict or rename the workspace or change its plan.

Arguments: workspaceId (optional, defaults to CS_WORKSPACE_ID), restricted, name, planId. Set at least one of restricted, name or planId; omitted fields stay unchanged.

Returns the workspace as it is after the change, read back from the API.`

func RegisterAccessTools(server *mcp.Server, deps Deps) {
	destructive := true
	openWorld := true
	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_workspace",
		Title:       "Update workspace access, name or plan (asks the user)",
		Description: updateWorkspaceDescription,
		Annotations: &mcp.ToolAnnotations{
			Title:           "Update workspace access, name or plan",
			DestructiveHint: &destructive,
			IdempotentHint:  true,
			OpenWorldHint:   &openWorld,
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, args UpdateWorkspaceArgs) (*mcp.CallToolResult, any, error) {
		res, err := updateWorkspace(deps, args)
		if err != nil {
			return nil, nil, err
		}
		return nil, res, nil
	})
}

func updateWorkspace(deps Deps, args UpdateWorkspaceArgs) (*UpdateWorkspaceResult, error) {
	if args.Restricted == nil && args.Name == nil && args.PlanId == nil {
		return nil, errors.New("nothing to change: set at least one of restricted, name or planId")
	}
	if args.PlanId != nil && *args.PlanId < 0 {
		return nil, errors.New("planId must be non-negative")
	}
	wsId, err := deps.workspaceId(args.WorkspaceId)
	if err != nil {
		return nil, err
	}

	err = deps.Client.UpdateWorkspace(wsId, api.UpdateWorkspaceArgs{
		Restricted: args.Restricted,
		Name:       args.Name,
		PlanId:     args.PlanId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update workspace %d: %w", wsId, err)
	}

	// The PATCH answers without a body, read the result back.
	ws, err := deps.Client.GetWorkspace(wsId)
	if err != nil {
		return nil, fmt.Errorf("workspace %d was updated, but reading it back failed: %w", wsId, err)
	}

	res := &UpdateWorkspaceResult{Workspace: ws, Restricted: ws.Restricted}
	if ws.DevDomain != nil && *ws.DevDomain != "" {
		res.DevDomain = "https://" + *ws.DevDomain + "/"
	}
	if ws.Restricted {
		res.Access = "restricted: visitors are redirected to the Codesphere sign-in"
	} else {
		res.Access = "public: reachable without sign-in"
	}
	return res, nil
}
