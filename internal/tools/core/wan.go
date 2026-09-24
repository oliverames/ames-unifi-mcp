package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliverames/ames-unifi-mcp/internal/client"
	"github.com/oliverames/ames-unifi-mcp/internal/permissions"
)

func BuildWANTools(c *client.Client) []*BaseTool {

	return []*BaseTool{
		{
			ToolName: "wan_list", ToolDesc: "List WAN interfaces (Network 9.0+)",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoIntegrationSite(ctx, "GET", "/wans", nil)
			},
		},
		// --- VPN server CRUD (Integration API 10.1+) ---
		{
			ToolName: "vpn_server_list", ToolDesc: "List VPN server configurations (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatVPN, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoIntegrationSite(ctx, "GET", "/vpn/servers", nil)
			},
		},
		{
			ToolName: "vpn_server_get", ToolDesc: "Get a VPN server configuration by ID (Network 9.0+)",
			ToolCategory: permissions.CatVPN, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.DoIntegrationSite(ctx, "GET", fmt.Sprintf("/vpn/servers/%s", p.ID), nil)
			},
		},
		{
			ToolName: "vpn_server_create", ToolDesc: "Create a VPN server (Integration API, Network 10.1+)",
			ToolCategory: permissions.CatVPN, ToolAction: permissions.ActionCreate, Mutating: true, MinVer: "10.1.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"VPN server configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Config json.RawMessage `json:"config"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.DoIntegrationSite(ctx, "POST", "/vpn/servers", p.Config)
			},
		},
		{
			ToolName: "vpn_server_update", ToolDesc: "Update a VPN server by ID (Integration API, Network 10.1+)",
			ToolCategory: permissions.CatVPN, ToolAction: permissions.ActionUpdate, Mutating: true, MinVer: "10.1.0",
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
				return c.DoIntegrationSite(ctx, "PUT", fmt.Sprintf("/vpn/servers/%s", p.ID), p.Config)
			},
		},
		{
			ToolName: "vpn_server_delete", ToolDesc: "Delete a VPN server by ID (Integration API, Network 10.1+)",
			ToolCategory: permissions.CatVPN, ToolAction: permissions.ActionDelete, Mutating: true, MinVer: "10.1.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.DoIntegrationSite(ctx, "DELETE", fmt.Sprintf("/vpn/servers/%s", p.ID), nil)
			},
		},
		// --- Site-to-site VPN tunnel CRUD ---
		{
			ToolName: "vpn_tunnel_list", ToolDesc: "List site-to-site VPN tunnels (Network 9.0+)",
			ToolCategory: permissions.CatVPN, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoIntegrationSite(ctx, "GET", "/vpn/site-to-site-tunnels", nil)
			},
		},
		{
			ToolName: "vpn_tunnel_get", ToolDesc: "Get a site-to-site VPN tunnel by ID (Network 9.0+)",
			ToolCategory: permissions.CatVPN, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.DoIntegrationSite(ctx, "GET", fmt.Sprintf("/vpn/site-to-site-tunnels/%s", p.ID), nil)
			},
		},
		{
			ToolName: "vpn_tunnel_create", ToolDesc: "Create a site-to-site VPN tunnel (Integration API, Network 10.1+)",
			ToolCategory: permissions.CatVPN, ToolAction: permissions.ActionCreate, Mutating: true, MinVer: "10.1.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"S2S VPN tunnel configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Config json.RawMessage `json:"config"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.DoIntegrationSite(ctx, "POST", "/vpn/site-to-site-tunnels", p.Config)
			},
		},
		{
			ToolName: "vpn_tunnel_update", ToolDesc: "Update a site-to-site VPN tunnel by ID (Integration API, Network 10.1+)",
			ToolCategory: permissions.CatVPN, ToolAction: permissions.ActionUpdate, Mutating: true, MinVer: "10.1.0",
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
				return c.DoIntegrationSite(ctx, "PUT", fmt.Sprintf("/vpn/site-to-site-tunnels/%s", p.ID), p.Config)
			},
		},
		{
			ToolName: "vpn_tunnel_delete", ToolDesc: "Delete a site-to-site VPN tunnel by ID (Integration API, Network 10.1+)",
			ToolCategory: permissions.CatVPN, ToolAction: permissions.ActionDelete, Mutating: true, MinVer: "10.1.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.DoIntegrationSite(ctx, "DELETE", fmt.Sprintf("/vpn/site-to-site-tunnels/%s", p.ID), nil)
			},
		},
	}
}
