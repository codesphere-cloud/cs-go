package api

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func (client *RealClient) WakeUpWorkspace(wsId int, token string, profile string, timeout time.Duration) error {
	workspace, err := client.GetWorkspace(wsId)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Check if workspace is already running
	status, err := client.WorkspaceStatus(wsId)
	if err != nil {
		return fmt.Errorf("failed to get workspace status: %w", err)
	}

	if !status.IsRunning {
		log.Printf("Waking up workspace %d (%s)...\n", wsId, workspace.Name)

		// Scale workspace to at least 1 replica to wake it up
		// If workspace already has replicas configured (but not running), preserve that count
		targetReplicas := 1
		if workspace.Replicas > 1 {
			targetReplicas = workspace.Replicas
		}

		err = client.ScaleWorkspace(wsId, targetReplicas)
		if err != nil {
			return fmt.Errorf("failed to scale workspace: %w", err)
		}

		log.Printf("Waiting for workspace %d to be running...\n", wsId)
		err = client.WaitForWorkspaceRunning(&workspace, timeout)
		if err != nil {
			return fmt.Errorf("workspace did not become running: %w", err)
		}
	} else {
		log.Printf("Workspace %d (%s) is already running\n", wsId, workspace.Name)
	}

	log.Printf("Deploying landscape for workspace %d...\n", wsId)
	err = client.DeployLandscape(wsId, profile)
	if err != nil {
		return fmt.Errorf("failed to deploy landscape: %w", err)
	}
	log.Printf("Landscape deployment initiated for workspace %d\n", wsId)

	if workspace.DevDomain == nil || *workspace.DevDomain == "" {
		log.Printf("Workspace %d does not have a dev domain, skipping health check\n", wsId)
		return nil
	}

	log.Printf("Checking health of workspace %d (%s)...\n", wsId, workspace.Name)

	err = client.waitForWorkspaceHealthy(*workspace.DevDomain, token, timeout)
	if err != nil {
		return fmt.Errorf("workspace did not become healthy: %w", err)
	}

	log.Printf("Workspace %d is healthy and ready\n", wsId)

	return nil
}

func (c *RealClient) waitForWorkspaceHealthy(devDomain string, token string, timeout time.Duration) error {
	url := fmt.Sprintf("https://%s", devDomain)
	delay := 5 * time.Second
	maxWaitTime := time.Now().Add(timeout)

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("X-CS-Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := httpClient.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			// Any HTTP response (even 502) means the workspace proxy is reachable
			// and the workspace is awake. A non-200 status just means no service
			// is listening on the target port yet, which is expected for fresh workspaces.
			log.Printf("Workspace %s responded with status code %d\n", devDomain, resp.StatusCode)
			return nil
		}

		if time.Now().After(maxWaitTime) {
			return fmt.Errorf("timeout waiting for workspace to be healthy at %s", url)
		}

		time.Sleep(delay)
	}
}
