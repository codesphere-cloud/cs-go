// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"fmt"

	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/codesphere-cloud/cs-go/pkg/export"
	"github.com/codesphere-cloud/cs-go/pkg/git"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type GenerateDockerfileCmd struct {
	cmd  *cobra.Command
	Opts GenerateDockerfileOpts
}

type GenerateDockerfileOpts struct {
	GlobalOptions
	Input     *string
	Branch    *string
	BaseImage *string
	Output    *string
	Env       *[]string
}

func (c *GenerateDockerfileCmd) RunE(_ *cobra.Command, args []string) error {
	client, err := NewClient(c.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	fs := cs.NewOSFileSystem(".")
	exp := export.NewExporterService(fs)
	git := git.NewGitService(fs)

	if err := c.GenerateDockerfile(client, fs, exp, git); err != nil {
		return fmt.Errorf("failed to generate dockerfile: %w", err)
	}

	fmt.Println("Dockerfile created:")
	fmt.Printf("\nInput: %d\n", c.Opts.Input)
	fmt.Printf("\nOutput: %d\n", c.Opts.Output)
	fmt.Printf("To run it you can use 'cd %d && docker compose up'", c.Opts.Output)

	return nil
}

func AddGenerateDockerfileCmd(generate *cobra.Command) {
	dockerfile := GenerateDockerfileCmd{
		cmd: &cobra.Command{
			Use:   "dockerfile",
			Short: "Generates a dockerfile based on a ci.yml of a workspace",
			Long: io.Long(`If the input file is not found, cs will attempt to clone the repository of the workspace
				on your local machine to run the artifact generation.
				For that a folder will be generated containing the repository and the generated artifacts.

				The export then generates a subdirectory containing the following artifacts:

				./<service-n> Each service is exported to a separate folder.
				./<service-n>/Dockerfile Dockerfile to build the container of the service.
				./<service-n>/entrypoint.sh Entrypoint of the container (run stage of Codesphere workspace).
				./docker-compose.yml Environment to allow running the services with docker-compose.
				./export/nginx.conf Configuration for NGINX, which is used by as router between services.

				Codesphere recommends adding the generated artifacts to the source code repository.`),
			Example: io.FormatExampleCommands("generate dockerfile", []io.Example{
				{Cmd: "-w 1234", Desc: "Generate dockerfile for workspace 1234"},
				{Cmd: "-w 1234 -i ci.prod.yml", Desc: "Generate dockerfile for workspace 1234 based on ci profile ci.prod.yml"},
			}),
		},
	}
	dockerfile.Opts.Input = dockerfile.cmd.Flags().StringP("input", "i", "ci.yml", "CI profile to use as input for generation")
	dockerfile.Opts.Branch = dockerfile.cmd.Flags().String("branch", "main", "Branch of the repository to clone if the input file is not found")
	dockerfile.Opts.BaseImage = dockerfile.cmd.Flags().StringP("baseimage", "b", "", "Base image for the dockerfile")
	dockerfile.Opts.Output = dockerfile.cmd.Flags().StringP("output", "o", "./export", "Output path of the folder including generated artifacts")
	dockerfile.Opts.Env = dockerfile.cmd.Flags().StringArrayP("env", "e", []string{}, "Env vars to put into generated artifacts")

	generate.AddCommand(dockerfile.cmd)
	dockerfile.cmd.RunE = dockerfile.RunE
}

func (c *GenerateDockerfileCmd) GenerateDockerfile(client Client, fs *cs.FileSystem, exporter export.Exporter, git git.Git) error {
	if c.Opts.Input == nil || *c.Opts.Input == "" {
		return errors.New("input file is required")
	}
	if c.Opts.BaseImage == nil || *c.Opts.BaseImage == "" {
		return errors.New("baseimage is required")
	}

	envs := []string{}
	if c.Opts.Env != nil {
		envs = *c.Opts.Env
	}

	if !fs.FileExists(*c.Opts.Input) {
		fmt.Printf("Input file %s not found, attempting to clone repository...\n", *c.Opts.Input)
		if err := c.CloneRepository(client, fs, git); err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}
		if !fs.FileExists(*c.Opts.Input) {
			return fmt.Errorf("input file %s not found after cloning repository", *c.Opts.Input)
		}
	}

	err := exporter.ExportDockerArtifacts(*c.Opts.Input, *c.Opts.Output, *c.Opts.BaseImage, envs)
	if err != nil {
		return fmt.Errorf("failed to export docker artifacts: %w", err)
	}

	return nil
}

func (c *GenerateDockerfileCmd) CloneRepository(client Client, fs *cs.FileSystem, git git.Git) error {
	wsId, err := c.Opts.GetWorkspaceId()
	if err != nil {
		return fmt.Errorf("failed to get workspace ID: %w", err)
	}

	ws, err := client.GetWorkspace(wsId)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}
	if ws.GitUrl.Get() == nil || *ws.GitUrl.Get() == "" {
		return errors.New("workspace does not have a git repository")
	}

	repoUrl := *ws.GitUrl.Get()
	repoBranch := "main"
	if c.Opts.Branch != nil && *c.Opts.Branch != "" {
		repoBranch = *c.Opts.Branch
	} else if ws.InitialBranch.Get() != nil {
		repoBranch = *ws.InitialBranch.Get()
	}

	_, err = git.CloneRepository(fs, repoUrl, repoBranch)
	if err != nil {
		return fmt.Errorf("failed to clone repository %s: %w", repoUrl, err)
	}

	fmt.Printf("Repository %s, branch %s cloned.\n", repoUrl, repoBranch)
	return nil
}
