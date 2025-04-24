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
	ApiUrl      *string
	TeamId      *int
	WorkspaceId *int
	Env         Env
}

type Env interface {
	GetApiToken() (string, error)
	GetTeamId() (int, error)
	GetWorkspaceId() (int, error)
	GetApiUrl() string
}

func (o GlobalOptions) GetApiUrl() string {
	if o.ApiUrl != nil {
		return *o.ApiUrl
	}
	return o.Env.GetApiUrl()
}

func (o GlobalOptions) GetTeamId() (int, error) {
	if o.TeamId != nil {
		return *o.TeamId, nil
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
	if o.WorkspaceId != nil {
		return *o.WorkspaceId, nil
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

func Execute() {
	var rootCmd = &cobra.Command{
		Use:   "cs",
		Short: "The codesphere CLI",
		Long:  `Manage and debug resources deployed in Codesphere via command line.`,
	}

	opts := GlobalOptions{Env: cs.NewEnv()}

	addLogCmd(rootCmd, opts)
	addListCmd(rootCmd, opts)
	addSetEnvVarCmd(rootCmd, opts)

	opts.ApiUrl = rootCmd.PersistentFlags().StringP("api", "a", "https://codesphere.com/api", "URL of Codesphere API (can also be CS_API)")
	opts.TeamId = rootCmd.PersistentFlags().IntP("team", "t", -1, "Team ID (relevant for some commands, can also be CS_TEAM_ID)")
	opts.WorkspaceId = rootCmd.PersistentFlags().IntP("workspace", "w", -1, "Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID)")

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
