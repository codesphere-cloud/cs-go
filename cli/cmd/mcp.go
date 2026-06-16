// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	mcpserver "github.com/codesphere-cloud/cs-go/mcp-server"
	csio "github.com/codesphere-cloud/cs-go/pkg/io"
	"github.com/spf13/cobra"
)

type McpCmd struct {
	cmd *cobra.Command
}

func (c *McpCmd) RunE(_ *cobra.Command, args []string) error {
	return mcpserver.Run()
}

func AddMcpCmd(rootCmd *cobra.Command) {
	mcp := McpCmd{
		cmd: &cobra.Command{
			Use:   "mcp",
			Short: "Runs the MCP server for Codesphere",
			Long: csio.Long(`Runs the Model Context Protocol (MCP) server for Codesphere.
			
				The Codesphere MCP (Model Context Protocol) Server allows you to interact with your Codesphere workspaces and teams directly from within MCP-compatible AI assistants.
				Add the Codesphere MCP Server to your MCP client configuration settings.
				Example configuration:
				{
					"mcpServers": {
						"codesphere": {
							"command": "cs",
							"args": [
								"mcp"
							],
							"env": {
								"CS_TOKEN": "your-api-token-here",
								"CS_API": "https://codesphere.com/api"
							}
						}
					}
				}
			`),
		},
	}
	mcp.cmd.RunE = mcp.RunE
	AddCmd(rootCmd, mcp.cmd)
}
