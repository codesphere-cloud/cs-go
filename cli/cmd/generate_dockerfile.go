/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// GenerateDockerfileCmd represents the dockerfile command
type GenerateDockerfileCmd struct {
	cmd *cobra.Command
}

func (c *GenerateDockerfileCmd) RunE(_ *cobra.Command, args []string) error {

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
	dockerfile.cmd.Flags().StringP("input", "i", "ci.yml", "CI profile to use as input for generation")
	dockerfile.cmd.Flags().StringP("baseimage", "b", "", "Base image for the dockerfile")
	dockerfile.cmd.Flags().StringP("output", "o", "./export", "Output path of the folder including generated artifacts")
	dockerfile.cmd.Flags().StringArrayP("env", "e", []string{}, "Env vars to put into generated artifacts")

	generate.AddCommand(dockerfile.cmd)
	dockerfile.cmd.RunE = dockerfile.RunE
}
