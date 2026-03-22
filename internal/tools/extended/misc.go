package extended

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/oliveames/ames-unifi-mcp/internal/client"
	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
	"github.com/oliveames/ames-unifi-mcp/internal/tools/core"
)

func BuildMiscTools(c *client.Client) []*core.BaseTool {
	sp := func() string { return fmt.Sprintf("api/s/%s", c.Site()) }

	return []*core.BaseTool{
		{
			ToolName: "misc_rogueknown_list", ToolDesc: "List known/acknowledged rogue APs",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/rogueknown", nil)
			},
		},
		{
			ToolName: "misc_wlangroup_list", ToolDesc: "List WLAN groups",
			ToolCategory: permissions.CatWLAN, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/wlangroup", nil)
			},
		},
		{
			ToolName: "misc_hotspotop_list", ToolDesc: "List hotspot operators",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/hotspotop", nil)
			},
		},
		{
			ToolName: "misc_broadcastgroup_list", ToolDesc: "List broadcast/multicast groups",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/broadcastgroup", nil)
			},
		},
		{
			ToolName: "misc_dynamicdns_config", ToolDesc: "Get dynamic DNS configuration",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/dynamicdns", nil)
			},
		},
		{
			ToolName: "misc_dynamicdns_update", ToolDesc: "Update dynamic DNS configuration",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"id": {"type": "string", "description": "Dynamic DNS config _id"},
					"config": {"type": "object", "description": "Updated dynamic DNS configuration"}
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
				return c.Do(ctx, "PUT", sp()+"/rest/dynamicdns/"+p.ID, p.Config)
			},
		},
		{
			ToolName: "misc_firmware_check", ToolDesc: "Trigger firmware update check",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "POST", sp()+"/cmd/productinfo", map[string]interface{}{
					"cmd": "check-firmware-update",
				})
			},
		},
		{
			ToolName: "misc_firmware_cached", ToolDesc: "List cached firmware files",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "POST", sp()+"/cmd/firmware", map[string]interface{}{
					"cmd": "list-cached",
				})
			},
		},
		// --- More REST resources ---
		{
			ToolName: "misc_channelplan_list", ToolDesc: "List WiFi channel plans",
			ToolCategory: permissions.CatWLAN, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/channelplan", nil)
			},
		},
		{
			ToolName: "misc_dpiapp_list", ToolDesc: "List DPI application definitions",
			ToolCategory: permissions.CatDPI, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/dpiapp", nil)
			},
		},
		{
			ToolName: "misc_dpigroup_list", ToolDesc: "List DPI application groups",
			ToolCategory: permissions.CatDPI, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/dpigroup", nil)
			},
		},
		{
			ToolName: "misc_dpigroup_create", ToolDesc: "Create a DPI application group",
			ToolCategory: permissions.CatDPI, ToolAction: permissions.ActionCreate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"DPI group configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/rest/dpigroup", p.Config)
			},
		},
		{
			ToolName: "misc_dpigroup_update", ToolDesc: "Update a DPI application group",
			ToolCategory: permissions.CatDPI, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string","description":"DPI group _id"},"config":{"type":"object","description":"DPI group configuration"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "PUT", sp()+"/rest/dpigroup/"+p.ID, p.Config)
			},
		},
		{
			ToolName: "misc_dpigroup_delete", ToolDesc: "Delete a DPI application group",
			ToolCategory: permissions.CatDPI, ToolAction: permissions.ActionDelete, Mutating: true,
			Schema: core.IDSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "DELETE", sp()+"/rest/dpigroup/"+p.ID, nil)
			},
		},
		{
			ToolName: "misc_hotspot2conf_list", ToolDesc: "List Hotspot 2.0/Passpoint configurations",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/hotspot2conf", nil)
			},
		},
		{
			ToolName: "misc_hotspotpackage_list", ToolDesc: "List hotspot billing packages",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/hotspotpackage", nil)
			},
		},
		{
			ToolName: "misc_scheduletask_list", ToolDesc: "List scheduled tasks",
			ToolCategory: permissions.CatSettings, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/scheduletask", nil)
			},
		},
		{
			ToolName: "misc_map_list", ToolDesc: "List site maps",
			ToolCategory: permissions.CatSettings, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/map", nil)
			},
		},
		{
			ToolName: "misc_heatmap_list", ToolDesc: "List RF heatmaps",
			ToolCategory: permissions.CatSettings, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/heatmap", nil)
			},
		},
		{
			ToolName: "misc_heatmappoint_list", ToolDesc: "List RF heatmap data points",
			ToolCategory: permissions.CatSettings, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/heatmappoint", nil)
			},
		},
		// --- stat/ endpoints not yet covered ---
		{
			ToolName: "misc_stat_payment", ToolDesc: "Get hotspot payment history",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionRead,
			Schema: json.RawMessage(`{"type":"object","properties":{"within":{"type":"integer","description":"Hours of history","default":8760}}}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Within int `json:"within"` }
				json.Unmarshal(input, &p)
				if p.Within == 0 {
					p.Within = 8760
				}
				return c.Do(ctx, "GET", fmt.Sprintf("%s/stat/payment?within=%d", sp(), p.Within), nil)
			},
		},
		{
			ToolName: "misc_stat_portforward", ToolDesc: "Get port forwarding statistics",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/stat/portforward", nil)
			},
		},
		{
			ToolName: "misc_stat_dpi", ToolDesc: "Get raw DPI stats",
			ToolCategory: permissions.CatDPI, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/stat/dpi", nil)
			},
		},
		{
			ToolName: "misc_firmware_bundles", ToolDesc: "Get device firmware bundle mappings",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				host := strings.TrimRight(c.Config().Host, "/")
				return c.DoRaw(ctx, "GET", host+"/dl/firmware/bundles.json", nil)
			},
		},
		// --- list/ endpoints ---
		{
			ToolName: "misc_list_extension", ToolDesc: "List VoIP extensions",
			ToolCategory: permissions.CatSettings, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/list/extension", nil)
			},
		},
		{
			ToolName: "misc_alarm_list_filtered", ToolDesc: "List alarms with optional key/type filter (e.g., EVT_GW_WANTransition)",
			ToolCategory: permissions.CatEvents, ToolAction: permissions.ActionRead,
			Schema: json.RawMessage(`{"type":"object","properties":{"archived":{"type":"boolean","description":"Include archived alarms","default":false},"key":{"type":"string","description":"Filter by event key (e.g., EVT_GW_WANTransition, EVT_AP_Disconnected)"}}}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Archived bool   `json:"archived"`
					Key      string `json:"key"`
				}
				json.Unmarshal(input, &p)
				payload := map[string]interface{}{"archived": p.Archived}
				if p.Key != "" {
					payload["key"] = p.Key
				}
				return c.Do(ctx, "POST", sp()+"/list/alarm", payload)
			},
		},
		// --- REST event/resource endpoints ---
		{
			ToolName: "misc_event_list_rest", ToolDesc: "List events via REST endpoint (oldest first, use for pagination by _id)",
			ToolCategory: permissions.CatEvents, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/event", nil)
			},
		},
		{
			ToolName: "misc_dashboard_list", ToolDesc: "List custom dashboard configurations",
			ToolCategory: permissions.CatSettings, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/dashboard", nil)
			},
		},
		{
			ToolName: "misc_rogueknown_create", ToolDesc: "Mark a rogue AP as known/acknowledged",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionCreate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"Rogue AP acknowledgement (mac, name, etc.)"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/rest/rogueknown", p.Config)
			},
		},
		{
			ToolName: "misc_rogueknown_delete", ToolDesc: "Remove a rogue AP from the known list",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionDelete, Mutating: true,
			Schema: core.IDSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "DELETE", sp()+"/rest/rogueknown/"+p.ID, nil)
			},
		},
		{
			ToolName: "misc_cnt_resource", ToolDesc: "Count resources by type (e.g., alarm, sta, device, networkconf)",
			ToolCategory: permissions.CatStats, ToolAction: permissions.ActionRead,
			Schema: json.RawMessage(`{"type":"object","properties":{"resource":{"type":"string","description":"Resource name (e.g., alarm, sta, device, networkconf, wlanconf, firewallrule)"},"filter":{"type":"string","description":"Optional query filter (e.g., archived=false)"}},"required":["resource"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Resource string `json:"resource"`
					Filter   string `json:"filter"`
				}
				json.Unmarshal(input, &p)
				path := sp() + "/cnt/" + p.Resource
				if p.Filter != "" {
					path += "?" + p.Filter
				}
				return c.Do(ctx, "GET", path, nil)
			},
		},
		{
			ToolName: "misc_self", ToolDesc: "Get current logged-in user info",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", "api/users/self", nil)
			},
		},
		{
			ToolName: "misc_site_export", ToolDesc: "Export current site configuration",
			ToolCategory: permissions.CatBackup, ToolAction: permissions.ActionCreate, Mutating: true,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "POST", sp()+"/cmd/backup", map[string]interface{}{
					"cmd": "export-site",
				})
			},
		},
		{
			ToolName: "misc_site_restore", ToolDesc: "Restore site from a backup file",
			ToolCategory: permissions.CatBackup, ToolAction: permissions.ActionExecute, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"filename":{"type":"string","description":"Backup filename to restore"}},"required":["filename"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Filename string `json:"filename"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/cmd/backup", map[string]interface{}{
					"cmd": "restore-site", "filename": p.Filename,
				})
			},
		},
		// --- CRUD for hotspot operators ---
		{
			ToolName: "misc_hotspotop_create", ToolDesc: "Create a hotspot operator",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionCreate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"Hotspot operator config (name, x_password, note)"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/rest/hotspotop", p.Config)
			},
		},
		{
			ToolName: "misc_hotspotop_update", ToolDesc: "Update a hotspot operator",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string","description":"Operator _id"},"config":{"type":"object","description":"Updated operator config"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "PUT", sp()+"/rest/hotspotop/"+p.ID, p.Config)
			},
		},
		{
			ToolName: "misc_hotspotop_delete", ToolDesc: "Delete a hotspot operator",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionDelete, Mutating: true,
			Schema: core.IDSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "DELETE", sp()+"/rest/hotspotop/"+p.ID, nil)
			},
		},
		// --- CRUD for Hotspot 2.0/Passpoint ---
		{
			ToolName: "misc_hotspot2conf_create", ToolDesc: "Create a Hotspot 2.0/Passpoint configuration",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionCreate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"Hotspot 2.0 configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/rest/hotspot2conf", p.Config)
			},
		},
		{
			ToolName: "misc_hotspot2conf_update", ToolDesc: "Update a Hotspot 2.0/Passpoint configuration",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string","description":"Config _id"},"config":{"type":"object","description":"Updated config"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "PUT", sp()+"/rest/hotspot2conf/"+p.ID, p.Config)
			},
		},
		{
			ToolName: "misc_hotspot2conf_delete", ToolDesc: "Delete a Hotspot 2.0/Passpoint configuration",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionDelete, Mutating: true,
			Schema: core.IDSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "DELETE", sp()+"/rest/hotspot2conf/"+p.ID, nil)
			},
		},
		// --- CRUD for hotspot billing packages ---
		{
			ToolName: "misc_hotspotpackage_create", ToolDesc: "Create a hotspot billing package",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionCreate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"Billing package config (name, amount, currency, duration)"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/rest/hotspotpackage", p.Config)
			},
		},
		{
			ToolName: "misc_hotspotpackage_update", ToolDesc: "Update a hotspot billing package",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string","description":"Package _id"},"config":{"type":"object","description":"Updated package config"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "PUT", sp()+"/rest/hotspotpackage/"+p.ID, p.Config)
			},
		},
		{
			ToolName: "misc_hotspotpackage_delete", ToolDesc: "Delete a hotspot billing package",
			ToolCategory: permissions.CatHotspot, ToolAction: permissions.ActionDelete, Mutating: true,
			Schema: core.IDSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "DELETE", sp()+"/rest/hotspotpackage/"+p.ID, nil)
			},
		},
		// --- CRUD for scheduled tasks ---
		{
			ToolName: "misc_scheduletask_create", ToolDesc: "Create a scheduled task (e.g., WLAN schedule, device auto-upgrade)",
			ToolCategory: permissions.CatSettings, ToolAction: permissions.ActionCreate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"Scheduled task configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/rest/scheduletask", p.Config)
			},
		},
		{
			ToolName: "misc_scheduletask_update", ToolDesc: "Update a scheduled task",
			ToolCategory: permissions.CatSettings, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string","description":"Task _id"},"config":{"type":"object","description":"Updated task config"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "PUT", sp()+"/rest/scheduletask/"+p.ID, p.Config)
			},
		},
		{
			ToolName: "misc_scheduletask_delete", ToolDesc: "Delete a scheduled task",
			ToolCategory: permissions.CatSettings, ToolAction: permissions.ActionDelete, Mutating: true,
			Schema: core.IDSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "DELETE", sp()+"/rest/scheduletask/"+p.ID, nil)
			},
		},
	}
}
