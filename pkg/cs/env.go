// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cs

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/pkg/util"
)

type Env interface {
	GetApiToken() (string, error)
	GetTeamId() (int, error)
	GetWorkspaceId() (int, error)
	GetApiUrl() string
}
type Environment struct {
	statefile string
}

func NewEnv(statefile string) *Environment {
	return &Environment{
		statefile: statefile,
	}
}

func (e *Environment) GetApiToken() (string, error) {
	apiToken := os.Getenv("CS_TOKEN")
	if apiToken == "" {
		return "", errors.New("CS_TOKEN env var required, but not set")
	}
	return apiToken, nil
}

func (e *Environment) GetWorkspaceId() (int, error) {
	prefixedId, err := e.ReadNumericEnv("CS_WORKSPACE_ID")
	if prefixedId != -1 && err == nil {
		return prefixedId, nil
	}

	upState := &UpState{}
	err = upState.Load(e.statefile, &api.RealTime{}, util.NewOSFileSystem("."))

	if err == nil && upState.WorkspaceId > -1 {
		return upState.WorkspaceId, nil
	}

	return e.ReadNumericEnv("WORKSPACE_ID")
}

func (e *Environment) GetTeamId() (int, error) {
	upState := &UpState{}
	err := upState.Load(e.statefile, &api.RealTime{}, util.NewOSFileSystem("."))

	if err == nil && upState.TeamId > 0 {
		return upState.TeamId, nil
	}
	return e.ReadNumericEnv("CS_TEAM_ID")
}

func (e *Environment) ReadNumericEnv(env string) (int, error) {
	envValue := os.Getenv(env)
	if envValue == "" {
		return -1, nil
	}
	num, err := strconv.Atoi(envValue)
	if err != nil {
		return -1, fmt.Errorf("failed to convert %s %s to number: %w", env, envValue, err)
	}
	return num, nil
}

func (e *Environment) GetApiUrl() string {
	url := os.Getenv("CS_API")
	if url != "" {
		return url
	}
	return "https://codesphere.com/api"
}
