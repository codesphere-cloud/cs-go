// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type ScaleWorkspaceCmd struct {
	cmd  *cobra.Command
	Opts ScaleWorkspaceOpts
}

type ScaleWorkspaceOpts struct {
	*GlobalOptions
	Services []string // each entry is "service=replicas"
}

func (c *ScaleWorkspaceCmd) RunE(_ *cobra.Command, args []string) error {
	workspaceId, err := c.Opts.GetWorkspaceId()
	if err != nil {
		return fmt.Errorf("failed to get workspace ID: %w", err)
	}

	client, err := NewClient(*c.Opts.GlobalOptions)
	if err != nil {
		return fmt.Errorf("failed to create Codesphere client: %w", err)
	}

	return c.ScaleWorkspaceServices(client, workspaceId)
}

func AddScaleWorkspaceCmd(scale *cobra.Command, opts *GlobalOptions) {
	workspace := ScaleWorkspaceCmd{
		cmd: &cobra.Command{
			Use:   "workspace",
			Short: "Scale landscape services of a workspace",
			Long:  io.Long(`Scale landscape services of a workspace by specifying service name and replica count.`),
			Example: io.FormatExampleCommands("scale workspace", []io.Example{
				{Cmd: "--service frontend=2 --service backend=3", Desc: "scale frontend to 2 and backend to 3 replicas"},
				{Cmd: "-w 1234 --service web=1", Desc: "scale web service to 1 replica on workspace 1234"},
				{Cmd: "--service api=0", Desc: "scale api service to 0 replicas"},
			}),
		},
		Opts: ScaleWorkspaceOpts{GlobalOptions: opts},
	}

	workspace.cmd.Flags().StringArrayVar(&workspace.Opts.Services, "service", nil, "Service to scale (format: 'service=replicas'), can be specified multiple times")
	_ = workspace.cmd.MarkFlagRequired("service")

	workspace.cmd.RunE = workspace.RunE

	scale.AddCommand(workspace.cmd)
}

func (c *ScaleWorkspaceCmd) ScaleWorkspaceServices(client Client, wsId int) error {
	services, err := parseScaleServices(c.Opts.Services)
	if err != nil {
		return fmt.Errorf("failed to parse services: %w", err)
	}

	log.Printf("Scaling landscape services for workspace %d: %v\n", wsId, services)
	err = client.ScaleLandscapeServices(wsId, services)
	if err != nil {
		return fmt.Errorf("failed to scale landscape services: %w", err)
	}

	log.Printf("Landscape services scaled for workspace %d\n", wsId)
	return nil
}

// parseScaleServices parses a string slice like ["web=1", "api=2"] into a map[string]int
func parseScaleServices(s []string) (map[string]int, error) {
	result := make(map[string]int)

	for _, pair := range s {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid format '%s', expected 'service=replicas'", pair)
		}
		service := strings.TrimSpace(parts[0])
		if service == "" {
			return nil, fmt.Errorf("empty service name in '%s'", pair)
		}
		replicas, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid replica count '%s' for service '%s': %w", parts[1], service, err)
		}
		if replicas < 0 {
			return nil, fmt.Errorf("replica count must be non-negative for service '%s'", service)
		}
		result[service] = replicas
	}
	return result, nil
}
