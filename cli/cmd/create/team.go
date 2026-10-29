// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"errors"
	"fmt"

	"github.com/codesphere-cloud/cs-go/api"
	shared "github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type CreateTeamCmd struct {
	cmd           *cobra.Command
	Opts          CreateTeamOpts
	ClientFactory func(shared.RootOptions) (shared.Client, error)
}

type CreateTeamOpts struct {
	shared.RootOptions
	Name string
	DcId int
}

func AddCreateTeamCmd(create *cobra.Command, opts shared.RootOptions) {
	t := CreateTeamCmd{
		cmd: &cobra.Command{
			Use:     "resource-group",
			Aliases: []string{"team"},
			Short:   "Create resource group",
			Long:    `Create a resource group in Codesphere or an Organization`,
			Example: io.FormatExampleCommands("create resource-group", []io.Example{
				{Cmd: "-d <datacenterId> -n <resourceGroupName>", Desc: "Create a resource group in a specific datacenter"},
				{Cmd: "-d <datacenterId> -n <resourceGroupName> -g <orgId>", Desc: "Create a resource group in a specific datacenter within an organization"},
			}),
		},
		Opts: CreateTeamOpts{
			RootOptions: opts,
		},
		ClientFactory: func(opts shared.RootOptions) (shared.Client, error) { return opts.NewClient() },
	}
	t.cmd.RunE = t.RunE
	t.cmd.Flags().StringVarP(&t.Opts.Name, "name", "n", "", "Resource group name")
	_ = t.cmd.MarkFlagRequired("name")
	t.cmd.Flags().IntVarP(&t.Opts.DcId, "dc-id", "d", 0, "Data center ID")
	shared.AddCmd(create, t.cmd)
}

func (c *CreateTeamCmd) RunE(_ *cobra.Command, args []string) error {
	client, err := c.ClientFactory(c.Opts.RootOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	orgId, err := c.Opts.GetOrgId()
	if err != nil {
		return errors.Join(err, errors.New("failed to get organization ID"))
	}

	createdTeam, err := c.CreateTeam(client, orgId, c.Opts.Name, c.Opts.DcId)
	if err != nil {
		return err
	}

	if orgId != "" {
		fmt.Printf("Team created: %+v in Organization: %+v\n", createdTeam.Id, orgId)
	} else {
		fmt.Printf("Team created: %+v\n", createdTeam.Id)
	}
	return nil
}

func (c *CreateTeamCmd) CreateTeam(client shared.Client, orgId string, teamName string, dcId int) (*api.Team, error) {
	if teamName == "" {
		return nil, errors.New("team name cannot be empty")
	}

	createdTeam, err := client.CreateTeam(orgId, teamName, dcId)
	if err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}
	return createdTeam, nil
}
