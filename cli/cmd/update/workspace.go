// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package update

import (
	"errors"
	"fmt"
	"strings"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type UpdateWorkspaceCmd struct {
	cmd           *cobra.Command
	Opts          UpdateWorkspaceOpts
	ClientFactory func(shared.RootOptions) (shared.Client, error)
}

type UpdateWorkspaceOpts struct {
	shared.RootOptions
	Restricted *bool
	Name       *string
	PlanId     *int
}

func AddUpdateWorkspaceCmd(parent *cobra.Command, opts shared.RootOptions) {
	c := UpdateWorkspaceCmd{
		cmd: &cobra.Command{
			Use:   "workspace",
			Short: "Change workspace access, name or plan",
			Long: io.Long(`Change settings of a workspace: who can reach it, its name or its plan.

				Access: a restricted workspace (the default) sends visitors of its dev domain
				and of every landscape path on it to the Codesphere sign-in. --public (or
				--restricted=false) makes it reachable by anyone without signing in; --private
				(or --restricted=true) restricts it again.

				In AI agent sessions access changes need the user's approval: agents use the
				MCP tool update_workspace, which asks the user first.

				The command prints the workspace as it is after the change.`),
			Example: io.FormatExampleCommands("update workspace", []io.Example{
				{Cmd: "--public", Desc: "Make the current workspace (CS_WORKSPACE_ID) reachable without sign-in"},
				{Cmd: "-w 1234 --restricted=false", Desc: "Same as --public, for workspace 1234"},
				{Cmd: "--private", Desc: "Require sign-in again"},
				{Cmd: "--name docs --plan 21", Desc: "Rename the workspace and change its plan"},
			}),
			Args: cobra.NoArgs,
		},
		Opts:          UpdateWorkspaceOpts{RootOptions: opts},
		ClientFactory: func(opts shared.RootOptions) (shared.Client, error) { return opts.NewClient() },
	}
	flags := c.cmd.Flags()
	flags.Bool("restricted", true, "Whether visitors must sign in to Codesphere (--restricted=false makes the workspace public)")
	flags.Bool("public", false, "Make the workspace reachable without sign-in (same as --restricted=false)")
	flags.Bool("private", false, "Require sign-in to reach the workspace (same as --restricted=true)")
	flags.String("name", "", "New workspace name")
	flags.Int("plan", 0, "New workspace plan ID (see 'cs list plans')")
	c.cmd.MarkFlagsMutuallyExclusive("restricted", "public", "private")

	c.cmd.RunE = c.RunE
	shared.AddCmd(parent, c.cmd)
}

func (c *UpdateWorkspaceCmd) RunE(cmd *cobra.Command, _ []string) error {
	if cmd != nil {
		if err := c.Opts.readFlags(cmd.Flags()); err != nil {
			return err
		}
	}
	wsId, err := c.Opts.GetWorkspaceId()
	if err != nil {
		return fmt.Errorf("failed to get workspace ID: %w", err)
	}
	client, err := c.ClientFactory(c.Opts.RootOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}
	ws, err := c.UpdateWorkspace(client, wsId)
	if err != nil {
		return err
	}
	fmt.Print(FormatWorkspaceAccess(ws))
	return nil
}

// readFlags only sets the options for flags given on the command line, so unset
// values stay nil and are left unchanged by the API.
func (o *UpdateWorkspaceOpts) readFlags(flags *pflag.FlagSet) error {
	o.Restricted, o.Name, o.PlanId = nil, nil, nil
	for flag, restricted := range map[string]bool{"public": false, "private": true} {
		if set, err := flags.GetBool(flag); err != nil {
			return err
		} else if set {
			o.Restricted = &restricted
		}
	}
	if flags.Changed("restricted") {
		restricted, err := flags.GetBool("restricted")
		if err != nil {
			return err
		}
		o.Restricted = &restricted
	}
	if flags.Changed("name") {
		name, err := flags.GetString("name")
		if err != nil {
			return err
		}
		o.Name = &name
	}
	if flags.Changed("plan") {
		plan, err := flags.GetInt("plan")
		if err != nil {
			return err
		}
		o.PlanId = &plan
	}
	return nil
}

// UpdateWorkspace applies the requested changes and returns the workspace as read back afterwards.
func (c *UpdateWorkspaceCmd) UpdateWorkspace(client shared.Client, wsId int) (*api.Workspace, error) {
	args := api.UpdateWorkspaceArgs{
		Restricted: c.Opts.Restricted,
		Name:       c.Opts.Name,
		PlanId:     c.Opts.PlanId,
	}
	if args.Restricted == nil && args.Name == nil && args.PlanId == nil {
		return nil, errors.New("nothing to change: set --public, --private, --restricted, --name or --plan")
	}
	if args.PlanId != nil && *args.PlanId < 0 {
		return nil, errors.New("plan ID must be non-negative")
	}

	if err := client.UpdateWorkspace(wsId, args); err != nil {
		err = fmt.Errorf("failed to update workspace %d: %w", wsId, err)
		if args.Restricted != nil && isPermissionError(err) {
			err = fmt.Errorf("%w\nchanging who can reach a workspace needs the user's approval in agent sessions: ask the user, then use the MCP tool update_workspace", err)
		}
		return nil, err
	}

	// The PATCH answers without a body, read the result back.
	ws, err := client.GetWorkspace(wsId)
	if err != nil {
		return nil, fmt.Errorf("workspace %d was updated, but reading it back failed: %w", wsId, err)
	}
	return &ws, nil
}

// FormatWorkspaceAccess describes a workspace's name, plan and who can reach it.
func FormatWorkspaceAccess(ws *api.Workspace) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Workspace %d (%s), plan %d\n", ws.Id, ws.Name, ws.PlanId)
	url := ""
	if ws.DevDomain != nil && *ws.DevDomain != "" {
		url = "https://" + *ws.DevDomain + "/"
	}
	switch {
	case ws.Restricted && url != "":
		fmt.Fprintf(&b, "Access: restricted, visitors of %s are redirected to the Codesphere sign-in\n", url)
	case ws.Restricted:
		fmt.Fprintln(&b, "Access: restricted, visitors are redirected to the Codesphere sign-in")
	case url != "":
		fmt.Fprintf(&b, "Access: public, %s is reachable without sign-in\n", url)
	default:
		fmt.Fprintln(&b, "Access: public, reachable without sign-in")
	}
	return b.String()
}

func isPermissionError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "error 401") || strings.Contains(msg, "error 403")
}
