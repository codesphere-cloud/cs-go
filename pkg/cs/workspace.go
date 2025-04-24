package cs

import (
	"time"
)

type WaitForWorkspaceRunningOptions struct {
	Timeout time.Duration
	Delay   time.Duration
}

// Waits for a given workspace to be running.
//
// Returns [TimedOut] error if the workspace does not become running in time.

type DeployWorkspaceArgs struct {
	TeamId        int
	PlanId        int
	Name          string
	EnvVars       map[string]string
	VpnConfigName *string

	Timeout time.Duration
}
