package extended

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliverames/ames-unifi-mcp/internal/client"
	"github.com/oliverames/ames-unifi-mcp/internal/permissions"
	"github.com/oliverames/ames-unifi-mcp/internal/tools/core"
)

func BuildHotspotTools(c *client.Client) []*core.BaseTool {
	sp := func() string { return fmt.Sprintf("api/s/%s", c.Site()) }

	return []*core.BaseTool{
		{
			ToolName: "hotspot_authorize_guest", ToolDesc: "Authorize a guest client with optional time/bandwidth limits",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionExecute, Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"mac": {"type": "string", "description": "Client MAC address"},
					"minutes": {"type": "integer", "description": "Authorization duration in minutes"},
					"up": {"type": "integer", "description": "Upload speed limit (Kbps)"},
					"down": {"type": "integer", "description": "Download speed limit (Kbps)"},
					"bytes": {"type": "integer", "description": "Data transfer limit (bytes)"},
					"ap_mac": {"type": "string", "description": "AP MAC to authorize on (optional)"}
				},
				"required": ["mac", "minutes"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Mac     string `json:"mac"`
					Minutes int    `json:"minutes"`
					Up      int    `json:"up,omitempty"`
					Down    int    `json:"down,omitempty"`
					Bytes   int    `json:"bytes,omitempty"`
					APMac   string `json:"ap_mac,omitempty"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				payload := map[string]interface{}{"cmd": "authorize-guest", "mac": p.Mac, "minutes": p.Minutes}
				if p.Up > 0 {
					payload["up"] = p.Up
				}
				if p.Down > 0 {
					payload["down"] = p.Down
				}
				if p.Bytes > 0 {
					payload["bytes"] = p.Bytes
				}
				if p.APMac != "" {
					payload["ap_mac"] = p.APMac
				}
				return c.Do(ctx, "POST", sp()+"/cmd/stamgr", payload)
			},
		},
		{
			ToolName: "hotspot_unauthorize_guest", ToolDesc: "Revoke guest authorization",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionExecute, Mutating: true,
			Schema: core.MacSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Mac string `json:"mac"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.Do(ctx, "POST", sp()+"/cmd/stamgr", map[string]interface{}{"cmd": "unauthorize-guest", "mac": p.Mac})
			},
		},
		{
			ToolName: "hotspot_list_guests", ToolDesc: "List active guest sessions",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionRead,
			Schema: json.RawMessage(`{"type":"object","properties":{"within":{"type":"integer","description":"Hours of history (default 24)","default":24}}}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Within int `json:"within"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				if p.Within == 0 {
					p.Within = 24
				}
				return c.Do(ctx, "POST", sp()+"/stat/guest", map[string]interface{}{"within": p.Within})
			},
		},
		{
			ToolName: "hotspot_create_voucher", ToolDesc: "Create guest vouchers for hotspot access",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionCreate, Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"count": {"type": "integer", "description": "Number of vouchers to create", "default": 1},
					"expire_minutes": {"type": "integer", "description": "Validity in minutes (e.g. 1440 = 24h)"},
					"quota": {"type": "integer", "description": "0=multi-use, 1=single-use, n=n-use", "default": 1},
					"note": {"type": "string", "description": "Optional note"},
					"up": {"type": "integer", "description": "Upload limit (Kbps)"},
					"down": {"type": "integer", "description": "Download limit (Kbps)"},
					"bytes": {"type": "integer", "description": "Data limit (megabytes)"}
				},
				"required": ["expire_minutes"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Count         int    `json:"count"`
					ExpireMinutes int    `json:"expire_minutes"`
					Quota         *int   `json:"quota"`
					Note          string `json:"note,omitempty"`
					Up            int    `json:"up,omitempty"`
					Down          int    `json:"down,omitempty"`
					Bytes         int    `json:"bytes,omitempty"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				if p.Count == 0 {
					p.Count = 1
				}
				quota := 1
				if p.Quota != nil {
					quota = *p.Quota
				}
				payload := map[string]interface{}{
					"cmd": "create-voucher", "n": p.Count, "expire": p.ExpireMinutes, "quota": quota,
				}
				if p.Note != "" {
					payload["note"] = p.Note
				}
				if p.Up > 0 {
					payload["up"] = p.Up
				}
				if p.Down > 0 {
					payload["down"] = p.Down
				}
				if p.Bytes > 0 {
					payload["bytes"] = p.Bytes
				}
				return c.Do(ctx, "POST", sp()+"/cmd/hotspot", payload)
			},
		},
		{
			ToolName: "hotspot_voucher_delete", ToolDesc: "Delete/revoke a voucher",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionDelete, Mutating: true,
			Schema: core.IDSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.Do(ctx, "POST", sp()+"/cmd/hotspot", map[string]interface{}{
					"cmd": "delete-voucher", "_id": p.ID,
				})
			},
		},
		{
			ToolName: "hotspot_extend", ToolDesc: "Extend a guest's authorization",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: core.IDSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.Do(ctx, "POST", sp()+"/cmd/hotspot", map[string]interface{}{
					"cmd": "extend", "_id": p.ID,
				})
			},
		},
		// --- Integration API hotspot voucher tools (9.0+) ---
		{
			ToolName: "hotspot_voucher_list_v2", ToolDesc: "List hotspot vouchers (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoIntegrationSite(ctx, "GET", "/hotspot/vouchers", nil)
			},
		},
		{
			ToolName: "hotspot_voucher_get_v2", ToolDesc: "Get voucher details by ID (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: core.IDSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.DoIntegrationSite(ctx, "GET", fmt.Sprintf("/hotspot/vouchers/%s", p.ID), nil)
			},
		},
		{
			ToolName: "hotspot_voucher_create_v2", ToolDesc: "Generate vouchers (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionCreate, Mutating: true, MinVer: "9.0.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"Voucher generation config"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Config json.RawMessage `json:"config"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.DoIntegrationSite(ctx, "POST", "/hotspot/vouchers", p.Config)
			},
		},
		{
			ToolName: "hotspot_voucher_delete_v2", ToolDesc: "Delete a voucher by ID (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionDelete, Mutating: true, MinVer: "9.0.0",
			Schema: core.IDSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.DoIntegrationSite(ctx, "DELETE", fmt.Sprintf("/hotspot/vouchers/%s", p.ID), nil)
			},
		},
		{
			ToolName: "hotspot_voucher_bulk_delete", ToolDesc: "Delete all vouchers (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionDelete, Mutating: true, MinVer: "9.0.0",
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoIntegrationSite(ctx, "DELETE", "/hotspot/vouchers", nil)
			},
		},
		{
			ToolName: "hotspot_config", ToolDesc: "Get hotspot portal configuration (auth type, design)",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", fmt.Sprintf("guest/s/%s/hotspotconfig", c.Site()), nil)
			},
		},
		{
			ToolName: "hotspot_packages", ToolDesc: "List hotspot billing packages",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", fmt.Sprintf("guest/s/%s/hotspotpackages", c.Site()), nil)
			},
		},
	}
}
