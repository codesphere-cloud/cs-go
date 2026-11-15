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

The user is asked to approve every call: the server shows the change to the user through the MCP client and only applies it after the user accepts. Clients that cannot ask the user (no elicitation support) cannot use this tool. Before calling, tell the user what you are about to change and why. This is the only sanctioned way for an agent to change workspace access: raw API calls (e.g. curl -X PATCH .../workspaces/<id>) are refused.

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
		wsId, err := validateUpdate(deps, args)
		if err != nil {
			return nil, nil, err
		}
		// The tool annotations alone do not make a client ask the user, so the
		// server asks itself and applies nothing without an explicit accept.
		answer, asked := req.Params.InputResponses[approvalInputId]
		if !asked {
			if !canAskUser(req.Session) {
				return nil, nil, errors.New("update_workspace needs the user's approval, but this MCP client cannot ask the user (no elicitation support); nothing was changed. Ask the user to make the change in Codesphere or run `cs update workspace` themselves")
			}
			return approvalRequest(wsId, args), nil, nil
		}
		if err := checkApproval(answer); err != nil {
			return nil, nil, err
		}
		res, err := applyUpdate(deps, wsId, args)
		if err != nil {
			return nil, nil, err
		}
		return nil, res, nil
	})
}

// validateUpdate checks the arguments and resolves the workspace ID.
func validateUpdate(deps Deps, args UpdateWorkspaceArgs) (int, error) {
	if args.Restricted == nil && args.Name == nil && args.PlanId == nil {
		return -1, errors.New("nothing to change: set at least one of restricted, name or planId")
	}
	if args.PlanId != nil && *args.PlanId < 0 {
		return -1, errors.New("planId must be non-negative")
	}
	return deps.workspaceId(args.WorkspaceId)
}

// applyUpdate changes the workspace and reads it back.
func applyUpdate(deps Deps, wsId int, args UpdateWorkspaceArgs) (*UpdateWorkspaceResult, error) {
	err := deps.Client.UpdateWorkspace(wsId, api.UpdateWorkspaceArgs{
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

// approvalInputId identifies the approval prompt in the tool's input requests.
const approvalInputId = "approve"

// canAskUser reports whether the client declared elicitation support.
func canAskUser(session *mcp.ServerSession) bool {
	if session == nil {
		return false
	}
	params := session.InitializeParams()
	return params != nil && params.Capabilities != nil && params.Capabilities.Elicitation != nil
}

// approvalRequest asks the client to show the change to the user and to call
// the tool again with the user's answer (multi round-trip, SEP-2322). For
// clients on older protocol versions the SDK sends elicitation/create instead.
func approvalRequest(wsId int, args UpdateWorkspaceArgs) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		InputRequests: mcp.InputRequestMap{
			approvalInputId: &mcp.ElicitParams{
				Message:         approvalMessage(wsId, args),
				RequestedSchema: map[string]any{"type": "object", "properties": map[string]any{}},
			},
		},
	}
}

// checkApproval fails closed: anything but an explicit accept is an error.
func checkApproval(answer mcp.InputResponse) error {
	res, ok := answer.(*mcp.ElicitResult)
	if !ok || res == nil {
		return errors.New("the approval answer is not an elicitation result; nothing was changed")
	}
	if res.Action != "accept" {
		return fmt.Errorf("the user did not approve the change (%s); nothing was changed", res.Action)
	}
	return nil
}

// approvalMessage describes the requested change for the user.
func approvalMessage(wsId int, args UpdateWorkspaceArgs) string {
	msg := fmt.Sprintf("Approve these changes to Codesphere workspace %d?", wsId)
	if args.Restricted != nil {
		if *args.Restricted {
			msg += "\n- Restrict access: visitors are redirected to the Codesphere sign-in and only team members get in."
		} else {
			msg += "\n- Make it public: anyone on the internet can open its dev domain and every public landscape path without signing in."
		}
	}
	if args.Name != nil {
		msg += fmt.Sprintf("\n- Rename it to %q.", *args.Name)
	}
	if args.PlanId != nil {
		msg += fmt.Sprintf("\n- Change its plan to plan %d (this can change what it costs).", *args.PlanId)
	}
	return msg
}
