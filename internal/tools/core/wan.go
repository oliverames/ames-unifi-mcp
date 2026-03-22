package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliveames/ames-unifi-mcp/internal/client"
	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
)

func BuildWANTools(c *client.Client) []*BaseTool {
	base := c.Config().BaseURL() + "/integration"

	return []*BaseTool{
		{
			ToolName: "wan_list", ToolDesc: "List WAN interfaces (Network 9.0+)",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/wans", base, c.Site()), nil)
			},
		},
		// --- VPN server CRUD (Integration API 10.1+) ---
		{
			ToolName: "vpn_server_list", ToolDesc: "List VPN server configurations (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatVPN, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/vpn/servers", base, c.Site()), nil)
			},
		},
		{
			ToolName: "vpn_server_get", ToolDesc: "Get a VPN server configuration by ID (Network 9.0+)",
			ToolCategory: permissions.CatVPN, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/vpn/servers/%s", base, c.Site(), p.ID), nil)
			},
		},
		{
			ToolName: "vpn_server_create", ToolDesc: "Create a VPN server (Integration API, Network 10.1+)",
			ToolCategory: permissions.CatVPN, ToolAction: permissions.ActionCreate, Mutating: true, MinVer: "10.1.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"VPN server configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "POST", fmt.Sprintf("%s/v1/sites/%s/vpn/servers", base, c.Site()), p.Config)
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
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "PUT", fmt.Sprintf("%s/v1/sites/%s/vpn/servers/%s", base, c.Site(), p.ID), p.Config)
			},
		},
		{
			ToolName: "vpn_server_delete", ToolDesc: "Delete a VPN server by ID (Integration API, Network 10.1+)",
			ToolCategory: permissions.CatVPN, ToolAction: permissions.ActionDelete, Mutating: true, MinVer: "10.1.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "DELETE", fmt.Sprintf("%s/v1/sites/%s/vpn/servers/%s", base, c.Site(), p.ID), nil)
			},
		},
		// --- Site-to-site VPN tunnel CRUD ---
		{
			ToolName: "vpn_tunnel_list", ToolDesc: "List site-to-site VPN tunnels (Network 9.0+)",
			ToolCategory: permissions.CatVPN, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/vpn/site-to-site-tunnels", base, c.Site()), nil)
			},
		},
		{
			ToolName: "vpn_tunnel_get", ToolDesc: "Get a site-to-site VPN tunnel by ID (Network 9.0+)",
			ToolCategory: permissions.CatVPN, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/vpn/site-to-site-tunnels/%s", base, c.Site(), p.ID), nil)
			},
		},
		{
			ToolName: "vpn_tunnel_create", ToolDesc: "Create a site-to-site VPN tunnel (Integration API, Network 10.1+)",
			ToolCategory: permissions.CatVPN, ToolAction: permissions.ActionCreate, Mutating: true, MinVer: "10.1.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"S2S VPN tunnel configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "POST", fmt.Sprintf("%s/v1/sites/%s/vpn/site-to-site-tunnels", base, c.Site()), p.Config)
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
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "PUT", fmt.Sprintf("%s/v1/sites/%s/vpn/site-to-site-tunnels/%s", base, c.Site(), p.ID), p.Config)
			},
		},
		{
			ToolName: "vpn_tunnel_delete", ToolDesc: "Delete a site-to-site VPN tunnel by ID (Integration API, Network 10.1+)",
			ToolCategory: permissions.CatVPN, ToolAction: permissions.ActionDelete, Mutating: true, MinVer: "10.1.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "DELETE", fmt.Sprintf("%s/v1/sites/%s/vpn/site-to-site-tunnels/%s", base, c.Site(), p.ID), nil)
			},
		},
	}
}
