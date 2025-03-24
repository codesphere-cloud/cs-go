package cs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
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

func GetApiUrl() string {
	url := os.Getenv("CS_API")
	if url != "" {
		return url
	}
	return "https://codesphere.com/api"
}

func GetPipelineStatus(ws int, stage string) (res []ReplicaStatus, err error) {

	status, err := Get(fmt.Sprintf("workspaces/%d/pipeline/%s", ws, stage))
	if err != nil {
		err = fmt.Errorf("Failed to get pipeline status: %e", err)
		return
	}

	json.Unmarshal(status, &res)
	if err != nil {
		err = fmt.Errorf("Failed to unmarshal pipeline status: %e", err)
		return
	}
	return
}

func Get(path string) (body []byte, err error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", GetApiUrl(), strings.TrimPrefix(path, "/")), http.NoBody)
	SetAuthoriziationHeader(req)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf("GET failed: %e", err)
		return
	}
	defer res.Body.Close()
	body, err = io.ReadAll(res.Body)
	return
}

func SetAuthoriziationHeader(req *http.Request) error {

	apiToken := os.Getenv("CS_TOKEN")
	if apiToken == "" {
		return errors.New("CS_TOKEN env var required, but not set.")
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)
	return nil
}
