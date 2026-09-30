// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package delete

import (
	"fmt"

	"github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type DeleteOrganizationMemberCmd struct {
	cmd           *cobra.Command
	Opts          DeleteOrganizationMemberOpts
	ClientFactory func(shared.RootOptions) (shared.Client, error)
}

type DeleteOrganizationMemberOpts struct {
	shared.RootOptions
	UserId int
}

func AddDeleteOrganizationMemberCmd(parent *cobra.Command, opts shared.RootOptions) {
	c := DeleteOrganizationMemberCmd{
		cmd: &cobra.Command{
			Use:     "organization-member",
			Aliases: []string{"org-member"},
			Short:   "Remove organization member",
			Long:    "Remove organization member. Select the organization using --org or CS_ORG_ID.",
			Example: io.FormatExampleCommands("delete organization-member", []io.Example{
				{Cmd: "--org <orgId> -u <userId>", Desc: "Remove organization member"},
			}),
		},
		Opts:          DeleteOrganizationMemberOpts{RootOptions: opts},
		ClientFactory: func(opts shared.RootOptions) (shared.Client, error) { return opts.NewClient() },
	}
	c.cmd.Flags().IntVarP(&c.Opts.UserId, "user", "u", -1, "Organization member user ID")
	_ = c.cmd.MarkFlagRequired("user")

	c.cmd.RunE = c.RunE
	shared.AddCmd(parent, c.cmd)
}

func (c *DeleteOrganizationMemberCmd) RunE(_ *cobra.Command, _ []string) error {
	orgId, err := shared.RequireOrganizationID(c.Opts.RootOptions)
	if err != nil {
		return err
	}
	client, err := c.ClientFactory(c.Opts.RootOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}
	return c.DeleteOrganizationMember(client, orgId, c.Opts.UserId)
}

func (c *DeleteOrganizationMemberCmd) DeleteOrganizationMember(client shared.Client, orgId string, userId int) error {
	if err := shared.ValidateUserID(userId); err != nil {
		return err
	}

	if err := client.RemoveOrganizationMember(orgId, userId); err != nil {
		return fmt.Errorf("failed to remove organization member: %w", err)
	}
	return nil
}
