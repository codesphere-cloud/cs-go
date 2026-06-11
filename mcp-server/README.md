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
```

### Claude Desktop

For Claude Desktop, add the same configuration to your `claude_desktop_config.json` file:

```json
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
```

> **Note:** If the `cs` CLI is not in your system's `PATH`, replace `"command": "cs"` with the absolute path to the compiled `cs` executable on your system (e.g. `"/Users/youruser/bin/cs"`).

## Getting Started

1. **Build or Download**: Obtain the `cs` binary (for example, by building it using `go build -o cs ./cli` in the root of the project).
2. **API Token**: Get your `CS_TOKEN` from your Codesphere account under your workspace settings.
3. **API URL**: Get the correct `CS_API` for the codesphere instance you want to reach.
4. **Configure**: Add the JSON snippet above to your client's MCP configuration file and restart the client.
