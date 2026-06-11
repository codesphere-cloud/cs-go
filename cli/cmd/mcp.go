// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	mcpserver "github.com/codesphere-cloud/cs-go/mcp-server"
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
			Short: "Run the MCP server",
			Long:  `Run the Model Context Protocol (MCP) server for Codesphere.`,
		},
	}
	mcp.cmd.RunE = mcp.RunE
	AddCmd(rootCmd, mcp.cmd)
}
