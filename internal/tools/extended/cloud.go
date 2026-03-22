package extended

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/oliveames/ames-unifi-mcp/internal/client"
	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
	"github.com/oliveames/ames-unifi-mcp/internal/tools/core"
)

func BuildCloudTools(c *client.Client) []*core.BaseTool {
	baseURL := "https://api.ui.com"

	paginatedSchema := func(extra string) json.RawMessage {
		props := `"pageSize":{"type":"string","description":"Number of results per page"},"nextToken":{"type":"string","description":"Pagination token from previous response"}`
		if extra != "" {
			props = extra + "," + props
		}
		return json.RawMessage(fmt.Sprintf(`{"type":"object","properties":{%s}}`, props))
	}

	return []*core.BaseTool{
		{
			ToolName: "cloud_host_list", ToolDesc: "List all UniFi hosts/consoles linked to your account (Cloud API)",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionRead,
			Schema: paginatedSchema(""), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					PageSize  string `json:"pageSize"`
					NextToken string `json:"nextToken"`
				}
				json.Unmarshal(input, &p)
				u := baseURL + "/v1/hosts"
				sep := "?"
				if p.PageSize != "" {
					u += sep + "pageSize=" + p.PageSize
					sep = "&"
				}
				if p.NextToken != "" {
					u += sep + "nextToken=" + p.NextToken
				}
				return c.DoRaw(ctx, http.MethodGet, u, nil)
			},
		},
		{
			ToolName: "cloud_host_get", ToolDesc: "Get host details by ID (Cloud API)",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionRead,
			Schema: core.IDSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, http.MethodGet, baseURL+"/v1/hosts/"+p.ID, nil)
			},
		},
		{
			ToolName: "cloud_site_list", ToolDesc: "List all sites across all hosts (Cloud API)",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionRead,
			Schema: paginatedSchema(""), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					PageSize  string `json:"pageSize"`
					NextToken string `json:"nextToken"`
				}
				json.Unmarshal(input, &p)
				u := baseURL + "/v1/sites"
				sep := "?"
				if p.PageSize != "" {
					u += sep + "pageSize=" + p.PageSize
					sep = "&"
				}
				if p.NextToken != "" {
					u += sep + "nextToken=" + p.NextToken
				}
				return c.DoRaw(ctx, http.MethodGet, u, nil)
			},
		},
		{
			ToolName: "cloud_device_list", ToolDesc: "List devices grouped by host (Cloud API)",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionRead,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"hostIds": {"type": "array", "items": {"type": "string"}, "description": "Filter by host IDs"}
				}
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ HostIDs []string `json:"hostIds"` }
				json.Unmarshal(input, &p)
				u := baseURL + "/v1/devices"
				for i, id := range p.HostIDs {
					if i == 0 {
						u += "?"
					} else {
						u += "&"
					}
					u += "hostIds=" + id
				}
				return c.DoRaw(ctx, http.MethodGet, u, nil)
			},
		},
		{
			ToolName: "cloud_isp_metrics", ToolDesc: "Get ISP metrics (latency, bandwidth, uptime) for all sites (Cloud API)",
			ToolCategory: permissions.CatStats, ToolAction: permissions.ActionRead,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"type": {"type": "string", "enum": ["5m", "1h"], "description": "Metric interval type"},
					"duration": {"type": "string", "enum": ["24h", "7d", "30d"], "description": "Time range"}
				},
				"required": ["type"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Type     string `json:"type"`
					Duration string `json:"duration"`
				}
				json.Unmarshal(input, &p)
				u := baseURL + "/v1/isp-metrics/" + p.Type
				if p.Duration != "" {
					u += "?duration=" + p.Duration
				}
				return c.DoRaw(ctx, http.MethodGet, u, nil)
			},
		},
		{
			ToolName: "cloud_isp_metrics_query", ToolDesc: "Query ISP metrics for specific sites with per-site time ranges (Cloud API)",
			ToolCategory: permissions.CatStats, ToolAction: permissions.ActionRead,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"type": {"type": "string", "enum": ["5m", "1h"], "description": "Metric interval type"},
					"config": {"type": "object", "description": "Query payload with site-specific time ranges"}
				},
				"required": ["type", "config"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Type   string          `json:"type"`
					Config json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, http.MethodPost, baseURL+"/v1/isp-metrics/"+p.Type+"/query", p.Config)
			},
		},
		{
			ToolName: "cloud_sdwan_list", ToolDesc: "List SD-WAN configurations (Cloud API)",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoRaw(ctx, http.MethodGet, baseURL+"/v1/sd-wan-configs", nil)
			},
		},
		{
			ToolName: "cloud_sdwan_get", ToolDesc: "Get SD-WAN config details (Cloud API)",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionRead,
			Schema: core.IDSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, http.MethodGet, baseURL+"/v1/sd-wan-configs/"+p.ID, nil)
			},
		},
		{
			ToolName: "cloud_sdwan_status", ToolDesc: "Get SD-WAN deployment status (Cloud API)",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionRead,
			Schema: core.IDSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, http.MethodGet, baseURL+"/v1/sd-wan-configs/"+p.ID+"/status", nil)
			},
		},
	}
}
