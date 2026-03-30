// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cs

import (
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/api/errors"
	"github.com/codesphere-cloud/cs-go/pkg/git"
	"github.com/codesphere-cloud/cs-go/pkg/pipeline"
	"github.com/codesphere-cloud/cs-go/pkg/util"
)

type CodesphereDeploymentManager struct {
	Client          api.Client
	GitSvc          git.Git
	FileSys         *util.FileSystem
	State           *UpState
	Verbose         bool
	AskConfirmation bool
	ApiToken        string
	Time            api.Time
}

// Up deploys the code to Codesphere. It checks if the workspace already exists, if yes, it wakes it up and deploys the latest code changes,
// if not, it creates a new workspace and deploys the code.
// The workspace ID and other state is stored in .cs-up.yaml to be reused for subsequent runs.
func Up(client api.Client, gitSvc git.Git, time api.Time, fs *util.FileSystem, state *UpState, apiToken string, yes bool, verbose bool) error {
	mgr := &CodesphereDeploymentManager{
		Client:          client,
		GitSvc:          gitSvc,
		FileSys:         fs,
		State:           state,
		Verbose:         verbose,
		AskConfirmation: !yes,
		ApiToken:        apiToken,
		Time:            time,
	}

	err := mgr.UpdateGitIgnore()
	if err != nil {
		return fmt.Errorf("failed to update .gitignore: %w", err)
	}

	// Push code to branch
	log.Printf("Pushing code to branch %s repository...", state.Branch)
	err = mgr.PushChanges(!yes)
	if err != nil {
		return fmt.Errorf("failed to push branch: %w", err)
	}

	err = mgr.EnsureWorkspace()
	if err != nil {
		return fmt.Errorf("failed to ensure workspace: %w", err)
	}

	// Pull branch in workspace to update the workspace with the latest changes.
	// This is required in case the branch was created in a previous run and already exists in the remote repository.
	err = client.GitPull(state.WorkspaceId, state.Remote, state.Branch)
	if err != nil {
		return fmt.Errorf("failed to pull branch: %w", err)
	}

	err = mgr.DeployChanges()
	if err != nil {
		return fmt.Errorf("failed to deploy changes: %w", err)
	}

	pr := pipeline.NewPipelineRunner(client, state.Profile, state.Timeout, verbose)
	err = pr.StartPipelineStages(state.WorkspaceId, []string{"prepare", "test", "run"})
	if err != nil {
		return fmt.Errorf("failed to start pipeline stages: %w", err)
	}

	ws, err := client.GetWorkspace(state.WorkspaceId)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}
	devDomain := "not available"
	if ws.DevDomain != nil {
		devDomain = *ws.DevDomain
	}
	log.Printf("Workspace %d deployed successfully! Dev domain: %s\n", state.WorkspaceId, devDomain)

	return nil
}

// getPlan returns the plan ID to use for the workspace. If a plan ID is already set in the state,
// it returns that, otherwise it fetches the available plans from the API and returns the first one.
func (c *CodesphereDeploymentManager) getPlan() (int, error) {
	plan := c.State.Plan
	if plan == -1 {
		plans, err := c.Client.ListWorkspacePlans()
		if err != nil {
			return -1, fmt.Errorf("failed to list plans: %w", err)
		}
		if len(plans) == 0 {
			return -1, fmt.Errorf("no plans available for deployment")
		}
		c.State.Plan = plans[0].Id
	}
	return c.State.Plan, nil
}

// UpdateGitIgnore adds .cs-up.yaml to .gitignore if it doesn't exist to avoid committing the state file
func (c *CodesphereDeploymentManager) UpdateGitIgnore() error {
	// Add .cs-up.yaml to .gitignore if it doesn't exist to avoid committing the state file
	if c.FileSys.FileExists(".gitignore") {
		content, err := c.FileSys.ReadFile(".gitignore")
		if err != nil {
			return fmt.Errorf("failed to read .gitignore: %w", err)
		}
		if !strings.Contains(string(content), ".cs-up.yaml") {
			content = append(content, []byte("\n.cs-up.yaml\n")...)
			err = c.FileSys.WriteFile("./", ".gitignore", content, true)
			if err != nil {
				return fmt.Errorf("failed to write .gitignore: %w", err)
			}
		}
		return nil
	}
	err := c.FileSys.WriteFile("./", ".gitignore", []byte(".cs-up.yaml\n"), true)
	if err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}
	return nil
}

// PushChanges pushes the local changes to the remote branch. If the branch doesn't exist, it will be created.
func (c *CodesphereDeploymentManager) PushChanges(askConfirmation bool) error {
	// checkout branhch using git CLI, if it doesn't exist create it
	err := c.GitSvc.Checkout(c.State.Branch, false)
	if err != nil {
		err = c.GitSvc.Checkout(c.State.Branch, true)
		if err != nil {
			return fmt.Errorf("failed to checkout branch: %w", err)
		}

		err = c.GitSvc.Push(c.State.Remote, c.State.Branch)
		if err != nil {
			return fmt.Errorf("failed to push branch: %w", err)
		}
	}
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	// use git cli to add all files
	err = c.GitSvc.AddAll()
	if err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}

	// if nothing changed compared to the remote branch, skip commit and push
	hasChanges, err := c.GitSvc.HasChanges(c.State.Remote, c.State.Branch)
	if err != nil {
		return fmt.Errorf("failed to check for changes compared to remote: %w", err)
	}
	if !hasChanges {
		log.Printf("No changes to push to remote branch %s, skipping push\n", c.State.Branch)
		return nil
	}

	// print files being added and ask for confirmation before pushing
	err = c.GitSvc.PrintStatus()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	if askConfirmation {
		fmt.Print("Do you want to push these changes? (y/n): ")
		var response string
		_, err = fmt.Scanln(&response)
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		if strings.ToLower(response) != "y" {
			return fmt.Errorf("aborting push")
		}
	}

	err = c.GitSvc.Commit("cs up commit")
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	err = c.GitSvc.Push(c.State.Remote, c.State.Branch)
	if err != nil {
		return fmt.Errorf("failed to push changes: %w", err)
	}

	return nil
}

// EnsureWorkspace checks if the workspace already exists, if yes, it wakes it up, if not, it creates a new workspace.
// It also sets the environment variables on the workspace.
func (c *CodesphereDeploymentManager) EnsureWorkspace() error {
	envVars, err := ArgToEnvVarMap(c.State.Env)
	if err != nil {
		return fmt.Errorf("failed to parse environment variables: %w", err)
	}

	workspaceExists, err := c.wakeupWorkspaceIfExists(envVars)
	if err != nil {
		return fmt.Errorf("failed to wake up workspace: %w", err)
	}

	if c.State.WorkspaceId > 0 && workspaceExists {
		return nil
	}

	wsId, err := c.createWorkspace(envVars)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	c.State.WorkspaceId = wsId
	err = c.State.Save()
	if err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	log.Println("Your workspace is being deployed.")
	return nil
}

// WakeupWorkspaceIfExists checks if the workspace already exists, if yes, it wakes it up and sets the environment variables,
// returns true if the workspace was woken up, false if the workspace does not exist
func (c *CodesphereDeploymentManager) wakeupWorkspaceIfExists(envVars map[string]string) (bool, error) {
	if c.State.WorkspaceId <= 0 {
		return false, nil
	}

	log.Printf("Checking workspace with ID %d...\n", c.State.WorkspaceId)

	_, err := c.Client.GetWorkspace(c.State.WorkspaceId)
	if err != nil && !errors.IsNotFound(err) {
		return false, fmt.Errorf("failed to get workspace: %w", err)
	}
	if err != nil && errors.IsNotFound(err) {
		log.Printf("Workspace with ID %d does not exist.\n", c.State.WorkspaceId)
		return false, nil
	}

	err = c.Client.WakeUpWorkspace(c.State.WorkspaceId, c.ApiToken, c.State.Profile, c.State.Timeout)
	if err != nil {
		return false, fmt.Errorf("failed to wake up workspace: %w", err)
	}

	err = c.Client.SetEnvVarOnWorkspace(c.State.WorkspaceId, envVars)
	if err != nil {
		return false, fmt.Errorf("failed to set environment variables on workspace: %w", err)
	}
	return true, nil
}

// CreateWorkspace creates a new workspace with the specified configuration and returns the workspace ID.
func (c *CodesphereDeploymentManager) createWorkspace(envVars map[string]string) (int, error) {
	plan, err := c.getPlan()
	if err != nil {
		return -1, fmt.Errorf("failed to get plan: %w", err)
	}
	log.Println("Creating workspace ...")
	restricted := c.State.DomainType == PrivateDevDomain
	args := api.DeployWorkspaceArgs{
		TeamId:  c.State.TeamId,
		PlanId:  plan,
		Name:    c.State.WorkspaceName,
		EnvVars: envVars,

		IsPrivateRepo: c.State.RepoAccess == PrivateRepo,
		Restricted:    &restricted,

		Timeout: c.State.Timeout,
	}

	remoteUrl, err := c.GitSvc.GetRemoteUrl(c.State.Remote)
	if err != nil {
		return -1, fmt.Errorf("failed to get remote URL: %w", err)
	}

	validatedUrl, err := ValidateUrl(remoteUrl)
	if err != nil {
		return -1, fmt.Errorf("validation of repository URL failed: %w", err)
	}
	args.GitUrl = &validatedUrl

	if c.State.Branch != "" {
		args.Branch = &c.State.Branch
	}

	if c.State.BaseImage != "" {
		baseimages, err := c.Client.ListBaseimages()
		if err != nil {
			return -1, fmt.Errorf("failed to list base images: %w", err)
		}

		baseimageNames := make([]string, len(baseimages))
		for i, bi := range baseimages {
			baseimageNames[i] = bi.GetId()
		}

		if !slices.Contains(baseimageNames, c.State.BaseImage) {
			return -1, fmt.Errorf("base image '%s' not found, available options are: %s", c.State.BaseImage, strings.Join(baseimageNames, ", "))
		}

		args.BaseImage = &c.State.BaseImage
	}

	ws, err := c.Client.DeployWorkspace(args)
	if err != nil {
		return -1, fmt.Errorf("failed to create workspace: %w", err)
	}
	return ws.Id, nil
}

// DeployChanges deploys the latest code changes to the workspace.
// It retries the deployment up to 3 times in case of a server error (500).
func (c *CodesphereDeploymentManager) DeployChanges() error {
	// retrying the deployment in case it fails a 500
	for retries := 0; retries < 3; retries++ {
		err := c.Client.DeployLandscape(c.State.WorkspaceId, c.State.Profile)
		if err != nil {
			if strings.Contains(err.Error(), "500") {
				log.Printf("Deployment failed with a server error, retrying... (%d/3)\n", retries+1)
				c.Time.Sleep(5 * time.Second)
				continue
			}
			return fmt.Errorf("failed to deploy landscape: %w", err)
		}
		break
	}
	return nil
}
