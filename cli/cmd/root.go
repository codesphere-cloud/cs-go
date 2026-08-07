// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"fmt"
	"os"

	addcmd "github.com/codesphere-cloud/cs-go/cli/cmd/add"
	createcmd "github.com/codesphere-cloud/cs-go/cli/cmd/create"
	deletecmd "github.com/codesphere-cloud/cs-go/cli/cmd/delete"
	generatecmd "github.com/codesphere-cloud/cs-go/cli/cmd/generate"
	listcmd "github.com/codesphere-cloud/cs-go/cli/cmd/list"
	startcmd "github.com/codesphere-cloud/cs-go/cli/cmd/start"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

type GlobalOptions struct {
	ApiUrl      string
	TeamId      int
	WorkspaceId int
	OrgId       string
	Env         Env
	Verbose     bool
}

func (o GlobalOptions) GetApiToken() (string, error) {
	return o.Env.GetApiToken()
}

func (o GlobalOptions) GetVerbose() bool {
	return o.Verbose
}

type Env interface {
	GetApiToken() (string, error)
	GetTeamId() (int, error)
	GetWorkspaceId() (int, error)
	GetOrgId() string
	GetApiUrl() string
}

func (o GlobalOptions) GetApiUrl() string {
	if o.ApiUrl != "" {
		return o.ApiUrl
	}
	return o.Env.GetApiUrl()
}

func (o GlobalOptions) GetTeamId() (int, error) {
	if o.TeamId != -1 {
		return o.TeamId, nil
	}
	wsId, err := o.Env.GetTeamId()
	if err != nil {
		return -1, err
	}
	if wsId < 0 {
		return -1, errors.New("team ID not set, use -t or CS_TEAM_ID to set it")
	}
	return wsId, nil
}

func (o GlobalOptions) GetWorkspaceId() (int, error) {
	if o.WorkspaceId != -1 {
		return o.WorkspaceId, nil
	}
	wsId, err := o.Env.GetWorkspaceId()
	if err != nil {
		return -1, err
	}
	if wsId < 0 {
		return -1, errors.New("workspace ID not set, use -w or CS_WORKSPACE_ID to set it")
	}
	return wsId, nil
}

func (o GlobalOptions) GetOrgId() (string, error) {
	orgId := o.OrgId
	if orgId == "" {
		orgId = o.Env.GetOrgId()
	}

	if orgId == "" {
		return "", nil
	}

	_, err := uuid.Parse(orgId)
	if err != nil {
		return "", fmt.Errorf("invalid organization ID format: %w", err)
	}

	return orgId, nil
}

// AddCmd adds a command, inheriting the parent's Args validator if not explicitly set.
// Individual commands that need different argument rules can override this by setting their own Args validator.
func AddCmd(parent *cobra.Command, cmd *cobra.Command) {
	if cmd.Args == nil {
		cmd.Args = parent.Args
	}
	parent.AddCommand(cmd)
}

func GetRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:               "cs",
		Short:             "The Codesphere CLI",
		Long:              `Manage and debug resources deployed in Codesphere via command line.`,
		Args:              cobra.NoArgs,
		DisableAutoGenTag: true,
	}

	opts := GlobalOptions{Env: cs.NewEnv()}

	rootCmd.PersistentFlags().StringVarP(&opts.ApiUrl, "api", "a", "", "URL of Codesphere API (can also be CS_API)")
	rootCmd.PersistentFlags().IntVarP(&opts.TeamId, "team", "t", -1, "Team ID (relevant for some commands, can also be CS_TEAM_ID)")
	rootCmd.PersistentFlags().IntVarP(&opts.WorkspaceId, "workspace", "w", -1, "Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID)")
	rootCmd.PersistentFlags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().StringVarP(&opts.OrgId, "org", "O", "", "Organization ID (relevant for some commands)")

	AddExecCmd(rootCmd, &opts)
	listcmd.AddListCmd(rootCmd, &opts)
	AddVersionCmd(rootCmd)
	AddLicensesCmd(rootCmd)
	AddOpenCmd(rootCmd, &opts)
	generatecmd.AddGenerateCmd(rootCmd, &opts)
	createcmd.AddCreateCmd(rootCmd, &opts)
	deletecmd.AddDeleteCmd(rootCmd, &opts)
	addcmd.AddAddCmd(rootCmd, &opts)
	AddMonitorCmd(rootCmd, &opts)
	startcmd.AddStartCmd(rootCmd, &opts)
	AddGitCmd(rootCmd, &opts)
	AddSyncCmd(rootCmd, &opts)
	AddUpdateCmd(rootCmd)
	AddGoCmd(rootCmd)
	AddWakeUpCmd(rootCmd, &opts)
	AddCurlCmd(rootCmd, &opts)
	AddScaleCmd(rootCmd, &opts)
	AddMcpCmd(rootCmd)

	return rootCmd
}

func Execute() {
	err := GetRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
