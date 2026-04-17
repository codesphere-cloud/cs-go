// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"os"

	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/spf13/cobra"
)

type GlobalOptions struct {
	ApiUrl      string
	TeamId      int
	WorkspaceId int
	env         cs.Env
	Verbose     bool
	StateFile   string
}

// NewGlobalOptionsWithCustomEnv creates a new GlobalOptions with a custom environment, useful for testing
func NewGlobalOptionsWithCustomEnv(opts GlobalOptions, env cs.Env) *GlobalOptions {
	opts.env = env
	return &opts
}

func (o *GlobalOptions) Env() cs.Env {
	if o.env == nil {
		o.env = cs.NewEnv(o.StateFile)
	}
	return o.env
}

func (o GlobalOptions) GetApiUrl() string {
	if o.ApiUrl != "" {
		return o.ApiUrl
	}
	return o.Env().GetApiUrl()
}

func (o GlobalOptions) GetTeamId() (int, error) {
	if o.TeamId != -1 {
		return o.TeamId, nil
	}
	teamId, err := o.Env().GetTeamId()
	if err != nil {
		return -1, err
	}
	if teamId <= 0 {
		return -1, errors.New("team ID not set, use -t or CS_TEAM_ID to set it")
	}
	return teamId, nil
}

func (o GlobalOptions) GetWorkspaceId() (int, error) {
	if o.WorkspaceId != -1 {
		return o.WorkspaceId, nil
	}
	wsId, err := o.Env().GetWorkspaceId()
	if err != nil {
		return -1, err
	}
	if wsId <= 0 {
		return -1, errors.New("workspace ID not set, use -w or CS_WORKSPACE_ID to set it")
	}
	return wsId, nil
}

func GetRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:               "cs",
		Short:             "The Codesphere CLI",
		Long:              `Manage and debug resources deployed in Codesphere via command line.`,
		DisableAutoGenTag: true,
	}

	opts := GlobalOptions{}

	rootCmd.PersistentFlags().StringVarP(&opts.ApiUrl, "api", "a", "", "URL of Codesphere API (can also be CS_API)")
	rootCmd.PersistentFlags().IntVarP(&opts.TeamId, "team", "t", -1, "Team ID (relevant for some commands, can also be CS_TEAM_ID)")
	rootCmd.PersistentFlags().IntVarP(&opts.WorkspaceId, "workspace", "w", -1, "Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID)")
	rootCmd.PersistentFlags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().StringVarP(&opts.StateFile, "state-file", "", ".cs-up.yaml", "Path to the state file, defaults to .cs-up.yaml")

	AddExecCmd(rootCmd, &opts)
	AddLogCmd(rootCmd, &opts)
	AddListCmd(rootCmd, &opts)
	AddSetEnvVarCmd(rootCmd, &opts)
	AddVersionCmd(rootCmd)
	AddLicensesCmd(rootCmd)
	AddOpenCmd(rootCmd, &opts)
	AddGenerateCmd(rootCmd, &opts)
	AddCreateCmd(rootCmd, &opts)
	AddDeleteCmd(rootCmd, &opts)
	AddMonitorCmd(rootCmd, &opts)
	AddStartCmd(rootCmd, &opts)
	AddGitCmd(rootCmd, &opts)
	AddSyncCmd(rootCmd, &opts)
	AddUpdateCmd(rootCmd)
	AddGoCmd(rootCmd)
	AddWakeUpCmd(rootCmd, &opts)
	AddCurlCmd(rootCmd, &opts)
	AddScaleCmd(rootCmd, &opts)
	AddUpCmd(rootCmd, &opts)
	AddPsCmd(rootCmd, &opts)

	return rootCmd
}

func Execute() {
	err := GetRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
