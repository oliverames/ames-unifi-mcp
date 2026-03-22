package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliveames/ames-unifi-mcp/internal/client"
	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
)

func BuildWLANTools(c *client.Client) []*BaseTool {
	sp := func() string { return fmt.Sprintf("api/s/%s", c.Site()) }

	return []*BaseTool{
		{
			ToolName: "wlan_list", ToolDesc: "List all configured wireless networks (SSIDs)",
			ToolCategory: permissions.CatWLAN, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/wlanconf", nil)
			},
		},
		{
			ToolName: "wlan_get", ToolDesc: "Get details for a specific WLAN by ID",
			ToolCategory: permissions.CatWLAN, ToolAction: permissions.ActionRead,
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "GET", sp()+"/rest/wlanconf/"+p.ID, nil)
			},
		},
		{
			ToolName: "wlan_create", ToolDesc: "Create a new wireless network",
			ToolCategory: permissions.CatWLAN, ToolAction: permissions.ActionCreate, Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"config": {"type": "object", "description": "WLAN config (name, security, passphrase, etc.)"}
				},
				"required": ["config"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/rest/wlanconf", p.Config)
			},
		},
		{
			ToolName: "wlan_update", ToolDesc: "Update a WLAN configuration by ID",
			ToolCategory: permissions.CatWLAN, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"id": {"type": "string", "description": "WLAN _id"},
					"config": {"type": "object", "description": "Updated WLAN configuration"}
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
				return c.Do(ctx, "PUT", sp()+"/rest/wlanconf/"+p.ID, p.Config)
			},
		},
		{
			ToolName: "wlan_delete", ToolDesc: "Delete a WLAN by ID",
			ToolCategory: permissions.CatWLAN, ToolAction: permissions.ActionDelete, Mutating: true,
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "DELETE", sp()+"/rest/wlanconf/"+p.ID, nil)
			},
		},
		{
			ToolName: "wlan_enable", ToolDesc: "Enable a wireless network (turn SSID on)",
			ToolCategory: permissions.CatWLAN, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "PUT", sp()+"/rest/wlanconf/"+p.ID, map[string]interface{}{"enabled": true})
			},
		},
		{
			ToolName: "wlan_disable", ToolDesc: "Disable a wireless network (turn SSID off)",
			ToolCategory: permissions.CatWLAN, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "PUT", sp()+"/rest/wlanconf/"+p.ID, map[string]interface{}{"enabled": false})
			},
		},
	}
}
