// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"log"

	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type StopPipelineCmd struct {
	cmd  *cobra.Command
	Opts StopPipelineOpts
}

type StopPipelineOpts struct {
	*GlobalOptions
}

func (c *StopPipelineCmd) RunE(_ *cobra.Command, args []string) error {
	workspaceId, err := c.Opts.GetWorkspaceId()
	if err != nil {
		return fmt.Errorf("failed to get workspace ID: %w", err)
	}

	client, err := NewClient(*c.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	return c.StopPipelineStages(client, workspaceId, args)
}

func AddStopPipelineCmd(stop *cobra.Command, opts *GlobalOptions) {
	pipeline := StopPipelineCmd{
		cmd: &cobra.Command{
			Use:   "pipeline",
			Short: "Stop pipeline stages of a workspace",
			Args:  cobra.RangeArgs(1, 3),
			Long: io.Long(`Stop one or many pipeline stages of a workspace.

				Stages can be 'prepare', 'test', or 'run'.
				When multiple stages are specified, the command will stop them in the provided order.
				The command sends a stop request for each stage and returns after all requests succeed.`),
			Example: io.FormatExampleCommands("stop pipeline", []io.Example{
				{Cmd: "run", Desc: "Stop the run stage"},
				{Cmd: "prepare test", Desc: "Stop the prepare and test stages in order"},
				{Cmd: "prepare test run", Desc: "Stop the prepare, test, and run stages in order"},
			}),
		},
		Opts: StopPipelineOpts{GlobalOptions: opts},
	}
	AddCmd(stop, pipeline.cmd)
	pipeline.cmd.RunE = pipeline.RunE
}

func (c *StopPipelineCmd) StopPipelineStages(client Client, wsId int, stages []string) error {
	for _, stage := range stages {
		if !isValidStage(stage) {
			return fmt.Errorf("invalid pipeline stage: %s", stage)
		}
	}
	for _, stage := range stages {
		err := c.stopStage(client, wsId, stage)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *StopPipelineCmd) stopStage(client Client, wsId int, stage string) error {
	log.Printf("stopping %s stage on workspace %d...", stage, wsId)

	err := client.StopPipelineStage(wsId, stage)
	if err != nil {
		log.Println()
		return fmt.Errorf("failed to stop pipeline stage %s: %w", stage, err)
	}

	log.Printf("stage %s stop requested\n", stage)
	return nil
}
