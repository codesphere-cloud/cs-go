// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/codesphere-cloud/cs-go/api"
	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type CreateWorkspaceCmd struct {
	cmd  *cobra.Command
	Opts CreateWorkspaceOpts
}

type CreateWorkspaceOpts struct {
	GlobalOptions
	Repo            *string
	Vpn             *string
	Env             *[]string
	Plan            *int
	Private         *bool
	Timeout         *time.Duration
	Branch          *string
	Baseimage       *string
	PublicDevDomain *bool
}

func (c *CreateWorkspaceCmd) RunE(_ *cobra.Command, args []string) error {
	client, err := NewClient(c.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	teamId, err := c.Opts.GetTeamId()
	if err != nil {
		return fmt.Errorf("failed to get team ID: %w", err)
	}

	if len(args) != 1 {
		return errors.New("workspace name not set")
	}
	wsName := args[0]

	ws, err := c.CreateWorkspace(client, teamId, wsName)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	giturl := ""
	if ws.GitUrl.Get() != nil {
		giturl = *ws.GitUrl.Get()
	}
	branch := ""
	if ws.InitialBranch.Get() != nil {
		branch = *ws.InitialBranch.Get()
	}

	fmt.Println("Workspace created:")
	fmt.Printf("\nID: %d\n", ws.Id)
	fmt.Printf("Name: %s\n", ws.Name)
	fmt.Printf("Team ID: %d\n", ws.TeamId)
	fmt.Printf("Git Repository: %s\n", giturl)
	fmt.Printf("Branch: %s\n", branch)
	fmt.Printf("To open it in the Codesphere IDE run '%s open workspace -w %d'", os.Args[0], ws.Id)

	return nil
}

func AddCreateWorkspaceCmd(create *cobra.Command, opts GlobalOptions) {
	workspace := CreateWorkspaceCmd{
		cmd: &cobra.Command{
			Use:   "workspace",
			Short: "Create a workspace",
			Args:  cobra.RangeArgs(1, 1),
			Long: io.Long(`Create a workspace in Codesphere.

				Specify a (private) git repository or start an empty workspace.
				Environment variables can be set to initialize the workspace with a specific environment.
				The command will wait for the workspace to become running or a timeout is reached.

				To decide which plan suits your needs, run 'cs list plans'
			`),
			Example: io.FormatExampleCommands("create workspace my-workspace", []io.Example{
				{Cmd: "-p 20", Desc: "Create an empty workspace, using plan 20"},
				{Cmd: "--public-dev-domain=false", Desc: "Create a workspace with a publicly accessible API"},
				{Cmd: "-r https://github.com/codesphere-cloud/landingpage-temp.git", Desc: "Create a workspace from a git repository"},
				{Cmd: "-r https://github.com/codesphere-cloud/landingpage-temp.git -e DEPLOYMENT=prod -e A=B", Desc: "Create a workspace and set environment variables"},
				{Cmd: "-r https://github.com/codesphere-cloud/landingpage-temp.git --vpn myVpn", Desc: "Create a workspace and connect to VPN myVpn"},
				{Cmd: "-r https://github.com/codesphere-cloud/landingpage-temp.git --timeout 30s", Desc: "Create a workspace and wait 30 seconds for it to become running"},
				{Cmd: "-r https://github.com/codesphere-cloud/landingpage-temp.git -b staging", Desc: "Create a workspace from branch 'staging'"},
				{Cmd: "-r https://github.com/my-org/my-private-project.git -P", Desc: "Create a workspace from a private git repository"},
			}),
		},
		Opts: CreateWorkspaceOpts{GlobalOptions: opts},
	}
	workspace.Opts.Repo = workspace.cmd.Flags().StringP("repository", "r", "", "Git repository to create the workspace from")
	workspace.Opts.Vpn = workspace.cmd.Flags().String("vpn", "", "Vpn config to use")
	workspace.Opts.Env = workspace.cmd.Flags().StringArrayP("env", "e", []string{}, "Environment variables to set in the workspace in key=value form (e.g. --env DEPLOYMENT=prod)")
	workspace.Opts.Plan = workspace.cmd.Flags().IntP("plan", "p", 8, "Plan ID for the workspace")
	workspace.Opts.Private = workspace.cmd.Flags().BoolP("private", "P", false, "Use private repository")
	workspace.Opts.Timeout = workspace.cmd.Flags().Duration("timeout", 10*time.Minute, "Time to wait for the workspace to start (e.g. 5m for 5 minutes)")
	workspace.Opts.Branch = workspace.cmd.Flags().StringP("branch", "b", "", "branch to check out")
	workspace.Opts.Baseimage = workspace.cmd.Flags().String("base-image", "", "Base image to use for the workspace, e.g. 'ubuntu-24.04'")
	workspace.Opts.PublicDevDomain = workspace.cmd.Flags().Bool("public-dev-domain", false, "Whether to create enable a public development domain (defaults to the public api default)")

	create.AddCommand(workspace.cmd)
	workspace.cmd.RunE = workspace.RunE
}

func (c *CreateWorkspaceCmd) CreateWorkspace(client Client, teamId int, wsName string) (*api.Workspace, error) {
	envVars, err := cs.ArgToEnvVarMap(*c.Opts.Env)
	if err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	args := api.DeployWorkspaceArgs{
		TeamId:  teamId,
		PlanId:  *c.Opts.Plan,
		Name:    wsName,
		EnvVars: envVars,

		IsPrivateRepo: *c.Opts.Private,

		Timeout: *c.Opts.Timeout,
	}

	if c.Opts.Repo != nil && *c.Opts.Repo != "" {
		validatedUrl, err := cs.ValidateUrl(*c.Opts.Repo)
		if err != nil {
			return nil, fmt.Errorf("validation of repository URL failed: %w", err)
		}
		args.GitUrl = &validatedUrl
	}

	if c.Opts.Vpn != nil && *c.Opts.Vpn != "" {
		args.VpnConfigName = c.Opts.Vpn
	}

	if c.Opts.Branch != nil && *c.Opts.Branch != "" {
		args.Branch = c.Opts.Branch
	}

	if c.Opts.Baseimage != nil && *c.Opts.Baseimage != "" {
		baseimages, err := client.ListBaseimages()
		if err != nil {
			return nil, fmt.Errorf("failed to list base images: %w", err)
		}

		baseimageNames := make([]string, len(baseimages))
		for i, bi := range baseimages {
			baseimageNames[i] = bi.GetId()
		}

		if !slices.Contains(baseimageNames, *c.Opts.Baseimage) {
			return nil, fmt.Errorf("base image '%s' not found, available options are: %s", *c.Opts.Baseimage, strings.Join(baseimageNames, ", "))
		}

		args.BaseImage = c.Opts.Baseimage
	}

	if c.cmd != nil && c.cmd.Flag("public-dev-domain").Changed && c.Opts.PublicDevDomain != nil {
		var public bool = !*c.Opts.PublicDevDomain
		args.Restricted = &public
	}

	ws, err := client.DeployWorkspace(args)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	return ws, nil
}
