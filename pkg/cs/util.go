// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Step struct {
	State string
}

type ReplicaStatus struct {
	State   string `json:"state"`
	Steps   []Step `json:"steps"`
	Replica string `json:"replica"`
	Server  string `json:"server"`
}

func GetPipelineStatus(ws int, stage string) (res []ReplicaStatus, err error) {

	status, err := Get(fmt.Sprintf("workspaces/%d/pipeline/%s", ws, stage))
	if err != nil {
		err = fmt.Errorf("failed to get pipeline status: %w", err)
		return
	}

	err = json.Unmarshal(status, &res)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal pipeline status: %w", err)
		return
	}
	return
}

func Get(path string) (body []byte, err error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", NewEnv().GetApiUrl(), strings.TrimPrefix(path, "/")), http.NoBody)
	if err != nil {
		err = fmt.Errorf("failed to create request: %w", err)
		return
	}
	err = SetAuthoriziationHeader(req)
	if err != nil {
		err = fmt.Errorf("failed to set header: %w", err)
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf("GET failed: %w", err)
		return
	}
	defer func() { _ = res.Body.Close() }()
	body, err = io.ReadAll(res.Body)
	return
}

func SetAuthoriziationHeader(req *http.Request) error {
	token, err := NewEnv().GetApiToken()
	if err != nil {
		return fmt.Errorf("failed to get API token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func GetRoleName(role int) string {
	if role == 1 {
		return "Member"
	}
	return "Admin"
}

func ArgToEnvVarMap(input []string) (map[string]string, error) {
	res := map[string]string{}
	for _, v := range input {
		split := strings.Split(v, "=")
		if len(split) != 2 {
			return res, fmt.Errorf("invalid environment variable argument: %s", v)

		}
		res[split[0]] = split[1]
	}
	return res, nil
}
