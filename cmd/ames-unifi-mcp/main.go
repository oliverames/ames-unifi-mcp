package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/oliveames/ames-unifi-mcp/internal/client"
	"github.com/oliveames/ames-unifi-mcp/internal/config"
	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
	"github.com/oliveames/ames-unifi-mcp/internal/tools"
	"github.com/oliveames/ames-unifi-mcp/internal/tools/core"
	"github.com/oliveames/ames-unifi-mcp/internal/tools/extended"
	"github.com/oliveames/ames-unifi-mcp/internal/version"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	c, err := client.New(cfg)
	if err != nil {
		log.Fatalf("client: %v", err)
	}

	// Detect controller version. Skip if unauthenticated — the controller
	// won't respond and the warning would be noise. Tools are still
	// registered so the plugin appears installed; they'll short-circuit
	// with an auth-required error at call time.
	var ver version.Info
	if cfg.NeedsAuth {
		log.Printf("UniFi credentials not configured — starting in needs-auth mode (tools will return configure-me error)")
		ver = version.Info{Raw: "0.0.0"}
	} else {
		ver, err = version.Detect(context.Background(), c)
		if err != nil {
			log.Printf("warning: could not detect controller version: %v (some tools may be unavailable)", err)
			ver = version.Info{Raw: "0.0.0"}
		}
	}

	// Build registry
	permChecker := permissions.NewChecker(cfg.PermissionProfile)
	registry := tools.NewRegistry(permChecker, ver)

	// Register all core tools
	allTools := make([]*core.BaseTool, 0, 200)
	allTools = append(allTools, core.BuildDeviceTools(c)...)
	allTools = append(allTools, core.BuildClientTools(c)...)
	allTools = append(allTools, core.BuildNetworkTools(c)...)
	allTools = append(allTools, core.BuildWLANTools(c)...)
	allTools = append(allTools, core.BuildWiFiTools(c)...)
	allTools = append(allTools, core.BuildFirewallLegacyTools(c)...)
	allTools = append(allTools, core.BuildFirewallZBFTools(c)...)
	allTools = append(allTools, core.BuildACLTools(c)...)
	allTools = append(allTools, core.BuildDNSTools(c)...)
	allTools = append(allTools, core.BuildTrafficTools(c)...)
	allTools = append(allTools, core.BuildWANTools(c)...)
	allTools = append(allTools, core.BuildSwitchingTools(c)...)
	allTools = append(allTools, core.BuildStatsTools(c)...)
	allTools = append(allTools, core.BuildEventTools(c)...)
	allTools = append(allTools, core.BuildSystemTools(c)...)

	// Register extended tools
	allTools = append(allTools, extended.BuildPoETools(c)...)
	allTools = append(allTools, extended.BuildHotspotTools(c)...)
	allTools = append(allTools, extended.BuildCloudTools(c)...)
	allTools = append(allTools, extended.BuildAdminTools(c)...)
	allTools = append(allTools, extended.BuildSyslogTools(c)...)
	allTools = append(allTools, extended.BuildAPGroupTools(c)...)
	allTools = append(allTools, extended.BuildMiscTools(c)...)

	for _, t := range allTools {
		if err := registry.Register(t); err != nil {
			log.Printf("warning: failed to register %s: %v", t.Name(), err)
		}
	}

	// Create MCP server
	s := server.NewMCPServer(
		"ames-unifi-mcp",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	if cfg.ToolMode == config.ToolModeLazy {
		registerLazyTools(s, registry, cfg)
	} else {
		registerEagerTools(s, registry, cfg)
	}

	// Run stdio transport
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

// authGate returns a tool result containing the authentication-required
// hint, used to short-circuit every tool dispatch when cfg.NeedsAuth is set.
func authGate(cfg *config.Config) *mcp.CallToolResult {
	return mcp.NewToolResultError(cfg.AuthHint())
}

func registerLazyTools(s *server.MCPServer, registry *tools.Registry, cfg *config.Config) {
	// tool_index
	indexTool := tools.NewMetaToolIndex(registry)
	s.AddTool(mcp.Tool{
		Name:        indexTool.Name(),
		Description: indexTool.Description(),
		InputSchema: rawToSchema(indexTool.InputSchema()),
	}, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if cfg.NeedsAuth {
			return authGate(cfg), nil
		}
		input, _ := json.Marshal(req.Params.Arguments)
		data, err := indexTool.Execute(ctx, input)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	})

	// tool_execute
	execTool := tools.NewMetaToolExecute(registry)
	s.AddTool(mcp.Tool{
		Name:        execTool.Name(),
		Description: execTool.Description(),
		InputSchema: rawToSchema(execTool.InputSchema()),
	}, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if cfg.NeedsAuth {
			return authGate(cfg), nil
		}
		input, _ := json.Marshal(req.Params.Arguments)
		data, err := execTool.Execute(ctx, input)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	})

	// tool_batch
	batchTool := tools.NewMetaToolBatch(registry)
	s.AddTool(mcp.Tool{
		Name:        batchTool.Name(),
		Description: batchTool.Description(),
		InputSchema: rawToSchema(batchTool.InputSchema()),
	}, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if cfg.NeedsAuth {
			return authGate(cfg), nil
		}
		input, _ := json.Marshal(req.Params.Arguments)
		data, err := batchTool.Execute(ctx, input)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	})
}

func registerEagerTools(s *server.MCPServer, registry *tools.Registry, cfg *config.Config) {
	for _, t := range registry.All() {
		tool := t // capture
		s.AddTool(mcp.Tool{
			Name:        tool.Name(),
			Description: tool.Description(),
			InputSchema: rawToSchema(tool.InputSchema()),
		}, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			if cfg.NeedsAuth {
				return authGate(cfg), nil
			}
			input, _ := json.Marshal(req.Params.Arguments)
			data, err := tool.Execute(ctx, input)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(string(data)), nil
		})
	}
}

// rawToSchema converts a json.RawMessage to the mcp.ToolInputSchema type.
func rawToSchema(raw json.RawMessage) mcp.ToolInputSchema {
	var schema mcp.ToolInputSchema
	if err := json.Unmarshal(raw, &schema); err != nil {
		schema = mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		}
	}
	if schema.Type == "" {
		schema.Type = "object"
	}
	return schema
}
