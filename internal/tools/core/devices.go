package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliveames/ames-unifi-mcp/internal/client"
	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
)

// BuildDeviceTools returns all device tools.
func BuildDeviceTools(c *client.Client) []*BaseTool {
	sp := func() string { return fmt.Sprintf("api/s/%s", c.Site()) }

	return []*BaseTool{
		{
			ToolName: "device_list", ToolDesc: "List all adopted UniFi devices with full details (APs, switches, gateways)",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/stat/device", nil)
			},
		},
		{
			ToolName: "device_list_basic", ToolDesc: "List devices with minimal keys (mac, type, state). Faster than device_list.",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/stat/device-basic", nil)
			},
		},
		{
			ToolName: "device_get", ToolDesc: "Get full details for a specific device by MAC address",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionRead,
			Schema: macSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Mac string `json:"mac"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "GET", sp()+"/stat/device/"+p.Mac, nil)
			},
		},
		{
			ToolName: "device_restart", ToolDesc: "Restart a UniFi device. Use reboot_type 'soft' (default) or 'hard' (power-cycles PoE).",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionExecute, Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"mac": {"type": "string", "description": "Device MAC address"},
					"reboot_type": {"type": "string", "enum": ["soft", "hard"], "default": "soft"}
				},
				"required": ["mac"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Mac        string `json:"mac"`
					RebootType string `json:"reboot_type"`
				}
				json.Unmarshal(input, &p)
				if p.RebootType == "" {
					p.RebootType = "soft"
				}
				return c.Do(ctx, "POST", sp()+"/cmd/devmgr", map[string]interface{}{
					"cmd": "restart", "mac": p.Mac, "reboot_type": p.RebootType,
				})
			},
		},
		{
			ToolName: "device_adopt", ToolDesc: "Adopt a new device by MAC address",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionExecute, Mutating: true,
			Schema: macSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Mac string `json:"mac"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/cmd/devmgr", map[string]interface{}{"cmd": "adopt", "mac": p.Mac})
			},
		},
		{
			ToolName: "device_locate_on", ToolDesc: "Blink LED to locate device [undocumented]",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionExecute, Mutating: true, Undocumented: true,
			Schema: macSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Mac string `json:"mac"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/cmd/devmgr", map[string]interface{}{"cmd": "set-locate", "mac": p.Mac})
			},
		},
		{
			ToolName: "device_locate_off", ToolDesc: "Stop blinking LED [undocumented]",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionExecute, Mutating: true, Undocumented: true,
			Schema: macSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Mac string `json:"mac"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/cmd/devmgr", map[string]interface{}{"cmd": "unset-locate", "mac": p.Mac})
			},
		},
		{
			ToolName: "device_upgrade", ToolDesc: "Upgrade device firmware to latest stable",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionExecute, Mutating: true,
			Schema: macSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Mac string `json:"mac"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/cmd/devmgr", map[string]interface{}{"cmd": "upgrade", "mac": p.Mac})
			},
		},
		{
			ToolName: "device_force_provision", ToolDesc: "Force re-provision device (push current config)",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionExecute, Mutating: true,
			Schema: macSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Mac string `json:"mac"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/cmd/devmgr", map[string]interface{}{"cmd": "force-provision", "mac": p.Mac})
			},
		},
		// --- Integration API tools (9.0+) ---
		{
			ToolName: "device_list_v2", ToolDesc: "List all adopted devices (Integration API, structured response)",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				base := c.Config().BaseURL() + "/integration"
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/devices", base, c.Site()), nil)
			},
		},
		{
			ToolName: "device_get_v2", ToolDesc: "Get device details by device ID (Integration API)",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"device_id": {"type": "string", "description": "Device ID"}
				},
				"required": ["device_id"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ DeviceID string `json:"device_id"` }
				json.Unmarshal(input, &p)
				base := c.Config().BaseURL() + "/integration"
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/devices/%s", base, c.Site(), p.DeviceID), nil)
			},
		},
		{
			ToolName: "device_adopt_v2", ToolDesc: "Adopt devices (Integration API)",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionExecute, Mutating: true, MinVer: "9.0.0",
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"config": {"type": "object", "description": "Device adoption payload"}
				},
				"required": ["config"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				base := c.Config().BaseURL() + "/integration"
				return c.DoRaw(ctx, "POST", fmt.Sprintf("%s/v1/sites/%s/devices", base, c.Site()), p.Config)
			},
		},
		{
			ToolName: "device_action", ToolDesc: "Execute a device action (Integration API)",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionExecute, Mutating: true, MinVer: "9.0.0",
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"device_id": {"type": "string", "description": "Device ID"},
					"config": {"type": "object", "description": "Action payload"}
				},
				"required": ["device_id", "config"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					DeviceID string          `json:"device_id"`
					Config   json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				base := c.Config().BaseURL() + "/integration"
				return c.DoRaw(ctx, "POST", fmt.Sprintf("%s/v1/sites/%s/devices/%s/actions", base, c.Site(), p.DeviceID), p.Config)
			},
		},
		{
			ToolName: "device_pending_list", ToolDesc: "List devices pending adoption",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				base := c.Config().BaseURL() + "/integration"
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/pending-devices", base), nil)
			},
		},
		{
			ToolName: "device_stats_latest", ToolDesc: "Get latest statistics for a specific device (Integration API)",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"device_id": {"type": "string", "description": "Device ID"}
				},
				"required": ["device_id"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ DeviceID string `json:"device_id"` }
				json.Unmarshal(input, &p)
				base := c.Config().BaseURL() + "/integration"
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/devices/%s/statistics/latest", base, c.Site(), p.DeviceID), nil)
			},
		},
		{
			ToolName: "device_unadopt", ToolDesc: "Remove/unadopt a device",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionDelete, Mutating: true, MinVer: "9.0.0",
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"device_id": {"type": "string", "description": "Device ID"}
				},
				"required": ["device_id"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ DeviceID string `json:"device_id"` }
				json.Unmarshal(input, &p)
				base := c.Config().BaseURL() + "/integration"
				return c.DoRaw(ctx, "DELETE", fmt.Sprintf("%s/v1/sites/%s/devices/%s", base, c.Site(), p.DeviceID), nil)
			},
		},
		{
			ToolName: "device_port_action", ToolDesc: "Execute an action on a device port (Integration API)",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionExecute, Mutating: true, MinVer: "9.0.0",
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"device_id": {"type": "string", "description": "Device ID"},
					"port_idx": {"type": "integer", "description": "Port index"},
					"config": {"type": "object", "description": "Action configuration payload"}
				},
				"required": ["device_id", "port_idx", "config"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					DeviceID string          `json:"device_id"`
					PortIdx  int             `json:"port_idx"`
					Config   json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				base := c.Config().BaseURL() + "/integration"
				return c.DoRaw(ctx, "POST", fmt.Sprintf("%s/v1/sites/%s/devices/%s/interfaces/ports/%d/actions", base, c.Site(), p.DeviceID, p.PortIdx), p.Config)
			},
		},
		// --- Legacy API tools ---
		{
			ToolName: "device_update", ToolDesc: "Update device settings (name, radio, port overrides, etc.)",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"id": {"type": "string", "description": "Device _id"},
					"config": {"type": "object", "description": "Device configuration object"}
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
				return c.Do(ctx, "PUT", sp()+"/rest/device/"+p.ID, p.Config)
			},
		},
		{
			ToolName: "device_spectrum_scan", ToolDesc: "Trigger RF spectrum scan on an AP",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionExecute, Mutating: true,
			Schema: macSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Mac string `json:"mac"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/cmd/devmgr", map[string]interface{}{"cmd": "spectrum-scan", "mac": p.Mac})
			},
		},
		{
			ToolName: "device_speedtest_status", ToolDesc: "Check speed test status/results",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "POST", sp()+"/cmd/devmgr", map[string]interface{}{"cmd": "speedtest-status"})
			},
		},
		{
			ToolName: "device_rolling_upgrade_start", ToolDesc: "Start rolling firmware upgrade across all devices",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionExecute, Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"device_types": {
						"type": "array",
						"items": {"type": "string"},
						"description": "Device types to upgrade (default: uap, usw, ugw, uxg)",
						"default": ["uap", "usw", "ugw", "uxg"]
					}
				}
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					DeviceTypes []string `json:"device_types"`
				}
				json.Unmarshal(input, &p)
				if len(p.DeviceTypes) == 0 {
					p.DeviceTypes = []string{"uap", "usw", "ugw", "uxg"}
				}
				return c.Do(ctx, "POST", sp()+"/cmd/devmgr", map[string]interface{}{
					"cmd": "set-rollupgrade", "device_types": p.DeviceTypes,
				})
			},
		},
		{
			ToolName: "device_rolling_upgrade_cancel", ToolDesc: "Cancel rolling firmware upgrade",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionExecute, Mutating: true,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "POST", sp()+"/cmd/devmgr", map[string]interface{}{"cmd": "unset-rollupgrade"})
			},
		},
		{
			ToolName: "device_delete", ToolDesc: "Remove device from site (does not factory reset)",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionDelete, Mutating: true,
			Schema: macSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Mac string `json:"mac"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/cmd/sitemgr", map[string]interface{}{"cmd": "delete-device", "mac": p.Mac})
			},
		},
		{
			ToolName: "device_adv_adopt", ToolDesc: "Adopt device via custom SSH credentials",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionExecute, Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"mac": {"type": "string", "description": "Device MAC address"},
					"ip": {"type": "string", "description": "Device IP address"},
					"username": {"type": "string", "description": "SSH username"},
					"password": {"type": "string", "description": "SSH password"},
					"url": {"type": "string", "description": "Inform URL"},
					"port": {"type": "integer", "description": "SSH port (default: 22)"}
				},
				"required": ["mac", "ip", "username", "password", "url"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Mac      string `json:"mac"`
					IP       string `json:"ip"`
					Username string `json:"username"`
					Password string `json:"password"`
					URL      string `json:"url"`
					Port     int    `json:"port"`
				}
				json.Unmarshal(input, &p)
				payload := map[string]interface{}{
					"cmd": "adv-adopt", "mac": p.Mac, "ip": p.IP,
					"username": p.Username, "password": p.Password, "url": p.URL,
				}
				if p.Port > 0 {
					payload["port"] = p.Port
				}
				return c.Do(ctx, "POST", sp()+"/cmd/devmgr", payload)
			},
		},
		{
			ToolName: "device_upgrade_external", ToolDesc: "Upgrade device to firmware at a specific URL",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionExecute, Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"mac": {"type": "string", "description": "Device MAC address"},
					"url": {"type": "string", "description": "Firmware URL to upgrade to"}
				},
				"required": ["mac", "url"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Mac string `json:"mac"`
					URL string `json:"url"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/cmd/devmgr", map[string]interface{}{
					"cmd": "upgrade-external", "mac": p.Mac, "url": p.URL,
				})
			},
		},
		{
			ToolName: "device_migrate", ToolDesc: "Push new inform URL to device(s) to migrate to another controller",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionExecute, Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"macs": {"type": "array", "items": {"type": "string"}, "description": "Device MAC address(es)"},
					"inform_url": {"type": "string", "description": "New inform URL for the target controller"}
				},
				"required": ["macs", "inform_url"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Macs      []string `json:"macs"`
					InformURL string   `json:"inform_url"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/cmd/devmgr", map[string]interface{}{
					"cmd": "migrate", "macs": p.Macs, "inform_url": p.InformURL,
				})
			},
		},
		{
			ToolName: "device_cancel_migrate", ToolDesc: "Cancel device migration",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionExecute, Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"macs": {"type": "array", "items": {"type": "string"}, "description": "Device MAC address(es)"}
				},
				"required": ["macs"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Macs []string `json:"macs"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/cmd/devmgr", map[string]interface{}{
					"cmd": "cancel-migrate", "macs": p.Macs,
				})
			},
		},
		{
			ToolName: "device_upgrade_all", ToolDesc: "Upgrade firmware on all devices at once",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionExecute, Mutating: true,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "POST", sp()+"/cmd/devmgr", map[string]interface{}{"cmd": "upgrade-all"})
			},
		},
		{
			ToolName: "device_rename", ToolDesc: "Rename a device",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"id": {"type": "string", "description": "Device _id"},
					"name": {"type": "string", "description": "New device name"}
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
				return c.Do(ctx, "POST", sp()+"/upd/device/"+p.ID, map[string]interface{}{"name": p.Name})
			},
		},
	}
}
