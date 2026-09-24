package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliverames/ames-unifi-mcp/internal/client"
	"github.com/oliverames/ames-unifi-mcp/internal/permissions"
)

func BuildWiFiTools(c *client.Client) []*BaseTool {

	return []*BaseTool{
		{
			ToolName: "wifi_broadcast_list", ToolDesc: "List WiFi broadcasts/SSIDs (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatWLAN, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoIntegrationSite(ctx, "GET", "/wifi/broadcasts", nil)
			},
		},
		{
			ToolName: "wifi_broadcast_get", ToolDesc: "Get WiFi broadcast details by ID (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatWLAN, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.DoIntegrationSite(ctx, "GET", fmt.Sprintf("/wifi/broadcasts/%s", p.ID), nil)
			},
		},
		{
			ToolName: "wifi_broadcast_create", ToolDesc: "Create a WiFi broadcast (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatWLAN, ToolAction: permissions.ActionCreate, Mutating: true, MinVer: "9.0.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"WiFi broadcast configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Config json.RawMessage `json:"config"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.DoIntegrationSite(ctx, "POST", "/wifi/broadcasts", p.Config)
			},
		},
		{
			ToolName: "wifi_broadcast_update", ToolDesc: "Update a WiFi broadcast by ID (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatWLAN, ToolAction: permissions.ActionUpdate, Mutating: true, MinVer: "9.0.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string","description":"Broadcast ID"},"config":{"type":"object","description":"Updated broadcast configuration"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.DoIntegrationSite(ctx, "PUT", fmt.Sprintf("/wifi/broadcasts/%s", p.ID), p.Config)
			},
		},
		{
			ToolName: "wifi_broadcast_delete", ToolDesc: "Delete a WiFi broadcast by ID (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatWLAN, ToolAction: permissions.ActionDelete, Mutating: true, MinVer: "9.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.DoIntegrationSite(ctx, "DELETE", fmt.Sprintf("/wifi/broadcasts/%s", p.ID), nil)
			},
		},
	}
}
