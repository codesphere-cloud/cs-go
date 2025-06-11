// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type SetEnvVarCmd struct {
	Opts SetEnvVarOptions
	cmd  *cobra.Command
}

type SetEnvVarOptions struct {
	GlobalOptions
	EnvVar *[]string
}

func AddSetEnvVarCmd(p *cobra.Command, opts GlobalOptions) {
	l := SetEnvVarCmd{
		cmd: &cobra.Command{
			Use:   "set-env",
			Short: "Set environment varariables",
			Long:  `Set environment variables in a workspace`,
			Example: io.FormatExampleCommands("set-env", []io.Example{
				{Cmd: "--workspace <workspace-id> --env-var foo=bar", Desc: "Set single environment variable"},
				{Cmd: "--workspace <workspace-id> --env-var foo=bar --env-var hello=world", Desc: "Set multiple environment variables"},
			}),
		},
		Opts: SetEnvVarOptions{GlobalOptions: opts},
	}
	l.cmd.RunE = l.RunE
	l.parseFlags()
	p.AddCommand(l.cmd)
}

func (l *SetEnvVarCmd) parseFlags() {
	l.Opts.EnvVar = l.cmd.Flags().StringArrayP("env-var", "e", []string{}, "env vars to set in form key=val")
}

func (l *SetEnvVarCmd) RunE(_ *cobra.Command, args []string) (err error) {
	client, err := NewClient(l.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	return l.SetEnvironmentVariables(client)
}

func (l *SetEnvVarCmd) SetEnvironmentVariables(client Client) (err error) {
	envVarMap, err := cs.ArgToEnvVarMap(*l.Opts.EnvVar)
	if err != nil {
		return fmt.Errorf("failed to parse environment variables: %w", err)
	}
	wsId, err := l.Opts.GetWorkspaceId()
	if err != nil {
		return fmt.Errorf("failed to get workspace ID: %w", err)
	}

	err = client.SetEnvVarOnWorkspace(wsId, envVarMap)
	if err != nil {
		return fmt.Errorf("failed to set environment variables %v: %w", envVarMap, err)
	}

	return nil
}
