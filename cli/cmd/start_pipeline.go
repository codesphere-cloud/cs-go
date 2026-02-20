// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"time"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/codesphere-cloud/cs-go/pkg/pipeline"

	"github.com/spf13/cobra"
)

type StartPipelineCmd struct {
	cmd  *cobra.Command
	Opts StartPipelineOpts
	Time api.Time
}

type StartPipelineOpts struct {
	GlobalOptions
	Profile *string
	Timeout *time.Duration
}

func (c *StartPipelineCmd) RunE(_ *cobra.Command, args []string) error {

	workspaceId, err := c.Opts.GetWorkspaceId()
	if err != nil {
		return fmt.Errorf("failed to get workspace ID: %w", err)
	}

	client, err := NewClient(c.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	return c.StartPipelineStages(client, workspaceId, args)
}

func AddStartPipelineCmd(start *cobra.Command, opts GlobalOptions) {
	p := StartPipelineCmd{
		cmd: &cobra.Command{
			Use:   "pipeline",
			Short: "Start pipeline stages of a workspace",
			Args:  cobra.RangeArgs(1, 3),
			Long: io.Long(`Start one or many pipeline stages of a workspace.

				Stages can be 'prepare', 'test', or 'run'.
				When multiple stages are specified, the command will start the next stage when the previous stage is finished successfully.
				If a stage fails, the command won't attempt running the next stage.
				The command will not wait for the run stage to finish, but exit when the stage is running.

				When only a single stage is specified, the command will wait until the stage is finished, except for the run stage.
				Use '` + io.BinName() + ` log' to stream logs.`),
			Example: io.FormatExampleCommands("start pipeline", []io.Example{
				{Cmd: "prepare", Desc: "Start the prepare stage and wait for it to finish"},
				{Cmd: "prepare test", Desc: "Start the prepare and test stages sequencially and wait for them to finish"},
				{Cmd: "prepare test run", Desc: "Start the prepare, test, and run stages sequencially. Exits after the run stage is triggered"},
				{Cmd: "run", Desc: "Start the run stage and exit when running"},
				{Cmd: "-p prod run", Desc: "Start the run stage of the prod profile"},
				{Cmd: "-t 5m prepare", Desc: "start the prepare stage, timeout after 5 minutes."},
			}),
		},
		Opts: StartPipelineOpts{GlobalOptions: opts},
		Time: &api.RealTime{},
	}

	p.Opts.Timeout = p.cmd.Flags().Duration("timeout", 30*time.Minute, "Time to wait per stage before stopping the command execution (e.g. 10m)")
	p.Opts.Profile = p.cmd.Flags().StringP("profile", "p", "", "CI profile to use (e.g. 'prod' for the profile defined in 'ci.prod.yml'), defaults to the ci.yml profile")
	start.AddCommand(p.cmd)

	p.cmd.RunE = p.RunE
}

func (c *StartPipelineCmd) StartPipelineStages(client Client, wsId int, stages []string) error {
	runner := pipeline.NewRunner(client, c.Time)
	return runner.RunStages(wsId, stages, pipeline.Config{
		Profile: *c.Opts.Profile,
		Timeout: *c.Opts.Timeout,
	})
}
