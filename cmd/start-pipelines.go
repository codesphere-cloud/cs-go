// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

type PipelinesStartCmd struct {
	Opts PipelinesStartOptions
	cmd  *cobra.Command
}

type PipelinesStartOptions struct {
	GlobalOptions
}

func addPipelinesStartCmd(p *cobra.Command, opts GlobalOptions) {
	s := PipelinesStartCmd{
		cmd: &cobra.Command{
			Use:   "start",
			Short: "start",
			Long:  `start pipelines available in Codesphere`,
			Example: `
Start pipeline stage:

$ cs pipelines start <stage>
			`,
		},
		Opts: PipelinesStartOptions{GlobalOptions: opts},
	}
	s.cmd.RunE = s.RunE
	p.AddCommand(s.cmd)
}

func (s *PipelinesStartCmd) RunE(_ *cobra.Command, args []string) (err error) {
	client, err := NewClient(s.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	workspaceId, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("failed to parse workspaceId: %w", err)
	}

	pipelineStage, err := s.parsePipelineStage(args[1])

	return s.StartPipeline(client, workspaceId, pipelineStage)
}

func (s *PipelinesStartCmd) parsePipelineStage(stage string) (string, error) {
	switch stage {
	case "prepare", "test", "run":
		return stage, nil
	default:
		return "", fmt.Errorf("pipeline stage %s is not valid. Must be one of: prepare, test, run", stage)
	}
}

func (s *PipelinesStartCmd) StartPipeline(client Client, workspaceId int, pipelineStage string) error {
	return client.StartPipelines(workspaceId, pipelineStage)
}
