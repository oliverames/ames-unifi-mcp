package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliveames/ames-unifi-mcp/internal/client"
	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
)

func BuildNetworkTools(c *client.Client) []*BaseTool {
	sp := func() string { return fmt.Sprintf("api/s/%s", c.Site()) }
	base := c.Config().BaseURL() + "/integration"

	return []*BaseTool{
		// --- Integration API network tools (9.0+) ---
		{
			ToolName: "network_list_v2", ToolDesc: "List all networks (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/networks", base, c.Site()), nil)
			},
		},
		{
			ToolName: "network_get_v2", ToolDesc: "Get network details by ID (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/networks/%s", base, c.Site(), p.ID), nil)
			},
		},
		{
			ToolName: "network_create_v2", ToolDesc: "Create a network (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionCreate, Mutating: true, MinVer: "9.0.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"Network configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "POST", fmt.Sprintf("%s/v1/sites/%s/networks", base, c.Site()), p.Config)
			},
		},
		{
			ToolName: "network_update_v2", ToolDesc: "Update a network by ID (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionUpdate, Mutating: true, MinVer: "9.0.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string","description":"Network ID"},"config":{"type":"object","description":"Updated network configuration"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "PUT", fmt.Sprintf("%s/v1/sites/%s/networks/%s", base, c.Site(), p.ID), p.Config)
			},
		},
		{
			ToolName: "network_delete_v2", ToolDesc: "Delete a network by ID (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionDelete, Mutating: true, MinVer: "9.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "DELETE", fmt.Sprintf("%s/v1/sites/%s/networks/%s", base, c.Site(), p.ID), nil)
			},
		},
		// --- Legacy API ---
		{
			ToolName: "network_list", ToolDesc: "List all configured networks (VLANs, LANs, WANs)",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/networkconf", nil)
			},
		},
		{
			ToolName: "network_get", ToolDesc: "Get details for a specific network by ID",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionRead,
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "GET", sp()+"/rest/networkconf/"+p.ID, nil)
			},
		},
		{
			ToolName: "network_create", ToolDesc: "Create a new network",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionCreate, Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"config": {"type": "object", "description": "Network configuration (name, purpose, vlan, subnet, dhcp settings, etc.)"}
				},
				"required": ["config"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/rest/networkconf", p.Config)
			},
		},
		{
			ToolName: "network_update", ToolDesc: "Update an existing network by ID",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"id": {"type": "string", "description": "Network _id"},
					"config": {"type": "object", "description": "Updated network configuration"}
				},
				"required": ["id", "config"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "PUT", sp()+"/rest/networkconf/"+p.ID, p.Config)
			},
		},
		{
			ToolName: "network_delete", ToolDesc: "Delete a network by ID",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionDelete, Mutating: true,
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "DELETE", sp()+"/rest/networkconf/"+p.ID, nil)
			},
		},
	}
}
