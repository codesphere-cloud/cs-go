/*
Copyright © 2025 Codesphere Inc. <support@codesphere.com>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/codesphere-cloud/cs-go/pkg/api"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/codesphere-cloud/cs-go/pkg/out"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/spf13/cobra"
)

type ListWorkspacesCmd struct {
	cmd  *cobra.Command
	opts ListWorkspacesOptions
}

type ListWorkspacesOptions struct {
	GlobalOptions
	TeamId *int
}

func addListWorkspacesCmd(p *cobra.Command, opts GlobalOptions) {
	l := ListWorkspacesCmd{
		cmd: &cobra.Command{
			Use:   "workspaces",
			Short: "list resources",
			Long:  `list resources available in Codesphere`,
			Example: `
List all workspaces:

$ cs list workspaces --team-id <team-id>
			`,
		},
		opts: ListWorkspacesOptions{GlobalOptions: opts},
	}
	l.cmd.RunE = l.RunE
	l.parseLogCmdFlags()
	p.AddCommand(l.cmd)
}

func (l *ListWorkspacesCmd) parseLogCmdFlags() {
	l.opts.TeamId = l.cmd.Flags().IntP("team-id", "t", -1, "ID of team to query")
}

func (l *ListWorkspacesCmd) RunE(_ *cobra.Command, args []string) (err error) {
	if l.opts.TeamId == nil || *l.opts.TeamId < 0 {
		return errors.New("team ID not set or invalid, please use --team-id to set one")
	}
	token, err := cs.GetApiToken()
	if err != nil {
		return fmt.Errorf("failed to get API token: %e", err)
	}
	client := api.NewClient(context.Background(), api.Configuration{
		BaseUrl: l.opts.GetApiUrl(),
		Token:   token,
	})
	workspaces, err := client.ListWorkspaces(*l.opts.TeamId)
	if err != nil {
		return fmt.Errorf("failed to list workspaces: %e", err)
	}

	t := out.GetTableWriter()
	t.AppendHeader(table.Row{"ID", "Name", "Repository"})
	for _, w := range workspaces {
		t.AppendRow(table.Row{w.Id, w.Name, *w.GitUrl.Get()})
	}
	t.Render()

	return nil
}
