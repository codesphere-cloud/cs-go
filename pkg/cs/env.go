// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cs

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Environment struct {
}

func NewEnv() *Environment {
	return &Environment{}
}

func (e *Environment) GetApiToken() (string, error) {
	apiToken := os.Getenv("CS_TOKEN")
	if apiToken == "" {
		return "", errors.New("CS_TOKEN env var required, but not set")
	}
	return apiToken, nil
}

func (e *Environment) GetWorkspaceId() (int, error) {
	return e.ReadNumericEnv("CS_WORKSPACE_ID")
}

func (e *Environment) GetTeamId() (int, error) {
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
