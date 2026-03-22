package extended

import (
	"context"
	"encoding/json"
	"fmt"

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
	}
}
