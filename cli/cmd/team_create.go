package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
)

type CreateTeamCmd struct {
	cmd  *cobra.Command
	Opts CreateTeamOpts
}

type CreateTeamOpts struct {
	*GlobalOptions
	name  string
	dcId  int
	orgId string
}

func AddCreateTeamCmd(team *cobra.Command, opts *GlobalOptions) {
	t := CreateTeamCmd{
		cmd: &cobra.Command{
			Use:   "create",
			Short: "Create team",
			Long:  `Create a team in Codesphere or an Organization`,
		},
		Opts: CreateTeamOpts{
			GlobalOptions: opts,
		},
	}
	t.cmd.RunE = t.RunE
	t.cmd.Flags().StringVarP(&t.Opts.name, "name", "n", "", "Team name")
	t.cmd.Flags().IntVarP(&t.Opts.dcId, "dc-id", "d", 0, "Data center ID")
	AddCmd(team, t.cmd)
}

func (c *CreateTeamCmd) RunE(_ *cobra.Command, args []string) error {
	client, err := NewClient(*c.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	orgId, err := c.Opts.GetOrgId()
	if err != nil {
		return errors.New("organization ID not set, use -O or CS_ORG_ID to set it")
	}

	teamName := c.Opts.name
	dcId := c.Opts.dcId

	createdTeam, err := client.CreateTeam(orgId, teamName, dcId)
	if err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}

	fmt.Printf("Team created: %+v in Organization: %+v\n", createdTeam.Id, orgId)
	return nil
}
