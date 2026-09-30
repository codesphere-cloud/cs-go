// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package update

import (
	"fmt"

	"github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type UpdateOrganizationMemberCmd struct {
	cmd           *cobra.Command
	Opts          UpdateOrganizationMemberOpts
	ClientFactory func(shared.RootOptions) (shared.Client, error)
}

type UpdateOrganizationMemberOpts struct {
	shared.RootOptions
	UserId int
	Role   string
}

func AddUpdateOrganizationMemberCmd(parent *cobra.Command, opts shared.RootOptions) {
	c := UpdateOrganizationMemberCmd{
		cmd: &cobra.Command{
			Use:     "organization-member",
			Aliases: []string{"org-member"},
			Short:   "Change organization member role",
			Long:    "Change organization member role. Select the organization using --org or CS_ORG_ID.",
			Example: io.FormatExampleCommands("update organization-member", []io.Example{
				{Cmd: "--org <orgId> -u <userId> -r admin", Desc: "Change organization member role"},
			}),
		},
		Opts:          UpdateOrganizationMemberOpts{RootOptions: opts},
		ClientFactory: func(opts shared.RootOptions) (shared.Client, error) { return opts.NewClient() },
	}
	c.cmd.Flags().IntVarP(&c.Opts.UserId, "user", "u", -1, "Organization member user ID")
	_ = c.cmd.MarkFlagRequired("user")
	c.cmd.Flags().StringVarP(&c.Opts.Role, "role", "r", "", "Organization role (admin, member)")
	_ = c.cmd.MarkFlagRequired("role")

	c.cmd.RunE = c.RunE
	shared.AddCmd(parent, c.cmd)
}

func (c *UpdateOrganizationMemberCmd) RunE(_ *cobra.Command, _ []string) error {
	orgId, err := shared.RequireOrganizationID(c.Opts.RootOptions)
	if err != nil {
		return err
	}
	client, err := c.ClientFactory(c.Opts.RootOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}
	return c.ChangeOrganizationMemberRole(client, orgId, c.Opts.UserId, c.Opts.Role)
}

func (c *UpdateOrganizationMemberCmd) ChangeOrganizationMemberRole(client shared.Client, orgId string, userId int, role string) error {
	if err := shared.ValidateUserID(userId); err != nil {
		return err
	}
	if err := shared.ValidateOrganizationRole(role); err != nil {
		return err
	}

	if err := client.ChangeOrganizationMemberRole(orgId, userId, role); err != nil {
		return fmt.Errorf("failed to change organization member role: %w", err)
	}
	return nil
}
