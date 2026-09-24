package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliverames/ames-unifi-mcp/internal/client"
	"github.com/oliverames/ames-unifi-mcp/internal/permissions"
)

func BuildDNSTools(c *client.Client) []*BaseTool {

	return []*BaseTool{
		{
			ToolName: "dns_policy_list", ToolDesc: "List all DNS policies (Network 10.0+)",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionRead, MinVer: "10.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoIntegrationSite(ctx, "GET", "/dns/policies", nil)
			},
		},
		{
			ToolName: "dns_policy_create", ToolDesc: "Create a new DNS policy (Network 10.0+)",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionCreate, Mutating: true, MinVer: "10.0.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"DNS policy configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Config json.RawMessage `json:"config"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.DoIntegrationSite(ctx, "POST", "/dns/policies", p.Config)
			},
		},
		{
			ToolName: "dns_policy_get", ToolDesc: "Get a DNS policy by ID (Network 10.0+)",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionRead, MinVer: "10.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.DoIntegrationSite(ctx, "GET", fmt.Sprintf("/dns/policies/%s", p.ID), nil)
			},
		},
		{
			ToolName: "dns_policy_update", ToolDesc: "Update a DNS policy by ID (Network 10.0+)",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionUpdate, Mutating: true, MinVer: "10.0.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"},"config":{"type":"object"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.DoIntegrationSite(ctx, "PUT", fmt.Sprintf("/dns/policies/%s", p.ID), p.Config)
			},
		},
		{
			ToolName: "dns_policy_delete", ToolDesc: "Delete a DNS policy by ID (Network 10.0+)",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionDelete, Mutating: true, MinVer: "10.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.DoIntegrationSite(ctx, "DELETE", fmt.Sprintf("/dns/policies/%s", p.ID), nil)
			},
		},
	}
}
