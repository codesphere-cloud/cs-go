## cs mcp

Runs the MCP server for Codesphere

### Synopsis

Runs the Model Context Protocol (MCP) server for Codesphere.

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


```
cs mcp [flags]
```

### Options

```
  -h, --help   help for mcp
```

### Options inherited from parent commands

```
  -a, --api string      URL of Codesphere API (can also be CS_API)
  -O, --org string      Organization ID (relevant for some commands)
  -t, --team int        Team ID (relevant for some commands, can also be CS_TEAM_ID) (default -1)
  -v, --verbose         Verbose output
  -w, --workspace int   Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs](cs.md)	 - The Codesphere CLI

## cs mcp

Runs the MCP server for Codesphere

### Synopsis

Runs the Model Context Protocol (MCP) server for Codesphere.

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


```
cs mcp [flags]
```

### Options

```
  -h, --help   help for mcp
```

### Options inherited from parent commands

```
  -a, --api string      URL of Codesphere API (can also be CS_API)
  -t, --team int        Team ID (relevant for some commands, can also be CS_TEAM_ID) (default -1)
  -v, --verbose         Verbose output
  -w, --workspace int   Workspace ID (relevant for some commands, can also be CS_WORKSPACE_ID) (default -1)
```

### SEE ALSO

* [cs](cs.md)	 - The Codesphere CLI

