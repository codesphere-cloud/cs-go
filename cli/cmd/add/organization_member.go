// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"fmt"
	"github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
	"net/mail"
)

type AddOrganizationMemberCmd struct {
	cmd           *cobra.Command
	Opts          AddOrganizationMemberOpts
	ClientFactory func(shared.RootOptions) (shared.Client, error)
}

type AddOrganizationMemberOpts struct {
	shared.RootOptions
	Email string
	Role  string
}

func AddAddOrganizationMemberCmd(parent *cobra.Command, opts shared.RootOptions) {
	c := AddOrganizationMemberCmd{
		cmd: &cobra.Command{
			Use:     "organization-member",
			Aliases: []string{"org-member"},
			Short:   "Add organization member",
			Long:    "Add organization member. Select the organization using --org or CS_ORG_ID.",
			Example: io.FormatExampleCommands("add organization-member", []io.Example{
				{Cmd: "--org <orgId> -e user@example.com -r member", Desc: "Add organization member"},
			}),
		},
		Opts:          AddOrganizationMemberOpts{RootOptions: opts},
		ClientFactory: func(opts shared.RootOptions) (shared.Client, error) { return opts.NewClient() },
	}
	c.cmd.Flags().StringVarP(&c.Opts.Email, "email", "e", "", "Organization member email")
	_ = c.cmd.MarkFlagRequired("email")
	c.cmd.Flags().StringVarP(&c.Opts.Role, "role", "r", "member", "Organization role (admin, member)")

	c.cmd.RunE = c.RunE
	shared.AddCmd(parent, c.cmd)
}

func (c *AddOrganizationMemberCmd) RunE(_ *cobra.Command, _ []string) error {
	orgId, err := shared.RequireOrganizationID(c.Opts.RootOptions)
	if err != nil {
		return err
	}
	client, err := c.ClientFactory(c.Opts.RootOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}
	return c.AddOrganizationMember(client, orgId, c.Opts.Email, c.Opts.Role)
}

func (c *AddOrganizationMemberCmd) AddOrganizationMember(client shared.Client, orgId, email, role string) error {
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return fmt.Errorf("invalid email address")
	}
	if err := shared.ValidateOrganizationRole(role); err != nil {
		return err
	}

	if err := client.AddOrganizationMember(orgId, email, role); err != nil {
		return fmt.Errorf("failed to add organization member: %w", err)
	}
	return nil
}
