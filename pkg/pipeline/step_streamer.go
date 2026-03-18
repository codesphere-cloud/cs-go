// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package pipeline

import (
	"context"
	"log"
	"os"
	"sync"
)

// stepStreamer manages log streaming for individual pipeline steps.
// It ensures only one step streams at a time and drains the previous
// step's stream before starting the next one.
type stepStreamer struct {
	client  Client
	wsId    int
	stage   string
	enabled bool

	currentStep int
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// newStepStreamer creates a streamer for the given stage.
// Streaming is disabled for the "run" stage.
func newStepStreamer(client Client, wsId int, stage string) *stepStreamer {
	return &stepStreamer{
		client:      client,
		wsId:        wsId,
		stage:       stage,
		enabled:     stage != "run",
		currentStep: -1,
	}
}

// startStep begins streaming logs for a new step, draining any
// previous step first. It is safe to call multiple times with the
// same step number.
func (s *stepStreamer) startStep(step int, totalSteps int) {
	if !s.enabled || step <= s.currentStep {
		return
	}

	// Drain previous step before starting next
	s.drain()

	s.currentStep = step
	log.Printf("  📋 Step %d/%d", step+1, totalSteps)

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.client.StreamLogs(ctx, s.wsId, s.stage, step, os.Stdout); err != nil {
			if ctx.Err() == nil {
				log.Printf("⚠ log stream error (step %d): %v", step, err)
			}
		}
	}()
}

// drain cancels the active stream and waits for it to finish.
func (s *stepStreamer) drain() {
	if s.cancel == nil {
		return
	}
	s.cancel()
	s.wg.Wait()
	s.cancel = nil
}
