// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"errors"
	"fmt"
	"net/mail"

	"github.com/codesphere-cloud/cs-go/api"
	shared "github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type CreateOrganizationCmd struct {
	cmd  *cobra.Command
	Opts CreateOrganizationOpts
}

type CreateOrganizationOpts struct {
	shared.RootOptions
	Name       string
	AdminEmail string
}

func AddCreateOrganizationCmd(parent *cobra.Command, opts shared.RootOptions) {
	c := CreateOrganizationCmd{
		cmd: &cobra.Command{
			Use:     "organization",
			Aliases: []string{"organization", "org", "orgs"},
			Short:   "Create organization",
			Long:    `Create an organization in Codesphere`,
			Example: io.FormatExampleCommands("create organization", []io.Example{
				{Cmd: "-n <name> -e <adminEmail>", Desc: "Create an organization with a specific name and admin email"},
			}),
		},
		Opts: CreateOrganizationOpts{
			RootOptions: opts,
		},
	}
	c.cmd.RunE = c.RunE
	c.cmd.Flags().StringVarP(&c.Opts.Name, "name", "n", "", "Organization name")
	_ = c.cmd.MarkFlagRequired("name")
	c.cmd.Flags().StringVarP(&c.Opts.AdminEmail, "admin-email", "e", "", "Organization admin email")
	_ = c.cmd.MarkFlagRequired("admin-email")
	shared.AddCmd(parent, c.cmd)
}

func (c *CreateOrganizationCmd) RunE(_ *cobra.Command, args []string) error {
	client, err := c.Opts.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	createdOrg, err := c.CreateOrganization(client, c.Opts.Name, c.Opts.AdminEmail)
	if err != nil {
		return err
	}

	fmt.Printf("Organization created: %+v\n", createdOrg.Id)
	return nil
}

func (c *CreateOrganizationCmd) CreateOrganization(client Client, name string, adminEmail string) (*api.Organization, error) {
	if name == "" {
		return nil, errors.New("organization name cannot be empty")
	}
	if adminEmail == "" {
		return nil, errors.New("admin email cannot be empty")
	}
	if _, err := mail.ParseAddress(adminEmail); err != nil {
		return nil, errors.New("admin email is invalid")
	}

	createdOrg, err := client.CreateOrganization(name, adminEmail)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}
	return createdOrg, nil
}
