// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

type SetEnvVarCmd struct {
	Opts SetEnvVarOptions
	cmd  *cobra.Command
}

type SetEnvVarOptions struct {
	GlobalOptions
	EnvVar  *[]string
	EnvFile *string
}

func AddSetEnvVarCmd(p *cobra.Command, opts GlobalOptions) {
	l := SetEnvVarCmd{
		cmd: &cobra.Command{
			Use:   "set-env",
			Short: "Set environment varariables",
			Long:  `Set environment variables in a workspace from flags or a .env file.`,
			// BEISPIELE AKTUALISIERT
			Example: io.FormatExampleCommands("set-env", []io.Example{
				{Cmd: "--workspace <id> --env-var FOO=bar", Desc: "Set a single environment variable"},
				{Cmd: "--workspace <id> --env-var FOO=bar --env-var HELLO=world", Desc: "Set multiple environment variables"},
				{Cmd: "--workspace <id> --env-file ./.env", Desc: "Set environment variables from a .env file"},
				{Cmd: "--workspace <id> --env-file ./.env --env-var FOO=new_value", Desc: "Set from a file and override/add a specific variable"},
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
	l.Opts.EnvFile = l.cmd.Flags().StringP("env-file", "f", "", "path to a .env file")
}

func (l *SetEnvVarCmd) RunE(_ *cobra.Command, args []string) (err error) {
	client, err := NewClient(l.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	return l.SetEnvironmentVariables(client)
}

func (l *SetEnvVarCmd) SetEnvironmentVariables(client Client) (err error) {
	finalEnvVarMap := make(map[string]string)

	if l.Opts.EnvFile != nil && *l.Opts.EnvFile != "" {
		envFile := *l.Opts.EnvFile
		if _, err := os.Stat(envFile); os.IsNotExist(err) {
			return fmt.Errorf("env file does not exist: %s", envFile)
		}

		fileEnvMap, err := godotenv.Read(envFile)
		if err != nil {
			return fmt.Errorf("failed to parse env file %s: %w", envFile, err)
		}

		for key, value := range fileEnvMap {
			finalEnvVarMap[key] = value
		}
	}

	if l.Opts.EnvVar != nil && len(*l.Opts.EnvVar) > 0 {
		flagEnvVarMap, err := cs.ArgToEnvVarMap(*l.Opts.EnvVar)
		if err != nil {
			return fmt.Errorf("failed to parse environment variables from flags: %w", err)
		}

		for key, value := range flagEnvVarMap {
			finalEnvVarMap[key] = value
		}
	}

	if len(finalEnvVarMap) == 0 {
		fmt.Println("No environment variables provided to set.")
		return nil
	}

	wsId, err := l.Opts.GetWorkspaceId()
	if err != nil {
		return fmt.Errorf("failed to get workspace ID: %w", err)
	}

	err = client.SetEnvVarOnWorkspace(wsId, finalEnvVarMap)
	if err != nil {
		return fmt.Errorf("failed to set environment variables %v: %w", finalEnvVarMap, err)
	}

	fmt.Printf("Successfully set %d environment variable(s) on workspace %s\n", len(finalEnvVarMap), strconv.Itoa(wsId))
	return nil
}
