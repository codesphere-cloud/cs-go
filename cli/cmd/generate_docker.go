// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"fmt"
	"path"

	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/codesphere-cloud/cs-go/pkg/exporter"
	"github.com/codesphere-cloud/cs-go/pkg/git"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type GenerateDockerCmd struct {
	cmd  *cobra.Command
	Opts *GenerateDockerOpts
}

type GenerateDockerOpts struct {
	*GenerateOpts
	BaseImage string
	Envs      []string
}

func (c *GenerateDockerCmd) RunE(cc *cobra.Command, args []string) error {
	fmt.Println(c.Opts.Force)
	fs := cs.NewOSFileSystem(".")
	git := git.NewGitService(fs)

	client, err := NewClient(c.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	exporter := exporter.NewExporterService(fs, c.Opts.Output, c.Opts.BaseImage, c.Opts.Envs, c.Opts.RepoRoot, c.Opts.Force)
	if err := c.GenerateDocker(fs, exporter, git, client); err != nil {
		return fmt.Errorf("failed to generate docker: %w", err)
	}

	fmt.Println("docker artifacts created:")
	fmt.Printf("Input: %s\n", c.Opts.Input)
	fmt.Printf("Output: %s\n", c.Opts.Output)

	fmt.Println("To start with docker-compose, run:")
	fmt.Printf("cd %s && docker compose up\n\n", c.Opts.Output)

	fmt.Println("To build and push images, run:")
	fmt.Println("export REGISTRY=<your-registry>")
	fmt.Println("export IMAGE_PREFIX=<some-prefix>")
	fmt.Printf("%s generate images --reporoot %s -r $REGISTRY -p $IMAGE_PREFIX -i %s -o %s\n\n", io.BinName(), c.Opts.RepoRoot, c.Opts.Input, c.Opts.Output)

	return nil
}

func AddGenerateDockerCmd(generate *cobra.Command, opts *GenerateOpts) {
	docker := GenerateDockerCmd{
		cmd: &cobra.Command{
			Use:   "docker",
			Short: "Generates docker artifacts based on a ci.yml of a workspace",
			Long: io.Long(`The generated artifacts will be saved in the output folder (default is ./export).
				It then generates following artifacts inside the output folder:

				./<service-n> Each service is exported to a separate folder.
				./<service-n>/dockerfile docker to build the container of the service.
				./<service-n>/entrypoint.sh Entrypoint of the container (run stage of Codesphere workspace).
				./docker-compose.yml Environment to allow running the services with docker-compose.
				./nginx.conf Configuration for NGINX, which is used by as router between services.

				Codesphere recommends adding the generated artifacts to the source code repository.

				Limitations:
				- Environment variables have to be set explicitly as the Codesphere environment has its own way to provide env variables.
				- The workspace ID, team ID etc. are not automatically available and have to be set explicitly.
				- Hardcoded workspace urls don't work outside of the Codesphere environment.
				- Each dockerfile of your services contain all prepare steps. To have the smallest image possible you would have to delete all unused steps in each service.
				`),
			Example: io.FormatExampleCommands("generate docker", []io.Example{
				{Cmd: "-w 1234", Desc: "Generate docker for workspace 1234"},
				{Cmd: "-w 1234 -i ci.prod.yml", Desc: "Generate docker for workspace 1234 based on ci profile ci.prod.yml"},
			}),
		},
		Opts: &GenerateDockerOpts{
			GenerateOpts: opts,
		},
	}
	docker.cmd.Flags().StringVarP(&docker.Opts.BaseImage, "baseimage", "b", "", "Base image for the docker")
	docker.cmd.Flags().StringArrayVarP(&docker.Opts.Envs, "env", "e", []string{}, "Env vars to put into generated artifacts")

	generate.AddCommand(docker.cmd)
	docker.cmd.RunE = docker.RunE
}

func (c *GenerateDockerCmd) GenerateDocker(fs *cs.FileSystem, exp exporter.Exporter, git git.Git, csClient Client) error {
	if c.Opts.BaseImage == "" {
		return errors.New("baseimage is required")
	}

	ciInput := path.Join(c.Opts.RepoRoot, c.Opts.Input)
	if !fs.FileExists(ciInput) {
		fmt.Printf("Input file %s not found attempting to clone workspace repository...\n", c.Opts.Input)

		if err := c.CloneRepository(csClient, fs, git, c.Opts.RepoRoot); err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}
		if !fs.FileExists(ciInput) {
			return fmt.Errorf("input file %s not found after cloning repository", c.Opts.Input)
		}
	}

	_, err := exp.ReadYmlFile(ciInput)
	if err != nil {
		return fmt.Errorf("failed to export docker artifacts: %w", err)
	}

	err = exp.ExportDockerArtifacts()
	if err != nil {
		return fmt.Errorf("failed to export docker artifacts: %w", err)
	}

	return nil
}

func (c *GenerateDockerCmd) CloneRepository(client Client, fs *cs.FileSystem, git git.Git, clonedir string) error {
	fmt.Printf("Cloning repository into %s...\n", clonedir)

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
	if c.Opts.Branch != "" {
		repoBranch = c.Opts.Branch
	} else if ws.InitialBranch.Get() != nil {
		repoBranch = *ws.InitialBranch.Get()
	}

	_, err = git.CloneRepository(fs, repoUrl, repoBranch, clonedir)
	if err != nil {
		return fmt.Errorf("failed to clone repository %s branch %s: %w", repoUrl, repoBranch, err)
	}

	fmt.Printf("Repository %s, branch %s cloned.\n", repoUrl, repoBranch)
	return nil
}
