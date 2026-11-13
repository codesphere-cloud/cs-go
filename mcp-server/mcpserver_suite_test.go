// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMcpServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MCP Server Suite")
}

// connect registers tools on a fresh server and returns a connected client session.
func connect(register func(*mcp.Server)) *mcp.ClientSession {
	ctx := context.Background()
	server := mcp.NewServer(&mcp.Implementation{Name: "test-server", Version: "v0"}, nil)
	register(server)
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { _ = serverSession.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { _ = session.Close() })
	return session
}

// callTool calls a tool and returns the result.
func callTool(session *mcp.ClientSession, name string, args map[string]any) *mcp.CallToolResult {
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	Expect(err).NotTo(HaveOccurred())
	return res
}

// structured decodes the structured content of a successful tool result.
func structured(res *mcp.CallToolResult) map[string]any {
	Expect(res.IsError).To(BeFalse(), "tool returned an error: %s", text(res))
	raw, err := json.Marshal(res.StructuredContent)
	Expect(err).NotTo(HaveOccurred())
	out := map[string]any{}
	Expect(json.Unmarshal(raw, &out)).To(Succeed())
	return out
}

// text returns the concatenated text content of a tool result.
func text(res *mcp.CallToolResult) string {
	out := ""
	for _, c := range res.Content {
		if t, ok := c.(*mcp.TextContent); ok {
			out += t.Text
		}
	}
	return out
}

// findTool returns the listed tool with the given name.
func findTool(session *mcp.ClientSession, name string) *mcp.Tool {
	tools, err := session.ListTools(context.Background(), nil)
	Expect(err).NotTo(HaveOccurred())
	for _, t := range tools.Tools {
		if t.Name == name {
			return t
		}
	}
	Fail("tool not found: " + name)
	return nil
}

// requiredArgs returns the required properties of a tool's input schema.
func requiredArgs(tool *mcp.Tool) []string {
	raw, err := json.Marshal(tool.InputSchema)
	Expect(err).NotTo(HaveOccurred())
	schema := struct {
		Required []string `json:"required"`
	}{}
	Expect(json.Unmarshal(raw, &schema)).To(Succeed())
	return schema.Required
}
