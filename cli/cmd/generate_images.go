// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path"

	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/codesphere-cloud/cs-go/pkg/exporter"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type GenerateImagesCmd struct {
	cmd  *cobra.Command
	Opts *GenerateImagesOpts
}

type GenerateImagesOpts struct {
	*GenerateOpts
	Registry    string
	ImagePrefix string
}

func (c *GenerateImagesCmd) RunE(_ *cobra.Command, args []string) error {
	fs := cs.NewOSFileSystem(".")

	exporter := exporter.NewExporterService(fs, c.Opts.Output, "", []string{}, c.Opts.RepoRoot, c.Opts.Force)
	if err := c.GenerateImages(fs, exporter); err != nil {
		return fmt.Errorf("failed to generate images: %w", err)
	}

	log.Println("Images created:")
	log.Printf("Container images from %s pushed to %s\n", c.Opts.Input, c.Opts.Registry)
	log.Println("To generate kubernetes artifacts next, run:")
	log.Printf("%s generate kubernetes --reporoot %s -r %s -p %s -i %s -o %s", io.BinName(), c.Opts.RepoRoot, c.Opts.Registry, c.Opts.ImagePrefix, c.Opts.Input, c.Opts.Output)

	return nil
}

func AddGenerateImagesCmd(generate *cobra.Command, opts *GenerateOpts) {
	images := GenerateImagesCmd{
		cmd: &cobra.Command{
			Use:   "images",
			Short: "Builds and pushes container images from the output folder of the `generate docker` command.",
			Long: io.Long(`The generated images will be pushed to the specified registry.
			As the image name it uses '<registry>/<imagePrefix>-<service-name>:latest'.
			For the nginx router it uses '<registry>/<imagePrefix>-cs-router:latest'.
			If the imagePrefix is not set, it uses '<registry>/<service-name>:latest'.`),
			Example: io.FormatExampleCommands("generate images", []io.Example{
				{Cmd: "-r yourRegistry", Desc: "Generate images and push them to yourRegistry"},
				{Cmd: "-r yourRegistry -p customImagePrefix", Desc: "Build images and push them to yourRegistry with a custom image prefix"},
			}),
		},
		Opts: &GenerateImagesOpts{
			GenerateOpts: opts,
		},
	}
	images.cmd.Flags().StringVarP(&images.Opts.Registry, "registry", "r", "", "Registry to push the resulting images to")
	images.cmd.Flags().StringVarP(&images.Opts.ImagePrefix, "imagePrefix", "p", "", "Image prefix to use for the exported images")

	generate.AddCommand(images.cmd)
	images.cmd.RunE = images.RunE
}

func (c *GenerateImagesCmd) GenerateImages(fs *cs.FileSystem, exp exporter.Exporter) error {
	ciInput := path.Join(c.Opts.RepoRoot, c.Opts.Input)
	if c.Opts.Registry == "" {
		return errors.New("registry is required")
	}

	_, err := exp.ReadYmlFile(ciInput)
	if err != nil {
		return fmt.Errorf("failed to read input file %s artifacts: %w", ciInput, err)
	}

	ctx := context.Background()
	err = exp.ExportImages(ctx, c.Opts.Registry, c.Opts.ImagePrefix)
	if err != nil {
		return fmt.Errorf("failed to export docker artifacts: %w", err)
	}

	return nil
}
