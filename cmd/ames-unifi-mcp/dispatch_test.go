package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/server"
	"github.com/oliverames/ames-unifi-mcp/internal/config"
	"github.com/oliverames/ames-unifi-mcp/internal/permissions"
	"github.com/oliverames/ames-unifi-mcp/internal/tools"
	"github.com/oliverames/ames-unifi-mcp/internal/tools/core"
	"github.com/oliverames/ames-unifi-mcp/internal/version"
)

func TestMCPDispatchOmittedArguments(t *testing.T) {
	for _, mode := range []string{"lazy", "eager"} {
		t.Run(mode, func(t *testing.T) {
			calls := 0
			registry := tools.NewRegistry(permissions.NewChecker(config.PermAdmin), version.Info{Major: 10})
			for name, schema := range map[string]string{"no_input": `{"type":"object","properties":{}}`, "required_input": `{"type":"object","properties":{"id":{"type":"string"}},"required":["id"]}`} {
				err := registry.Register(&core.BaseTool{ToolName: name, ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionRead, Schema: json.RawMessage(schema), Handler: func(context.Context, json.RawMessage) (json.RawMessage, error) {
					calls++
					return json.RawMessage(`{"ok":true}`), nil
				}})
				if err != nil {
					t.Fatal(err)
				}
			}
			srv := server.NewMCPServer("test", "1")
			if mode == "lazy" {
				registerLazyTools(srv, registry, &config.Config{})
			} else {
				registerEagerTools(srv, registry, &config.Config{})
			}
			name := "no_input"
			if mode == "lazy" {
				name = "tool_index"
			}
			for _, arguments := range []string{"", `,"arguments":{}`, `,"arguments":null`, `,"arguments":[]`, `,"arguments":"bad"`} {
				response := srv.HandleMessage(context.Background(), []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"`+name+`"`+arguments+`}}`))
				wire, err := json.Marshal(response)
				if err != nil {
					t.Fatal(err)
				}
				var decoded struct {
					Error  json.RawMessage `json:"error"`
					Result struct {
						IsError bool `json:"isError"`
					} `json:"result"`
				}
				if err := json.Unmarshal(wire, &decoded); err != nil {
					t.Fatal(err)
				}
				failed := len(decoded.Error) > 0 || decoded.Result.IsError
				wantFailure := arguments != "" && arguments != `,"arguments":{}`
				if failed != wantFailure {
					t.Errorf("arguments %s: response %s", arguments, wire)
				}
			}
			requiredName := "required_input"
			if mode == "lazy" {
				requiredName = "tool_execute"
			}
			response := srv.HandleMessage(context.Background(), []byte(`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"`+requiredName+`"}}`))
			wire, _ := json.Marshal(response)
			var decoded struct {
				Result struct {
					IsError bool `json:"isError"`
				} `json:"result"`
			}
			json.Unmarshal(wire, &decoded)
			if !decoded.Result.IsError {
				t.Errorf("missing required accepted: %s", wire)
			}
			wantCalls := 2
			if mode == "lazy" {
				wantCalls = 0
			}
			if calls != wantCalls {
				t.Errorf("handler calls = %d, want %d", calls, wantCalls)
			}
		})
	}
}
