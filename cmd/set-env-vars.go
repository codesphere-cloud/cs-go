// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type SetEnvVarCmd struct {
	opts SetEnvVarOptions
	cmd  *cobra.Command
}

type SetEnvVarOptions struct {
	GlobalOptions
	WorkspaceId *int
	EnvVar      *[]string
}

func addSetEnvVarCmd(p *cobra.Command, opts GlobalOptions) {
	l := SetEnvVarCmd{
		cmd: &cobra.Command{
			Use:   "set-env",
			Short: "set env vars",
			Long:  `set an environment variable for your workspace`,
			Example: `
Set environment variable:

$ cs set env var --workspace-id <workspace-id> --name <env-var-name> --value <env-var-value>
			`,
		},
		opts: SetEnvVarOptions{GlobalOptions: opts},
	}
	l.cmd.RunE = l.RunE
	l.parseFlags()
	p.AddCommand(l.cmd)
}

func (l *SetEnvVarCmd) parseFlags() {
	l.opts.WorkspaceId = l.cmd.Flags().IntP("workspace-id", "w", -1, "ID of workspace to set var")
	l.opts.EnvVar = l.cmd.Flags().StringArrayP("env-var", "e", []string{}, "env vars to set in form key=val")
}

func (l *SetEnvVarCmd) RunE(_ *cobra.Command, args []string) (err error) {
	client, err := NewClient(l.opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}
	envVarMap := map[string]string{}
	for _, x := range *l.opts.EnvVar {
		split := strings.Split(x, "=")
		envVarMap[split[0]] = split[1]
	}
	err = client.SetEnvVarOnWorkspace(*l.opts.WorkspaceId, envVarMap)
	if err != nil {
		return fmt.Errorf("failed to set environment variables %v: %w", envVarMap, err)
	}

	return nil
}
