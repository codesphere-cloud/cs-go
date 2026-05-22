# Codesphere MCP Server

The Codesphere MCP (Model Context Protocol) Server allows you to interact with your Codesphere workspaces and teams directly from within MCP-compatible AI assistants.

## Configuration

Add the Codesphere MCP Server to your MCP client configuration settings.

### VS Code (Cline / Roo Code)

If you're using VS Code extensions like Cline or Roo Code, open your MCP settings (`cline_mcp_settings.json`) and add the server configuration:

```json
{
  "mcpServers": {
    "codesphere": {
      "command": "/path/to/cs-mcp",
      "env": {
        "CS_TOKEN": "your-api-token-here",
        "CS_API": "https://codesphere.com/api"
      }
    }
  }
}
```

### Claude Desktop

For Claude Desktop, add the same configuration to your `claude_desktop_config.json` file:

```json
{
  "mcpServers": {
    "codesphere": {
      "command": "/path/to/cs-mcp",
      "env": {
        "CS_TOKEN": "your-api-token-here",
        "CS_API": "https://codesphere.com/api"
      }
    }
  }
}
```

> **Note:** Replace `/path/to/cs-mcp` with the actual, absolute path to the compiled `cs-mcp` executable on your system.

## Getting Started

1. **Build or Download**: Obtain the `cs-mcp` binary (for example, by building it using `go build` inside this project).
2. **API Token**: Get your `CS_TOKEN` from your Codesphere account under your workspace settings.
3. **Configure**: Add the JSON snippet above to your client's MCP configuration file and restart the client.
