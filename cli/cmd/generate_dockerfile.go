package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/codesphere-cloud/cs-go/pkg/export"
	"github.com/spf13/cobra"
)

type GenerateDockerfileCmd struct {
	cmd  *cobra.Command
	Opts GenerateDockerfileOpts
}

type GenerateDockerfileOpts struct {
	GlobalOptions
	Input     *string
	BaseImage *string
	Output    *string
	Env       *[]string
}

func (c *GenerateDockerfileCmd) RunE(_ *cobra.Command, args []string) error {
	fs := cs.NewOSFileSystem(".")

	if err := c.GenerateDockerfile(fs); err != nil {
		return fmt.Errorf("failed to generate dockerfile: %w", err)
	}

	fmt.Println("Dockerfile created:")
	fmt.Printf("\nInput: %d\n", c.Opts.Input)
	fmt.Printf("\nOutput: %d\n", c.Opts.Output)
	fmt.Printf("To run it you can use 'cd %d && docker compose up'", c.Opts.Output)

	return nil
}

const shortDesc string = "Generates a dockerfile based on a ci.yml of a workspace"

func longDesc() string {
	return shortDesc + `.

  If the input file is not found, cs will attempt to clone the repository of the workspace
  on your local machine to run the artifact generation.
  For that a folder will be generated containing the repository and the generated artifacts.

  The export then generates a subdirectory containing the following artifacts:

  ./<service-n> Each service is exported to a separate folder.
  ./<service-n>/Dockerfile Dockerfile to build the container of the service.
  ./<service-n>/entrypoint.sh Entrypoint of the container (run stage of Codesphere workspace).
  ./docker-compose.yml Environment to allow running the services with docker-compose.
  ./export/nginx.conf Configuration for NGINX, which is used by as router between services.
  

  Codesphere recommends adding the generated artifacts to the source code repository.`
}

func example() string {
	return `  # Generate dockerfile for workspace 1234
  ` + os.Args[0] + ` generate dockerfile -w 1234

  # Generate dockerfile for workspace 1234 based on ci profile ci.prod.yml
  ` + os.Args[0] + ` generate dockerfile -w 1234 -i ci.prod.yml`
}

func AddGenerateDockerfileCmd(generate *cobra.Command) {
	dockerfile := GenerateDockerfileCmd{
		cmd: &cobra.Command{
			Use:     "dockerfile",
			Short:   shortDesc,
			Long:    longDesc(),
			Example: example(),
		},
	}
	dockerfile.Opts.Input = dockerfile.cmd.Flags().StringP("input", "i", "ci.yml", "CI profile to use as input for generation")
	dockerfile.Opts.BaseImage = dockerfile.cmd.Flags().StringP("baseimage", "b", "", "Base image for the dockerfile")
	dockerfile.Opts.Output = dockerfile.cmd.Flags().StringP("output", "o", "./export", "Output path of the folder including generated artifacts")
	dockerfile.Opts.Env = dockerfile.cmd.Flags().StringArrayP("env", "e", []string{}, "Env vars to put into generated artifacts")

	generate.AddCommand(dockerfile.cmd)
	dockerfile.cmd.RunE = dockerfile.RunE
}

func (c *GenerateDockerfileCmd) GenerateDockerfile(fs *cs.FileSystem) error {
	if c.Opts.Input == nil || *c.Opts.Input == "" {
		return errors.New("input file is required")
	}
	if c.Opts.Output == nil || *c.Opts.Output == "" {
		return errors.New("output path is required")
	}

	if !fs.FileExists(*c.Opts.Input) {
		if err := c.CloneRepository(fs); err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}
		if !fs.FileExists(*c.Opts.Input) {
			return fmt.Errorf("input file %s not found after cloning repository", *c.Opts.Input)
		}
	}

	envs := append([]string{}, *c.Opts.Env...)
	err := export.ExportDockerArtifacts(fs, *c.Opts.Input, *c.Opts.Output, *c.Opts.BaseImage, envs)
	if err != nil {
		return fmt.Errorf("failed to export docker artifacts: %w", err)
	}

	return nil
}

func (c *GenerateDockerfileCmd) CloneRepository(fs *cs.FileSystem) error {
	client, err := NewClient(c.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	wsId, err := c.Opts.GetWorkspaceId()
	if err != nil {
		return fmt.Errorf("failed to get workspace ID: %w", err)
	}

	ws, err := client.GetWorkspace(wsId)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}
	if ws.GitUrl.Get() == nil {
		return errors.New("workspace does not have a git repository")
	}

	repoUrl := *ws.GitUrl.Get()
	repoBranch := "main"
	if ws.InitialBranch.Get() != nil {
		repoBranch = *ws.InitialBranch.Get()
	}

	_, err = cs.CloneRepository(fs, repoUrl, repoBranch)
	if err != nil {
		return fmt.Errorf("failed to clone repository %s: %w", repoUrl, err)
	}

	fmt.Printf("Repository %s, branch %s cloned.\n", repoUrl, repoBranch)
	return nil
}
