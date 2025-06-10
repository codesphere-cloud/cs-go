package cmd

import (
	"fmt"
	"slices"
	"time"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/pkg/io"

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
	pipeline := StartPipelineCmd{
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
				{Cmd: "-p ci.prod.yml run", Desc: "Start the run stage of the ci.prod.yml profile"},
				{Cmd: "-t 5m prepare", Desc: "start the prepare stage, timeout after 5 minutes."},
			}),
		},
		Opts: StartPipelineOpts{GlobalOptions: opts},
		Time: &api.RealTime{},
	}

	pipeline.Opts.Timeout = pipeline.cmd.Flags().Duration("timeout", 30*time.Minute, "Time to wait per stage before stopping the command execution (e.g. 10m)")
	pipeline.Opts.Profile = pipeline.cmd.Flags().StringP("profile", "p", "", "CI profile to use (e.g. 'prod' for the profile defined in 'ci.prod.yml'), defaults to the ci.yml profile")
	start.AddCommand(pipeline.cmd)

	pipeline.cmd.RunE = pipeline.RunE
}

func (c *StartPipelineCmd) StartPipelineStages(client Client, wsId int, stages []string) error {
	for _, stage := range stages {
		if !IsValidStage(stage) {
			return fmt.Errorf("invalid pipeline stage: %s", stage)
		}
	}
	for _, stage := range stages {
		err := c.StartStage(client, wsId, stage)
		if err != nil {
			return err
		}
	}
	return nil
}

func IsValidStage(stage string) bool {
	return slices.Contains([]string{"prepare", "test", "run"}, stage)
}

func (c *StartPipelineCmd) StartStage(client Client, wsId int, stage string) error {
	fmt.Printf("starting %s stage on workspace %d...", stage, wsId)

	err := client.StartPipelineStage(wsId, *c.Opts.Profile, stage)
	if err != nil {
		fmt.Println()
		return fmt.Errorf("failed to start pipeline stage %s: %w", stage, err)
	}

	err = c.waitForPipelineStage(client, wsId, stage)
	if err != nil {
		return fmt.Errorf("failed waiting for stage %s to finish: %w", stage, err)

	}
	return nil
}

func (c *StartPipelineCmd) waitForPipelineStage(client Client, wsId int, stage string) error {
	delay := 5 * time.Second

	maxWaitTime := c.Time.Now().Add(*c.Opts.Timeout)
	for {
		status, err := client.GetPipelineState(wsId, stage)

		if err != nil {
			fmt.Printf("\nError getting pipeline status: %s, trying again...", err.Error())
			c.Time.Sleep(delay)
			continue
		}

		if allFinished(status) {
			fmt.Println("(finished)")
			break
		}

		if allRunning(status) && stage == "run" {
			fmt.Println("(running)")
			break
		}

		err = shouldAbort(status)
		if err != nil {
			fmt.Println()
			return fmt.Errorf("stage %s failed: %w", stage, err)
		}

		fmt.Print(".")
		if c.Time.Now().After(maxWaitTime) {
			fmt.Println()
			return fmt.Errorf("timed out waiting for pipeline stage %s to be complete", stage)
		}
		c.Time.Sleep(delay)
	}
	return nil
}

func allRunning(status []api.PipelineStatus) bool {
	for _, s := range status {
		if s.State != "running" {
			return false
		}
	}
	return true
}

func allFinished(status []api.PipelineStatus) bool {
	for _, s := range status {
		if s.State != "success" {
			return false
		}
	}
	return true
}

func shouldAbort(status []api.PipelineStatus) error {
	for _, s := range status {
		if slices.Contains([]string{"failure", "aborted"}, s.State) {
			return fmt.Errorf("server %s, replica %s reached unexpected state %s", s.Server, s.Replica, s.State)
		}
	}
	return nil
}
