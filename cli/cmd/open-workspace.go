package cmd

import (
	"fmt"

	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/codesphere-cloud/cs-go/pkg/out"
	"github.com/spf13/cobra"
)

// OpenWorkspaceCmd represents the workspace command
type OpenWorkspaceCmd struct {
	cmd  *cobra.Command
	Opts GlobalOptions
}

type Browser interface {
	OpenIde(path string) error
}

func (c *OpenWorkspaceCmd) RunE(_ *cobra.Command, args []string) error {
	client, err := NewClient(c.Opts)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	wsId, err := c.Opts.GetWorkspaceId()
	if err != nil {
		return fmt.Errorf("failed to get workspace ID: %w", err)
	}

	return c.OpenWorkspace(cs.NewBrowser(), client, wsId)
}

func AddOpenWorkspaceCmd(open *cobra.Command, opts GlobalOptions) {
	workspace := OpenWorkspaceCmd{
		cmd: &cobra.Command{
			Use:   "workspace",
			Short: "Open workspace in web browser",
			Long:  `Open workspace in the Codesphere IDE.`,
			Example: out.FormatExampleCommands("open workspace", []out.Example{
				{Cmd: "-w 42", Desc: "open workspace 42 in web browser"},
				{Cmd: "", Desc: "open workspace set by environment variable CS_WORKSPACE_ID"},
			}),
		},
		Opts: opts,
	}
	open.AddCommand(workspace.cmd)
	workspace.cmd.RunE = workspace.RunE
}

func (cmd *OpenWorkspaceCmd) OpenWorkspace(browser Browser, client Client, wsId int) error {
	workspace, err := client.GetWorkspace(wsId)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	fmt.Printf("Opening workspace %d in Codesphere IDE\n", wsId)

	err = browser.OpenIde(fmt.Sprintf("teams/%d/workspaces/%d", workspace.TeamId, wsId))
	if err != nil {
		return fmt.Errorf("failed to open web browser: %w", err)
	}

	return nil
}
