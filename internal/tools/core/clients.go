package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliveames/ames-unifi-mcp/internal/client"
	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
)

func BuildClientTools(c *client.Client) []*BaseTool {
	sp := func() string { return fmt.Sprintf("api/s/%s", c.Site()) }

	return []*BaseTool{
		{
			ToolName: "client_list_active", ToolDesc: "List all currently connected clients",
			ToolCategory: permissions.CatClients, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/stat/sta", nil)
			},
		},
		{
			ToolName: "client_list_all", ToolDesc: "List all known clients (including historical/offline)",
			ToolCategory: permissions.CatClients, ToolAction: permissions.ActionRead,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"within": {"type": "integer", "description": "Hours of history (default: 8760)", "default": 8760}
				}
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Within int `json:"within"` }
				json.Unmarshal(input, &p)
				if p.Within == 0 {
					p.Within = 8760
				}
				return c.Do(ctx, "POST", sp()+"/stat/alluser", map[string]interface{}{
					"type": "all", "conn": "all", "within": p.Within,
				})
			},
		},
		{
			ToolName: "client_get", ToolDesc: "Get details for a specific client by MAC address",
			ToolCategory: permissions.CatClients, ToolAction: permissions.ActionRead,
			Schema: macSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Mac string `json:"mac"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "GET", sp()+"/stat/sta/"+p.Mac, nil)
			},
		},
		{
			ToolName: "client_block", ToolDesc: "Block a client device by MAC address",
			ToolCategory: permissions.CatClients, ToolAction: permissions.ActionExecute, Mutating: true,
			Schema: macSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Mac string `json:"mac"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/cmd/stamgr", map[string]interface{}{"cmd": "block-sta", "mac": p.Mac})
			},
		},
		{
			ToolName: "client_unblock", ToolDesc: "Unblock a client device by MAC address",
			ToolCategory: permissions.CatClients, ToolAction: permissions.ActionExecute, Mutating: true,
			Schema: macSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Mac string `json:"mac"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/cmd/stamgr", map[string]interface{}{"cmd": "unblock-sta", "mac": p.Mac})
			},
		},
		{
			ToolName: "client_reconnect", ToolDesc: "Disconnect and force a client to reconnect",
			ToolCategory: permissions.CatClients, ToolAction: permissions.ActionExecute, Mutating: true,
			Schema: macSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Mac string `json:"mac"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/cmd/stamgr", map[string]interface{}{"cmd": "kick-sta", "mac": p.Mac})
			},
		},
		{
			ToolName: "client_forget", ToolDesc: "Forget a client permanently (remove from known clients). Can be slow.",
			ToolCategory: permissions.CatClients, ToolAction: permissions.ActionDelete, Mutating: true, MinVer: "5.9.0",
			Schema: macSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Mac string `json:"mac"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/cmd/stamgr", map[string]interface{}{"cmd": "forget-sta", "macs": []string{p.Mac}})
			},
		},
		{
			ToolName: "client_list_configured", ToolDesc: "List all configured/known clients via REST (includes static IPs, names, groups)",
			ToolCategory: permissions.CatClients, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/user", nil)
			},
		},
		{
			ToolName: "client_get_configured", ToolDesc: "Get configured/known client details by MAC (includes historical data)",
			ToolCategory: permissions.CatClients, ToolAction: permissions.ActionRead,
			Schema: macSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Mac string `json:"mac"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "GET", sp()+"/stat/user/"+p.Mac, nil)
			},
		},
		{
			ToolName: "client_rename", ToolDesc: "Rename a client device",
			ToolCategory: permissions.CatClients, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"id": {"type": "string", "description": "Client _id"},
					"name": {"type": "string", "description": "New name"}
				},
				"required": ["id", "name"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/upd/user/"+p.ID, map[string]interface{}{"name": p.Name})
			},
		},
		{
			ToolName: "client_update", ToolDesc: "Update client settings (fixed IP, user group, note, etc.)",
			ToolCategory: permissions.CatClients, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"id": {"type": "string", "description": "Client _id"},
					"config": {"type": "object", "description": "Client configuration object"}
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
				return c.Do(ctx, "PUT", sp()+"/rest/user/"+p.ID, p.Config)
			},
		},
		// --- Integration API tools (9.0+) ---
		{
			ToolName: "client_list_v2", ToolDesc: "List connected clients (Integration API, structured response)",
			ToolCategory: permissions.CatClients, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				base := c.Config().BaseURL() + "/integration"
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/clients", base, c.Site()), nil)
			},
		},
		{
			ToolName: "client_get_v2", ToolDesc: "Get client details by client ID (Integration API)",
			ToolCategory: permissions.CatClients, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"client_id": {"type": "string", "description": "Client ID"}
				},
				"required": ["client_id"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ClientID string `json:"client_id"` }
				json.Unmarshal(input, &p)
				base := c.Config().BaseURL() + "/integration"
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/clients/%s", base, c.Site(), p.ClientID), nil)
			},
		},
		{
			ToolName: "client_action", ToolDesc: "Execute a client action (Integration API)",
			ToolCategory: permissions.CatClients, ToolAction: permissions.ActionExecute, Mutating: true, MinVer: "9.0.0",
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"client_id": {"type": "string", "description": "Client ID"},
					"config": {"type": "object", "description": "Action payload"}
				},
				"required": ["client_id", "config"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ClientID string          `json:"client_id"`
					Config   json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				base := c.Config().BaseURL() + "/integration"
				return c.DoRaw(ctx, "POST", fmt.Sprintf("%s/v1/sites/%s/clients/%s/actions", base, c.Site(), p.ClientID), p.Config)
			},
		},
		{
			ToolName: "client_sessions", ToolDesc: "List client login sessions (optionally filtered by MAC, time range, and type)",
			ToolCategory: permissions.CatClients, ToolAction: permissions.ActionRead,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"mac": {"type": "string", "description": "Client MAC address (optional, filters to one client)"},
					"start": {"type": "integer", "description": "Start time (unix timestamp seconds)"},
					"end": {"type": "integer", "description": "End time (unix timestamp seconds)"},
					"type": {"type": "string", "description": "Session type (default: all)", "default": "all"}
				}
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Mac   string `json:"mac"`
					Start int64  `json:"start"`
					End   int64  `json:"end"`
					Type  string `json:"type"`
				}
				json.Unmarshal(input, &p)
				if p.Type == "" {
					p.Type = "all"
				}
				body := map[string]interface{}{"type": p.Type}
				if p.Mac != "" {
					body["mac"] = p.Mac
				}
				if p.Start != 0 {
					body["start"] = p.Start
				}
				if p.End != 0 {
					body["end"] = p.End
				}
				return c.Do(ctx, "POST", sp()+"/stat/session", body)
			},
		},
	}
}
