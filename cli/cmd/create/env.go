// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"fmt"
	"log"

	shared "github.com/codesphere-cloud/cs-go/cli/cmd/shared"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type CreateEnvCmd struct {
	Opts CreateEnvOptions
	cmd  *cobra.Command
}

type CreateEnvOptions struct {
	shared.RootOptions
	EnvVar *[]string
}

func AddCreateEnvCmd(p *cobra.Command, opts shared.RootOptions) {
	l := CreateEnvCmd{
		cmd: &cobra.Command{
			Use:   "env",
			Short: "Set environment variables",
			Long:  `Set environment variables in a workspace`,
			Example: io.FormatExampleCommands("create env", []io.Example{
				{Cmd: "--workspace <workspace-id> --env-var foo=bar", Desc: "Set single environment variable"},
				{Cmd: "--workspace <workspace-id> --env-var foo=bar --env-var hello=world", Desc: "Set multiple environment variables"},
			}),
		},
		Opts: CreateEnvOptions{RootOptions: opts},
	}
	l.cmd.RunE = l.RunE
	l.parseFlags()
	shared.AddCmd(p, l.cmd)
}

func (l *CreateEnvCmd) parseFlags() {
	l.Opts.EnvVar = l.cmd.Flags().StringArrayP("env-var", "e", []string{}, "env vars to set in form key=val")
}

func (l *CreateEnvCmd) RunE(_ *cobra.Command, args []string) (err error) {
	client, err := l.Opts.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	return l.SetEnvironmentVariables(client)
}

func (l *CreateEnvCmd) SetEnvironmentVariables(client Client) (err error) {
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

	log.Printf("Environment variables set successfully on workspace %d\n", wsId)
	return nil
}
