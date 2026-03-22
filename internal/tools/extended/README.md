# Adding New Tools

This directory is the extension layer for `ames-unifi-mcp`. Adding a new tool should be trivial.

## Steps

### 1. Document the endpoint first

Update `docs/api-research.md` with the endpoint path, HTTP method, payload shape, source, and confidence level. This is a hard requirement — no tool without documentation.

### 2. Create a new `.go` file

Create a file in this directory (e.g., `backup.go`). Follow the pattern in `poe.go` or `hotspot.go`.

### 3. Implement the tool

Every tool is a `*core.BaseTool` struct with these fields:

```go
&core.BaseTool{
    ToolName:     "category_action",           // e.g., "poe_power_cycle"
    ToolDesc:     "Human-readable description", // shown to the LLM
    ToolCategory: permissions.CatPoE,           // permission category
    ToolAction:   permissions.ActionExecute,     // read, create, update, delete, execute
    Mutating:     true,                          // triggers confirm gate if true
    MinVer:       "9.0.0",                       // minimum controller version (optional)
    Undocumented: true,                          // adds [undocumented] disclaimer
    Schema:       json.RawMessage(`{...}`),      // JSON Schema for input
    Client:       c,                             // the HTTP client
    Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
        // Your implementation here
    },
}
```

### 4. Return tools from a builder function

```go
func BuildMyTools(c *client.Client) []*core.BaseTool {
    return []*core.BaseTool{
        // your tools
    }
}
```

### 5. Register in main.go

Add one line to `cmd/ames-unifi-mcp/main.go`:

```go
allTools = append(allTools, extended.BuildMyTools(c)...)
```

## Key patterns

- **Mutating tools** get wrapped with the confirm gate automatically. The LLM must pass `confirm: true` to execute.
- **Version-gated tools**: Set `MinVer` and the registry silently skips the tool if the controller is too old.
- **Undocumented endpoints**: Set `Undocumented: true`. The description will include a disclaimer.
- **Use `c.Do()`** for legacy API endpoints (auto-parses the `{"meta":...,"data":...}` envelope).
- **Use `c.DoRaw()`** for Integration API or v2 endpoints that return different response formats.
- **Site path**: Use `fmt.Sprintf("api/s/%s", c.Site())` for the legacy site prefix.
- **Integration API base**: Use `c.Config().BaseURL() + "/integration"`.

## Helper schemas

Import from `core` package:
- `core.NoInputSchema()` — no parameters needed
- `core.MacSchema()` — requires a MAC address
- `core.IDSchema()` — requires a resource `_id`
