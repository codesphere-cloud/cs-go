package pipeline

import (
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/pkg/io"
)

const IdeServer string = "codesphere-ide"

type PipelineRunner struct {
	Client        api.Client
	Profile       string
	Time          api.Time
	Timeout       time.Duration
	VerboseOutput bool
}

func NewPipelineRunner(client api.Client, profile string, timeout time.Duration, verboseOutput bool) *PipelineRunner {
	return &PipelineRunner{
		Client:        client,
		Profile:       profile,
		Time:          &api.RealTime{},
		Timeout:       timeout,
		VerboseOutput: verboseOutput,
	}
}

func NewPipelineRunnerWidthCustomDeps(client api.Client, profile string, time api.Time, timeout time.Duration, verboseOutput bool) *PipelineRunner {
	return &PipelineRunner{
		Client:        client,
		Profile:       profile,
		Time:          time,
		Timeout:       timeout,
		VerboseOutput: verboseOutput,
	}
}

func (pr *PipelineRunner) StartPipelineStages(wsId int, stages []string) error {
	for _, stage := range stages {
		if !isValidStage(stage) {
			return fmt.Errorf("invalid pipeline stage: %s", stage)
		}
	}
	for _, stage := range stages {
		err := pr.StartStage(wsId, pr.Profile, stage, pr.Time, pr.Timeout, pr.VerboseOutput)
		if err != nil {
			return err
		}
	}
	return nil
}

func isValidStage(stage string) bool {
	return slices.Contains([]string{"prepare", "test", "run"}, stage)
}

func (pr *PipelineRunner) StartStage(wsId int, profile string, stage string, timeType api.Time, timeout time.Duration, verboseOutput bool) error {
	fmt.Printf("starting %s stage on workspace %d...", stage, wsId)

	err := pr.Client.StartPipelineStage(wsId, profile, stage)
	if err != nil {
		fmt.Println()
		return fmt.Errorf("failed to start pipeline stage %s: %w", stage, err)
	}

	err = pr.waitForPipelineStage(wsId, stage, timeType, timeout, verboseOutput)
	if err != nil {
		return fmt.Errorf("failed waiting for stage %s to finish: %w", stage, err)

	}
	return nil
}

func (pr *PipelineRunner) waitForPipelineStage(wsId int, stage string, timeType api.Time, timeout time.Duration, verboseOutput bool) error {
	delay := 5 * time.Second

	maxWaitTime := timeType.Now().Add(timeout)
	for {
		status, err := pr.Client.GetPipelineState(wsId, stage)
		if err != nil {
			log.Printf("\nError getting pipeline status: %s, trying again...", err.Error())
			timeType.Sleep(delay)
			continue
		}

		if allFinished(status, verboseOutput) {
			log.Println("(finished)")
			break
		}

		if allRunning(status) && stage == "run" {
			log.Println("(running)")
			break
		}

		err = shouldAbort(status)
		if err != nil {
			log.Println("(failed)")
			return fmt.Errorf("stage %s failed: %w", stage, err)
		}

		fmt.Print(".")
		if timeType.Now().After(maxWaitTime) {
			log.Println()
			return fmt.Errorf("timed out waiting for pipeline stage %s to be complete", stage)
		}
		timeType.Sleep(delay)
	}
	return nil
}

func allRunning(status []api.PipelineStatus) bool {
	for _, s := range status {
		// Run stage is only running customer servers, ignore IDE server
		if s.Server != IdeServer && s.State != "running" {
			return false
		}
	}
	return true
}

func allFinished(status []api.PipelineStatus, verboseOutput bool) bool {
	for _, s := range status {
		io.Verbosef(verboseOutput, "Server: %s, State: %s, Replica: %s\n", s.Server, s.State, s.Replica)
	}
	for _, s := range status {
		// Prepare and Test stage is only running in the IDE server, ignore customer servers
		if s.Server == IdeServer && s.State != "success" {
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
