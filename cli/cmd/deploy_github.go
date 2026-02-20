// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/codesphere-cloud/cs-go/pkg/deploy"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type DeployGitHubCmd struct {
	cmd  *cobra.Command
	Opts DeployGitHubOpts
}

type DeployGitHubOpts struct {
	GlobalOptions
	PlanId    *int
	Env       *[]string
	VpnConfig *string
	Branch    *string
	Stages    *string
	Timeout   *time.Duration
	Profile   *string
}

func (c *DeployGitHubCmd) RunE(_ *cobra.Command, args []string) error {
	client, err := NewClient(c.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	teamId, err := c.Opts.GetTeamId()
	if err != nil {
		return fmt.Errorf("failed to get team ID: %w", err)
	}

	// Load GitHub context
	eventName := os.Getenv("GITHUB_EVENT_NAME")
	prAction, prNumber := loadGitHubEvent()
	repository := os.Getenv("GITHUB_REPOSITORY")
	serverUrl := os.Getenv("GITHUB_SERVER_URL")

	// Determine workspace name: <repo>-#<pr>
	parts := strings.Split(repository, "/")
	repo := parts[len(parts)-1]
	wsName := fmt.Sprintf("%s-#%s", repo, prNumber)

	// Resolve branch
	branch := c.resolveBranch()

	// Resolve repo URL
	repoUrl := fmt.Sprintf("%s/%s.git", serverUrl, repository)

	// Parse stages
	var stages []string
	for _, s := range strings.Fields(*c.Opts.Stages) {
		if s != "" {
			stages = append(stages, s)
		}
	}

	// Parse env vars
	envVars := make(map[string]string)
	for _, e := range *c.Opts.Env {
		if idx := strings.Index(e, "="); idx > 0 {
			envVars[e[:idx]] = e[idx+1:]
		}
	}

	cfg := deploy.Config{
		TeamId:    teamId,
		PlanId:    *c.Opts.PlanId,
		Name:      wsName,
		EnvVars:   envVars,
		VpnConfig: *c.Opts.VpnConfig,
		Branch:    branch,
		Stages:    stages,
		RepoUrl:   repoUrl,
		Timeout:   *c.Opts.Timeout,
		Profile:   *c.Opts.Profile,
		ApiUrl:    c.Opts.GetApiUrl(),
	}

	// Determine if this is a delete operation
	isDelete := eventName == "pull_request" && prAction == "closed"

	deployer := deploy.NewDeployer(client)
	result, err := deployer.Deploy(cfg, isDelete)
	if err != nil {
		return err
	}

	// Write GitHub-specific outputs
	if result != nil {
		setGitHubOutputs(result.WorkspaceId, result.WorkspaceURL)
	}

	return nil
}

// resolveBranch determines the branch to deploy with priority:
// flag > GITHUB_HEAD_REF > GITHUB_REF_NAME > "main"
func (c *DeployGitHubCmd) resolveBranch() string {
	if c.Opts.Branch != nil && *c.Opts.Branch != "" {
		return *c.Opts.Branch
	}
	if headRef := os.Getenv("GITHUB_HEAD_REF"); headRef != "" {
		return headRef
	}
	if refName := os.Getenv("GITHUB_REF_NAME"); refName != "" {
		return refName
	}
	return "main"
}

// loadGitHubEvent reads the PR action and number from GITHUB_EVENT_PATH.
func loadGitHubEvent() (action string, number string) {
	path := os.Getenv("GITHUB_EVENT_PATH")
	if path == "" {
		return "", ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", ""
	}
	var event struct {
		Action string `json:"action"`
		Number int    `json:"number"`
	}
	if json.Unmarshal(data, &event) == nil {
		return event.Action, strconv.Itoa(event.Number)
	}
	return "", ""
}

// setGitHubOutputs writes deployment results to GitHub Actions output files.
func setGitHubOutputs(wsId int, url string) {
	if f := os.Getenv("GITHUB_OUTPUT"); f != "" {
		appendToFile(f, fmt.Sprintf("deployment-url=%s\nworkspace-id=%d\n", url, wsId))
	}

	if f := os.Getenv("GITHUB_STEP_SUMMARY"); f != "" {
		appendToFile(f, fmt.Sprintf(
			"### 🚀 Codesphere Deployment\n\n| Property | Value |\n|----------|-------|\n| **URL** | [%s](%s) |\n| **Workspace** | `%d` |\n",
			url, url, wsId,
		))
	}
}

func appendToFile(path, content string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close() //nolint:errcheck // best-effort append
	_, _ = f.WriteString(content)
}

func AddDeployGitHubCmd(deployCmd *cobra.Command, opts GlobalOptions) {
	github := DeployGitHubCmd{
		cmd: &cobra.Command{
			Use:   "github",
			Short: "Deploy from GitHub Actions",
			Long: io.Long(`Deploy workspaces from GitHub Actions.

				Automatically detects the PR context from GitHub Actions environment
				variables (GITHUB_EVENT_NAME, GITHUB_HEAD_REF, GITHUB_REPOSITORY, etc.)
				and creates, updates, or deletes workspaces accordingly.

				On PR open/synchronize: creates or updates a workspace.
				On PR close: deletes the workspace.

				Designed to be used from GitHub Actions workflows.`),
			Example: io.FormatExampleCommands("deploy github", []io.Example{
				{Cmd: "", Desc: "Deploy using GitHub Actions environment variables"},
				{Cmd: "--plan-id 20", Desc: "Deploy with a specific plan"},
				{Cmd: "--stages 'prepare test run'", Desc: "Deploy and run specific pipeline stages"},
				{Cmd: "--branch feature-x", Desc: "Override the branch to deploy"},
			}),
		},
		Opts: DeployGitHubOpts{GlobalOptions: opts},
	}

	github.Opts.PlanId = github.cmd.Flags().Int("plan-id", 8, "Plan ID for the workspace")
	github.Opts.Env = github.cmd.Flags().StringArray("env", []string{}, "Environment variables in KEY=VALUE format")
	github.Opts.VpnConfig = github.cmd.Flags().String("vpn-config", "", "VPN config name to connect the workspace to")
	github.Opts.Branch = github.cmd.Flags().StringP("branch", "b", "", "Git branch to deploy (auto-detected from GitHub context if not set)")
	github.Opts.Stages = github.cmd.Flags().String("stages", "prepare run", "Pipeline stages to run (space-separated: prepare test run)")
	github.Opts.Timeout = github.cmd.Flags().Duration("timeout", 5*time.Minute, "Timeout for workspace creation/readiness")
	github.Opts.Profile = github.cmd.Flags().StringP("profile", "p", "", "CI profile to use (e.g. 'prod' for ci.prod.yml), defaults to ci.yml")

	deployCmd.AddCommand(github.cmd)
	github.cmd.RunE = github.RunE
}
