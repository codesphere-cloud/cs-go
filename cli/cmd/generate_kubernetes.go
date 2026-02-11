// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"fmt"
	"log"
	"path"

	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/codesphere-cloud/cs-go/pkg/exporter"
	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type GenerateKubernetesCmd struct {
	cmd  *cobra.Command
	Opts *GenerateKubernetesOpts
}

type GenerateKubernetesOpts struct {
	*GenerateOpts
	Registry     string
	ImagePrefix  string
	Namespace    string
	PullSecret   string
	Hostname     string
	IngressClass string
}

func (c *GenerateKubernetesCmd) RunE(_ *cobra.Command, args []string) error {
	fs := cs.NewOSFileSystem(".")

	exporter := exporter.NewExporterService(fs, c.Opts.Output, "", []string{}, c.Opts.RepoRoot, c.Opts.Force)
	if err := c.GenerateKubernetes(fs, exporter); err != nil {
		return fmt.Errorf("failed to generate kubernetes: %w", err)
	}

	log.Println("Kubernetes artifacts export successful. You can apply the resources with the following command:")
	log.Printf("kubectl apply -f %s\n", path.Join(c.Opts.RepoRoot, c.Opts.Output, "kubernetes"))
	return nil
}

func AddGenerateKubernetesCmd(generate *cobra.Command, opts *GenerateOpts) {
	kubernetes := GenerateKubernetesCmd{
		cmd: &cobra.Command{
			Use:   "kubernetes",
			Short: "Generates kubernetes artifacts based on a ci.yml of a workspace",
			Long: io.Long(`The generated artifacts will be saved in the output folder (default is ./export).
				In the deployment files the image name is set to '<registry>/<imagePrefix>-<service-name>:latest'.
				The nginx router is set to '<registry>/<imagePrefix>-cs-router:latest' as image name.
				If the imagePrefix is not set, it uses '<registry>/<service-name>:latest'.
				The imagePrefix is used as the namespace for the kubernetes resources, if the prefix is not set, it defaults to 'default'.
				It then generates following artifacts inside the output folder:

				./<service-n> Each service deployment file is exported to a separate folder.
				./<service-n>/<service-n>.yml Kubernetes deployment and service resource to run a pod for the service.
				./ingress.yml ingress resource to route traffic to the different services.

				Codesphere recommends adding the generated artifacts to the source code repository.

				Limitations:
				- Environment variables have to be set explicitly as the Codesphere environment has its own way to provide env variables.
				- The workspace ID, team ID etc. are not automatically available and have to be set explicitly.
				- Hardcoded workspace urls don't work outside of the Codesphere environment.
				- Each dockerfile of your services contain all prepare steps. To have the smallest image possible you would have to delete all unused steps in each service.
				`),
			Example: io.FormatExampleCommands("generate kubernetes", []io.Example{
				{Cmd: "-w 1234", Desc: "Generate kubernetes for workspace 1234"},
				{Cmd: "-w 1234 -i ci.prod.yml", Desc: "Generate kubernetes for workspace 1234 based on ci profile ci.prod.yml"},
			}),
		},
		Opts: &GenerateKubernetesOpts{
			GenerateOpts: opts,
		},
	}
	kubernetes.cmd.Flags().StringVarP(&kubernetes.Opts.Registry, "registry", "r", "", "Registry where images are pushed to (should be the same as used in generate images)")
	kubernetes.cmd.Flags().StringVarP(&kubernetes.Opts.ImagePrefix, "imagePrefix", "p", "", "Image prefix used for the exported images (should be the same as used in generate images)")
	kubernetes.cmd.Flags().StringVarP(&kubernetes.Opts.Namespace, "namespace", "n", "default", "namespace of generated kubernetes artifacts")
	kubernetes.cmd.Flags().StringVar(&kubernetes.Opts.PullSecret, "pullsecret", "", "pullsecret for the pod's images (e.g. for a private registry)")
	kubernetes.cmd.Flags().StringVar(&kubernetes.Opts.Hostname, "hostname", "localhost", "hostname for the ingress to match")
	kubernetes.cmd.Flags().StringVar(&kubernetes.Opts.IngressClass, "ingressClass", "nginx", "ingress class for the ingress resource")

	generate.AddCommand(kubernetes.cmd)
	kubernetes.cmd.RunE = kubernetes.RunE
}

func (c *GenerateKubernetesCmd) GenerateKubernetes(fs *cs.FileSystem, exp exporter.Exporter) error {
	ciInput := path.Join(c.Opts.RepoRoot, c.Opts.Input)
	if c.Opts.Registry == "" {
		return errors.New("registry is required")
	}

	_, err := exp.ReadYmlFile(ciInput)
	if err != nil {
		return fmt.Errorf("failed to read CI definition: %w", err)
	}

	err = exp.ExportKubernetesArtifacts(
		c.Opts.Registry,
		c.Opts.ImagePrefix,
		c.Opts.Namespace,
		c.Opts.PullSecret,
		c.Opts.Hostname,
		c.Opts.IngressClass,
	)
	if err != nil {
		return fmt.Errorf("failed to export kubernetes artifacts: %w", err)
	}

	return nil
}
