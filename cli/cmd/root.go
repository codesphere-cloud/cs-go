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
	installcmd "github.com/codesphere-cloud/cs-go/cli/cmd/install"
	listcmd "github.com/codesphere-cloud/cs-go/cli/cmd/list"
	startcmd "github.com/codesphere-cloud/cs-go/cli/cmd/start"
	updatecmd "github.com/codesphere-cloud/cs-go/cli/cmd/update"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	rootCmd.PersistentFlags().IntVarP(&opts.TeamId, "team", "t", -1, "Resource group (team) ID (relevant for some commands, alias --resource-group, can also be CS_RESOURCE_GROUP_ID or CS_TEAM_ID)")
	rootCmd.PersistentFlags().IntVarP(&opts.WorkspaceId, "workspace", "w", -1, "Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID)")
	rootCmd.PersistentFlags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().StringVarP(&opts.OrgId, "org", "g", "", "Organization ID (relevant for some commands)")
	// Keep the previous shorthand bound to the same value for existing scripts.
	rootCmd.PersistentFlags().StringVarP(&opts.OrgId, "org-legacy", "O", "", "Alias for --org")
	_ = rootCmd.PersistentFlags().MarkHidden("org-legacy")
	rootCmd.SetGlobalNormalizationFunc(normalizeFlagAliases)

	listcmd.AddListCmd(rootCmd, &opts)
	generatecmd.AddGenerateCmd(rootCmd, &opts)
	createcmd.AddCreateCmd(rootCmd, &opts)
	deletecmd.AddDeleteCmd(rootCmd, &opts)
	addcmd.AddAddCmd(rootCmd, &opts)
	startcmd.AddStartCmd(rootCmd, &opts)
	installcmd.AddInstallCmd(rootCmd)
	AddExecCmd(rootCmd, &opts)
	AddVersionCmd(rootCmd)
	AddLicensesCmd(rootCmd)
	AddOpenCmd(rootCmd, &opts)
	AddMonitorCmd(rootCmd, &opts)
	AddStopCmd(rootCmd, &opts)
	AddGitCmd(rootCmd, &opts)
	AddSyncCmd(rootCmd, &opts)
	updatecmd.AddUpdateCmd(rootCmd, &opts)
	AddGoCmd(rootCmd)
	AddWakeUpCmd(rootCmd, &opts)
	AddCurlCmd(rootCmd, &opts)
	AddScaleCmd(rootCmd, &opts)
	AddMcpCmd(rootCmd)

	AddLegacyCmds(rootCmd, &opts)

	return rootCmd
}

// normalizeFlagAliases makes --resource-group resolve to the --team flag, so both names share one value.
func normalizeFlagAliases(_ *pflag.FlagSet, name string) pflag.NormalizedName {
	if name == "resource-group" {
		return "team"
	}
	return pflag.NormalizedName(name)
}

func Execute() {
	err := GetRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
