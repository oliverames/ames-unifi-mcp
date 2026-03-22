package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/oliveames/ames-unifi-mcp/internal/client"
	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
)

func BuildStatsTools(c *client.Client) []*BaseTool {
	sp := func() string { return fmt.Sprintf("api/s/%s", c.Site()) }

	return []*BaseTool{
		{
			ToolName: "stats_site_health", ToolDesc: "Get site health status (subsystem health, ISP info, latency)",
			ToolCategory: permissions.CatStats, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/stat/health", nil)
			},
		},
		{
			ToolName: "stats_sysinfo", ToolDesc: "Get controller system info (version, uptime, memory)",
			ToolCategory: permissions.CatStats, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/stat/sysinfo", nil)
			},
		},
		{
			ToolName: "stats_dashboard", ToolDesc: "Get dashboard metrics (bandwidth, client counts)",
			ToolCategory: permissions.CatStats, ToolAction: permissions.ActionRead,
			Schema: json.RawMessage(`{"type":"object","properties":{"scale":{"type":"string","description":"Time scale (e.g. 5minutes)","default":"5minutes"}}}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Scale string `json:"scale"` }
				json.Unmarshal(input, &p)
				path := sp() + "/stat/dashboard"
				if p.Scale != "" {
					path += "?scale=" + p.Scale
				}
				return c.Do(ctx, "GET", path, nil)
			},
		},
		{
			ToolName: "stats_dpi_site", ToolDesc: "Get site-wide DPI statistics by application or category",
			ToolCategory: permissions.CatDPI, ToolAction: permissions.ActionRead,
			Schema: json.RawMessage(`{"type":"object","properties":{"type":{"type":"string","enum":["by_app","by_cat"],"default":"by_app"}}}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Type string `json:"type"` }
				json.Unmarshal(input, &p)
				if p.Type == "" {
					p.Type = "by_app"
				}
				return c.Do(ctx, "POST", sp()+"/stat/sitedpi", map[string]interface{}{"type": p.Type})
			},
		},
		{
			ToolName: "stats_dpi_client", ToolDesc: "Get DPI stats for a specific client [undocumented]",
			ToolCategory: permissions.CatDPI, ToolAction: permissions.ActionRead, Undocumented: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"mac":{"type":"string","description":"Client MAC"},"type":{"type":"string","enum":["by_app","by_cat"],"default":"by_app"}},"required":["mac"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Mac  string `json:"mac"`
					Type string `json:"type"`
				}
				json.Unmarshal(input, &p)
				if p.Type == "" {
					p.Type = "by_app"
				}
				return c.Do(ctx, "POST", sp()+"/stat/stadpi", map[string]interface{}{"type": p.Type, "macs": []string{p.Mac}})
			},
		},
		{
			ToolName: "stats_speedtest_history", ToolDesc: "Get speed test history",
			ToolCategory: permissions.CatStats, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "POST", sp()+"/stat/report/archive.speedtest", map[string]interface{}{
					"attrs": []string{"xput_download", "xput_upload", "latency", "time"},
				})
			},
		},
		{
			ToolName: "stats_speedtest_run", ToolDesc: "Start a speed test on the gateway",
			ToolCategory: permissions.CatStats, ToolAction: permissions.ActionExecute, Mutating: true,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "POST", sp()+"/cmd/devmgr", map[string]interface{}{"cmd": "speedtest"})
			},
		},
		{
			ToolName: "stats_active_routes", ToolDesc: "Get active routing table",
			ToolCategory: permissions.CatRouting, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/stat/routing", nil)
			},
		},
		// --- new stats tools ---
		{
			ToolName: "stats_report", ToolDesc: "Get time-series report (bandwidth, clients, performance). Interval: 5minutes/hourly/daily/monthly. Type: site/ap/user/gw.",
			ToolCategory: permissions.CatStats, ToolAction: permissions.ActionRead,
			Schema: json.RawMessage(`{"type":"object","properties":{"interval":{"type":"string","enum":["5minutes","hourly","daily","monthly"],"description":"Report interval"},"type":{"type":"string","enum":["site","ap","user","gw"],"description":"Report type"},"start":{"type":"integer","description":"Start timestamp in milliseconds"},"end":{"type":"integer","description":"End timestamp in milliseconds"},"attrs":{"type":"array","items":{"type":"string"},"description":"Attributes to include"},"macs":{"type":"array","items":{"type":"string"},"description":"Filter by MAC addresses"}},"required":["interval","type","start","end"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Interval string   `json:"interval"`
					Type     string   `json:"type"`
					Start    int64    `json:"start"`
					End      int64    `json:"end"`
					Attrs    []string `json:"attrs,omitempty"`
					Macs     []string `json:"macs,omitempty"`
				}
				json.Unmarshal(input, &p)
				payload := map[string]interface{}{"start": p.Start, "end": p.End}
				if len(p.Attrs) > 0 {
					payload["attrs"] = p.Attrs
				}
				if len(p.Macs) > 0 {
					payload["macs"] = p.Macs
				}
				return c.Do(ctx, "POST", sp()+"/stat/report/"+p.Interval+"."+p.Type, payload)
			},
		},
		{
			ToolName: "stats_ips_events", ToolDesc: "List IPS/IDS security events (intrusion detections, blocked threats)",
			ToolCategory: permissions.CatStats, ToolAction: permissions.ActionRead, MinVer: "5.9.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"start":{"type":"integer","description":"Start timestamp in milliseconds"},"end":{"type":"integer","description":"End timestamp in milliseconds"},"limit":{"type":"integer","description":"Max events to return","default":1000}},"required":["start","end"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Start int64 `json:"start"`
					End   int64 `json:"end"`
					Limit int   `json:"limit"`
				}
				json.Unmarshal(input, &p)
				if p.Limit == 0 {
					p.Limit = 1000
				}
				return c.Do(ctx, "POST", sp()+"/stat/ips/event", map[string]interface{}{
					"start": p.Start, "end": p.End, "_limit": p.Limit,
				})
			},
		},
		{
			ToolName: "stats_rogueap", ToolDesc: "Detect neighboring/rogue access points within range",
			ToolCategory: permissions.CatStats, ToolAction: permissions.ActionRead,
			Schema: json.RawMessage(`{"type":"object","properties":{"within":{"type":"integer","description":"Hours to look back","default":24}}}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Within int `json:"within"`
				}
				json.Unmarshal(input, &p)
				if p.Within == 0 {
					p.Within = 24
				}
				return c.Do(ctx, "POST", sp()+"/stat/rogueap", map[string]interface{}{"within": p.Within})
			},
		},
		{
			ToolName: "stats_rf_channels", ToolDesc: "List available RF channels based on site country setting",
			ToolCategory: permissions.CatStats, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/stat/current-channel", nil)
			},
		},
		{
			ToolName: "stats_country_codes", ToolDesc: "List available country codes (ISO 3166-1)",
			ToolCategory: permissions.CatStats, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/stat/ccode", nil)
			},
		},
		{
			ToolName: "stats_controller_status", ToolDesc: "Get controller status (version, UUID, uptime). No auth required.",
			ToolCategory: permissions.CatStats, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				url := strings.TrimRight(c.Config().Host, "/") + "/status"
				return c.DoRaw(ctx, "GET", url, nil)
			},
		},
		{
			ToolName: "stats_sites_health", ToolDesc: "List all sites with health and alert data",
			ToolCategory: permissions.CatStats, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", "api/stat/sites", nil)
			},
		},
		{
			ToolName: "stats_firmware_update_check", ToolDesc: "Check if controller updates are available",
			ToolCategory: permissions.CatStats, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/stat/fwupdate/latest-version", nil)
			},
		},
		{
			ToolName: "stats_dpi_apps", ToolDesc: "List all DPI application definitions",
			ToolCategory: permissions.CatDPI, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				base := c.Config().BaseURL() + "/integration"
				return c.DoRaw(ctx, "GET", base+"/v1/dpi/applications", nil)
			},
		},
		{
			ToolName: "stats_dpi_categories", ToolDesc: "List all DPI application categories",
			ToolCategory: permissions.CatDPI, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				base := c.Config().BaseURL() + "/integration"
				return c.DoRaw(ctx, "GET", base+"/v1/dpi/categories", nil)
			},
		},
		{
			ToolName: "stats_dpi_reset", ToolDesc: "Reset site DPI counters [undocumented]",
			ToolCategory: permissions.CatDPI, ToolAction: permissions.ActionExecute, Mutating: true, Undocumented: true,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "POST", sp()+"/cmd/stat", map[string]interface{}{"cmd": "clear-dpi"})
			},
		},
		{
			ToolName: "stats_spectrumscan", ToolDesc: "Get RF spectrum scan results for all APs",
			ToolCategory: permissions.CatStats, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/stat/spectrumscan", nil)
			},
		},
		{
			ToolName: "stats_spectrumscan_device", ToolDesc: "Get RF spectrum scan results for a specific AP",
			ToolCategory: permissions.CatStats, ToolAction: permissions.ActionRead,
			Schema: macSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Mac string `json:"mac"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "GET", sp()+"/stat/spectrumscan/"+p.Mac, nil)
			},
		},
		{
			ToolName: "stats_voucher_list", ToolDesc: "List hotspot vouchers",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionRead,
			Schema: json.RawMessage(`{"type":"object","properties":{"create_time":{"type":"integer","description":"Filter by creation time (unix timestamp)"}}}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					CreateTime *int64 `json:"create_time,omitempty"`
				}
				json.Unmarshal(input, &p)
				var payload map[string]interface{}
				if p.CreateTime != nil {
					payload = map[string]interface{}{"create_time": *p.CreateTime}
				}
				return c.Do(ctx, "POST", sp()+"/stat/voucher", payload)
			},
		},
		{
			ToolName: "stats_authorization", ToolDesc: "List authorization codes used in timeframe [undocumented]",
			ToolCategory: permissions.CatStats, ToolAction: permissions.ActionRead, Undocumented: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"start":{"type":"integer","description":"Start timestamp (unix seconds)"},"end":{"type":"integer","description":"End timestamp (unix seconds)"}},"required":["start","end"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Start int64 `json:"start"`
					End   int64 `json:"end"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/stat/authorization", map[string]interface{}{
					"start": p.Start, "end": p.End,
				})
			},
		},
		{
			ToolName: "stats_sdn", ToolDesc: "Get UniFi Cloud/SSO connection status [undocumented]",
			ToolCategory: permissions.CatStats, ToolAction: permissions.ActionRead, Undocumented: true,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/stat/sdn", nil)
			},
		},
	}
}
