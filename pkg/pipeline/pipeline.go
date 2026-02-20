// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package pipeline

import (
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/codesphere-cloud/cs-go/api"
)

const IdeServer string = "codesphere-ide"

// Client defines the API operations needed for pipeline execution.
type Client interface {
	StartPipelineStage(wsId int, profile string, stage string) error
	GetPipelineState(wsId int, stage string) ([]api.PipelineStatus, error)
	DeployLandscape(wsId int, profile string) error
}

// Config holds parameters for pipeline execution.
type Config struct {
	Profile string
	Timeout time.Duration
}

// Runner orchestrates pipeline stage execution.
type Runner struct {
	Client Client
	Time   api.Time
}

// NewRunner creates a new pipeline runner with the given API client.
func NewRunner(client Client, clock api.Time) *Runner {
	if clock == nil {
		clock = &api.RealTime{}
	}
	return &Runner{Client: client, Time: clock}
}

// RunStages runs pipeline stages sequentially: prepare and test are awaited,
// the run stage is preceded by a landscape sync and then fire-and-forget.
func (r *Runner) RunStages(wsId int, stages []string, cfg Config) error {
	for _, stage := range stages {
		if !IsValidStage(stage) {
			return fmt.Errorf("invalid pipeline stage: %s", stage)
		}
	}

	for _, stage := range stages {
		// Sync the landscape before the run stage
		if stage == "run" {
			fmt.Println("  🔄 Syncing landscape...")
			if err := r.Client.DeployLandscape(wsId, cfg.Profile); err != nil {
				return fmt.Errorf("syncing landscape: %w", err)
			}
			fmt.Println("  ✅ Landscape synced.")
		}

		if err := r.runStage(wsId, stage, cfg); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) runStage(wsId int, stage string, cfg Config) error {
	log.Printf("starting %s stage on workspace %d...", stage, wsId)

	if err := r.Client.StartPipelineStage(wsId, cfg.Profile, stage); err != nil {
		log.Println()
		return fmt.Errorf("failed to start pipeline stage %s: %w", stage, err)
	}

	return r.waitForStage(wsId, stage, cfg)
}

func (r *Runner) waitForStage(wsId int, stage string, cfg Config) error {
	delay := 5 * time.Second
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Minute
	}

	maxWaitTime := r.Time.Now().Add(timeout)
	for {
		status, err := r.Client.GetPipelineState(wsId, stage)
		if err != nil {
			log.Printf("\nError getting pipeline status: %s, trying again...", err.Error())
			r.Time.Sleep(delay)
			continue
		}

		if AllFinished(status) {
			log.Println("(finished)")
			break
		}

		if AllRunning(status) && stage == "run" {
			log.Println("(running)")
			break
		}

		if err = ShouldAbort(status); err != nil {
			log.Println("(failed)")
			return fmt.Errorf("stage %s failed: %w", stage, err)
		}

		log.Print(".")
		if r.Time.Now().After(maxWaitTime) {
			log.Println()
			return fmt.Errorf("timed out waiting for pipeline stage %s to be complete", stage)
		}
		r.Time.Sleep(delay)
	}
	return nil
}

// IsValidStage returns true if the given stage name is valid.
func IsValidStage(stage string) bool {
	return slices.Contains([]string{"prepare", "test", "run"}, stage)
}

// AllFinished returns true when all IDE server replicas have succeeded.
// Prepare and test stages only run in the IDE server; customer servers are ignored.
func AllFinished(status []api.PipelineStatus) bool {
	for _, s := range status {
		if s.Server == IdeServer && s.State != "success" {
			return false
		}
	}
	return true
}

// AllRunning returns true when all customer server replicas are running.
// The IDE server is ignored since the run stage only applies to customer servers.
func AllRunning(status []api.PipelineStatus) bool {
	for _, s := range status {
		if s.Server != IdeServer && s.State != "running" {
			return false
		}
	}
	return true
}

// ShouldAbort returns an error if any replica has reached a terminal failure state.
func ShouldAbort(status []api.PipelineStatus) error {
	for _, s := range status {
		if slices.Contains([]string{"failure", "aborted"}, s.State) {
			return fmt.Errorf("server %s, replica %s reached unexpected state %s", s.Server, s.Replica, s.State)
		}
	}
	return nil
}
