// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/codesphere-cloud/cs-go/pkg/git"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/codesphere-cloud/cs-go/pkg/util"
	"github.com/spf13/cobra"
)

// UpCmd represents the up command
type UpCmd struct {
	cmd   *cobra.Command
	Opts  *UpOptions
	State *cs.UpState
}

type UpOptions struct {
	*GlobalOptions

	DomainTypeString string
	RepoAccessString string
	Yes              bool
	Verbose          bool
}

func randomAdjective() string {
	adjectives := []string{"amazing", "incredible", "fantastic", "wonderful", "awesome", "brilliant", "marvelous", "spectacular", "fabulous", "magnificent"}
	return adjectives[rand.Intn(len(adjectives))]
}

func randomNoun() string {
	nouns := []string{"workspace", "deployment", "project", "environment", "instance", "code", "work", "app", "service", "application"}
	return nouns[rand.Intn(len(nouns))]
}

func (c *UpCmd) RunE(_ *cobra.Command, args []string) error {
	log.Printf("Deploying your %s %s ...", randomAdjective(), randomNoun())

	client, err := NewClient(*c.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	fs := util.NewOSFileSystem("./")
	gitSvc := git.NewGitService(fs)

	if c.Opts.DomainTypeString != "" {
		if c.Opts.DomainTypeString != string(cs.PrivateDevDomain) && c.Opts.DomainTypeString != string(cs.PublicDevDomain) {
			return fmt.Errorf("invalid value for --public-dev-domain: %s, allowed values are 'public' or 'private'", c.Opts.DomainTypeString)
		}
		c.State.DomainType = cs.DomainType(c.Opts.DomainTypeString)
	}

	if c.Opts.RepoAccessString != "" {
		if c.Opts.RepoAccessString != string(cs.PublicRepo) && c.Opts.RepoAccessString != string(cs.PrivateRepo) {
			return fmt.Errorf("invalid value for --private-repo: %s, allowed values are 'public' or 'private'", c.Opts.RepoAccessString)
		}
		c.State.RepoAccess = cs.RepoAccess(c.Opts.RepoAccessString)
	}

	c.State.TeamId, err = c.Opts.GetTeamId()
	if err != nil {
		return fmt.Errorf("failed to get team ID: %w", err)
	}

	err = c.State.Load(c.Opts.StateFile, &api.RealTime{}, fs)
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	err = c.State.Save()
	if err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	apiToken, err := c.Opts.Env().GetApiToken()
	if err != nil {
		return fmt.Errorf("failed to get API token: %w", err)
	}

	return cs.Up(client, gitSvc, &api.RealTime{}, fs, c.State, apiToken, c.Opts.Yes, c.Opts.Verbose)
}

func AddUpCmd(rootCmd *cobra.Command, opts *GlobalOptions) {
	up := UpCmd{
		cmd: &cobra.Command{
			Use:   "up",
			Short: "Deploy your local code to Codesphere",
			Long: io.Long(`Deploys your local code to a new or existing Codepshere workspace.

			Prerequisite: Your code needs to be located in a git repository where you can create and push WIP branches.
			When running cs up, the cs CLI will do the following:

			* Push local changes to a branch (customizable with -b), if none is specified, cs go creates a WIP branch (will be stored and reused in .cs-up.yaml)
			* Create workspace if it doesn't exist yet (state in .cs-up.yaml)
			* Start the deployment in Codesphere, customize which profile to use with -p flag (defaults to 'ci.yml')
			* Print the dev domain of the workspace to the console once the deployment is successful`),
		},
		State: &cs.UpState{},
		Opts:  &UpOptions{GlobalOptions: opts},
	}
	rootCmd.AddCommand(up.cmd)
	up.cmd.RunE = up.RunE
	up.cmd.Flags().StringVarP(&up.State.Profile, "profile", "p", "", "CI profile to use (e.g. 'ci.dev.yml' for a dev profile, you may have defined in 'ci.dev.yml'), defaults to the ci.yml profile")
	up.cmd.Flags().DurationVar(&up.State.Timeout, "timeout", 0, "Timeout for the deployment process, e.g. 10m, 1h, defaults to 1m")
	up.cmd.Flags().IntVarP(&up.State.Plan, "plan", "", -1, "Plan ID to use for the workspace, if not set, the first available plan will be used")
	up.cmd.Flags().StringArrayVarP(&up.State.Env, "env", "e", []string{}, "Environment variables to set in the format KEY=VALUE, can be specified multiple times for multiple variables")
	up.cmd.Flags().StringVarP(&up.State.Branch, "branch", "b", "", "Branch to push to, if not set, a WIP branch will be created and reused for subsequent runs")
	up.cmd.Flags().StringVarP(&up.State.WorkspaceName, "workspace-name", "", "", "Name of the workspace to create, if not set, a random name will be generated")
	up.cmd.Flags().StringVarP(&up.State.BaseImage, "base-image", "", "", "Base image to use for the workspace, if not set, the default base image will be used")
	up.cmd.Flags().StringVarP(&up.Opts.DomainTypeString, "public-dev-domain", "", "", "Whether to create a public or private dev domain for the workspace (only applies to new workspaces), defaults to 'public'")
	up.cmd.Flags().StringVarP(&up.Opts.RepoAccessString, "private-repo", "", "", "Whether the git repository is public or private (requires authentication), defaults to 'public'")
	up.cmd.Flags().StringVarP(&up.State.Remote, "remote", "", "origin", "Git remote to use for pushing the code, defaults to 'origin'")
	up.cmd.Flags().BoolVarP(&up.Opts.Yes, "yes", "y", false, "Skip confirmation prompt for pushing changes to the git repository")
	up.cmd.Flags().BoolVarP(&up.Opts.Verbose, "verbose", "v", false, "Enable verbose output")
}
