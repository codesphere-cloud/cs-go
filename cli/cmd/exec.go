// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/codesphere-cloud/cs-go/pkg/io"

	"github.com/spf13/cobra"
)

type ExecCmd struct {
	cmd  *cobra.Command
	Opts ExecOptions
}

type ExecOptions struct {
	GlobalOptions
	EnvVar  *[]string
	WorkDir *string
}

func (c *ExecCmd) RunE(_ *cobra.Command, args []string) error {
	command := strings.Join(args, " ")
	fmt.Printf("running command %s\n", command)

	client, err := NewClient(c.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	return c.ExecCommand(client, command)
}

func AddExecCmd(rootCmd *cobra.Command, opts GlobalOptions) {
	exec := ExecCmd{
		cmd: &cobra.Command{
			Use:   "exec",
			Args:  cobra.MinimumNArgs(1),
			Short: "Run a command in Codesphere workspace",
			Long: io.Long(`Run a command in a Codesphere workspace.
				Output will be printed to STDOUT, errors to STDERR.`),
			Example: io.FormatExampleCommands("exec", []io.Example{
				{Cmd: "-- echo hello world", Desc: "Print `hello world`"},
				{Cmd: "-- find .", Desc: "List all files in workspace"},
				{Cmd: "-d user -- find .", Desc: "List all files in the user directory"},
				{Cmd: "-e FOO=bar -- 'echo $FOO'", Desc: "Set custom environment variables for this command"},
			}),
		},
		Opts: ExecOptions{GlobalOptions: opts},
	}
	exec.Opts.EnvVar = exec.cmd.Flags().StringArrayP("env", "e", []string{}, "Additional environment variables to pass to the command in the form key=val")
	exec.Opts.WorkDir = exec.cmd.Flags().StringP("workdir", "d", "", "Working directory for the command")
	rootCmd.AddCommand(exec.cmd)
	exec.cmd.RunE = exec.RunE
}

func (c *ExecCmd) ExecCommand(client Client, command string) error {
	wsId, err := c.Opts.GetWorkspaceId()
	if err != nil {
		return fmt.Errorf("failed to get workspace ID: %w", err)
	}

	envVarMap, err := cs.ArgToEnvVarMap(*c.Opts.EnvVar)
	if err != nil {
		return fmt.Errorf("failed to parse environment variables: %w", err)
	}

	stdout, stderr, err := client.ExecCommand(wsId, command, *c.Opts.WorkDir, envVarMap)

	fmt.Println("STDOUT:")
	fmt.Println(stdout)
	if stderr != "" {
		fmt.Println("STDERR:")
		fmt.Fprintln(os.Stderr, stderr)
	}
	return err
}
