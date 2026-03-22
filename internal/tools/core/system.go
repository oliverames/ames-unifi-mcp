package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/oliveames/ames-unifi-mcp/internal/client"
	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
)

func BuildSystemTools(c *client.Client) []*BaseTool {
	sp := func() string { return fmt.Sprintf("api/s/%s", c.Site()) }
	base := c.Config().BaseURL() + "/integration"
	return []*BaseTool{
		// --- Integration API system tools ---
		{
			ToolName: "system_app_info", ToolDesc: "Get application info — version, runtime (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoRaw(ctx, "GET", base+"/v1/info", nil)
			},
		},
		{
			ToolName: "system_sites_v2", ToolDesc: "List local sites (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoRaw(ctx, "GET", base+"/v1/sites", nil)
			},
		},
		{
			ToolName: "system_countries", ToolDesc: "List available countries (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoRaw(ctx, "GET", base+"/v1/countries", nil)
			},
		},
		{
			ToolName: "system_reboot", ToolDesc: "Reboot the UniFi OS console (requires admin permission, X-CSRF-Token)",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionExecute, Mutating: true,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				host := strings.TrimRight(c.Config().Host, "/")
				url := host + "/api/system/reboot"
				return c.DoRaw(ctx, "POST", url, map[string]interface{}{})
			},
		},
		{
			ToolName: "system_poweroff", ToolDesc: "Power off the UniFi OS console (requires admin permission, X-CSRF-Token)",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionExecute, Mutating: true,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				host := strings.TrimRight(c.Config().Host, "/")
				url := host + "/api/system/poweroff"
				return c.DoRaw(ctx, "POST", url, map[string]interface{}{})
			},
		},
		// --- Legacy API tools ---
		{
			ToolName: "system_sites", ToolDesc: "List all sites accessible to the current user",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", "api/self/sites", nil)
			},
		},
		{
			ToolName: "system_settings", ToolDesc: "Get all site settings",
			ToolCategory: permissions.CatSettings, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/get/setting", nil)
			},
		},
		{
			ToolName: "system_admins", ToolDesc: "List administrators for the current site",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "POST", sp()+"/cmd/sitemgr", map[string]interface{}{"cmd": "get-admins"})
			},
		},
		{
			ToolName: "system_backup_trigger", ToolDesc: "Trigger a controller backup",
			ToolCategory: permissions.CatBackup, ToolAction: permissions.ActionExecute, Mutating: true,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "POST", sp()+"/cmd/backup", map[string]interface{}{"cmd": "backup"})
			},
		},
		{
			ToolName: "system_backup_list", ToolDesc: "List available auto-backup files",
			ToolCategory: permissions.CatBackup, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "POST", sp()+"/cmd/backup", map[string]interface{}{"cmd": "list-backups"})
			},
		},
		{
			ToolName: "system_firmware_available", ToolDesc: "List available firmware updates",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "POST", sp()+"/cmd/firmware", map[string]interface{}{"cmd": "list-available"})
			},
		},
		{
			ToolName: "system_port_forward_list", ToolDesc: "List all port forwarding rules",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/portforward", nil)
			},
		},
		{
			ToolName: "system_static_routes", ToolDesc: "List user-defined static routes",
			ToolCategory: permissions.CatRouting, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/routing", nil)
			},
		},
		{
			ToolName: "system_dyndns", ToolDesc: "Get dynamic DNS status",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/stat/dynamicdns", nil)
			},
		},
		// --- new system tools ---
		{
			ToolName: "system_setting_update", ToolDesc: "Update a site setting by key (e.g., mgmt for LED toggle, ips for IDS/IPS, guest_access, dpi)",
			ToolCategory: permissions.CatSettings, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"key":{"type":"string","description":"Setting key e.g. mgmt, ips, guest_access, dpi"},"config":{"type":"object","description":"Setting configuration"}},"required":["key","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Key    string                 `json:"key"`
					Config map[string]interface{} `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/set/setting/"+p.Key, p.Config)
			},
		},
		{
			ToolName: "system_setting_update_by_id", ToolDesc: "Update a specific site setting by key and _id (use for settings like super_mgmt, super_smtp that have unique _ids)",
			ToolCategory: permissions.CatSettings, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"key":{"type":"string","description":"Setting key"},"id":{"type":"string","description":"Setting _id"},"config":{"type":"object","description":"Setting configuration"}},"required":["key","id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Key    string                 `json:"key"`
					ID     string                 `json:"id"`
					Config map[string]interface{} `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "PUT", sp()+"/set/setting/"+p.Key+"/"+p.ID, p.Config)
			},
		},
		{
			ToolName: "system_led_toggle", ToolDesc: "Toggle site LED status on all devices",
			ToolCategory: permissions.CatSettings, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"enabled":{"type":"boolean","description":"true = LEDs on, false = LEDs off"}},"required":["enabled"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Enabled bool `json:"enabled"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/set/setting/mgmt", map[string]interface{}{"led_enabled": p.Enabled})
			},
		},
		{
			ToolName: "system_ips_update", ToolDesc: "Update IPS/IDS settings (enable/disable, sensitivity, etc.)",
			ToolCategory: permissions.CatSettings, ToolAction: permissions.ActionUpdate, Mutating: true, MinVer: "5.9.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"IPS settings"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Config map[string]interface{} `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/set/setting/ips", p.Config)
			},
		},
		{
			ToolName: "system_port_forward_create", ToolDesc: "Create a port forwarding rule",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionCreate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"Port forward rule configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Config map[string]interface{} `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/rest/portforward", p.Config)
			},
		},
		{
			ToolName: "system_port_forward_update", ToolDesc: "Update a port forwarding rule",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string","description":"The resource _id"},"config":{"type":"object","description":"Port forward rule configuration"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string                 `json:"id"`
					Config map[string]interface{} `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "PUT", sp()+"/rest/portforward/"+p.ID, p.Config)
			},
		},
		{
			ToolName: "system_port_forward_delete", ToolDesc: "Delete a port forwarding rule",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionDelete, Mutating: true,
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "DELETE", sp()+"/rest/portforward/"+p.ID, nil)
			},
		},
		{
			ToolName: "system_static_route_create", ToolDesc: "Create a static route",
			ToolCategory: permissions.CatRouting, ToolAction: permissions.ActionCreate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"Static route configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Config map[string]interface{} `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/rest/routing", p.Config)
			},
		},
		{
			ToolName: "system_static_route_update", ToolDesc: "Update a static route",
			ToolCategory: permissions.CatRouting, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string","description":"The resource _id"},"config":{"type":"object","description":"Static route configuration"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string                 `json:"id"`
					Config map[string]interface{} `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "PUT", sp()+"/rest/routing/"+p.ID, p.Config)
			},
		},
		{
			ToolName: "system_static_route_delete", ToolDesc: "Delete a static route",
			ToolCategory: permissions.CatRouting, ToolAction: permissions.ActionDelete, Mutating: true,
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "DELETE", sp()+"/rest/routing/"+p.ID, nil)
			},
		},
		{
			ToolName: "system_usergroup_list", ToolDesc: "List user groups (bandwidth limit profiles)",
			ToolCategory: permissions.CatSettings, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/usergroup", nil)
			},
		},
		{
			ToolName: "system_usergroup_create", ToolDesc: "Create a user group",
			ToolCategory: permissions.CatSettings, ToolAction: permissions.ActionCreate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"User group configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Config map[string]interface{} `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/rest/usergroup", p.Config)
			},
		},
		{
			ToolName: "system_usergroup_update", ToolDesc: "Update a user group",
			ToolCategory: permissions.CatSettings, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string","description":"The resource _id"},"config":{"type":"object","description":"User group configuration"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string                 `json:"id"`
					Config map[string]interface{} `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "PUT", sp()+"/rest/usergroup/"+p.ID, p.Config)
			},
		},
		{
			ToolName: "system_usergroup_delete", ToolDesc: "Delete a user group",
			ToolCategory: permissions.CatSettings, ToolAction: permissions.ActionDelete, Mutating: true,
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "DELETE", sp()+"/rest/usergroup/"+p.ID, nil)
			},
		},
		{
			ToolName: "system_setting_get", ToolDesc: "Get a specific site setting section by key name (e.g., mgmt, ips, guest_access, dpi, country, locale, connectivity)",
			ToolCategory: permissions.CatSettings, ToolAction: permissions.ActionRead,
			Schema: json.RawMessage(`{"type":"object","properties":{"key":{"type":"string","description":"Setting key name (e.g. mgmt, ips, guest_access, dpi, country, connectivity, snmp, ntp, rsyslogd, radius, super_smtp)"}},"required":["key"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Key string `json:"key"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "GET", sp()+"/get/setting/"+p.Key, nil)
			},
		},
		{
			ToolName: "system_portprofile_list", ToolDesc: "List switch port profiles",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/portconf", nil)
			},
		},
		{
			ToolName: "system_portprofile_create", ToolDesc: "Create a switch port profile",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionCreate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"Port profile configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/rest/portconf", p.Config)
			},
		},
		{
			ToolName: "system_portprofile_update", ToolDesc: "Update a switch port profile",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string","description":"Port profile _id"},"config":{"type":"object","description":"Port profile configuration"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "PUT", sp()+"/rest/portconf/"+p.ID, p.Config)
			},
		},
		{
			ToolName: "system_portprofile_delete", ToolDesc: "Delete a switch port profile",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionDelete, Mutating: true,
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "DELETE", sp()+"/rest/portconf/"+p.ID, nil)
			},
		},
		{
			ToolName: "system_dhcpoption_list", ToolDesc: "List custom DHCP options",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/dhcpoption", nil)
			},
		},
		{
			ToolName: "system_dhcpoption_create", ToolDesc: "Create a custom DHCP option",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionCreate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"DHCP option configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Config map[string]interface{} `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/rest/dhcpoption", p.Config)
			},
		},
		{
			ToolName: "system_dhcpoption_update", ToolDesc: "Update a DHCP option",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string","description":"The resource _id"},"config":{"type":"object","description":"DHCP option configuration"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string                 `json:"id"`
					Config map[string]interface{} `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "PUT", sp()+"/rest/dhcpoption/"+p.ID, p.Config)
			},
		},
		{
			ToolName: "system_dhcpoption_delete", ToolDesc: "Delete a DHCP option",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionDelete, Mutating: true,
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "DELETE", sp()+"/rest/dhcpoption/"+p.ID, nil)
			},
		},
		{
			ToolName: "system_radiusprofile_list", ToolDesc: "List RADIUS profiles",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionRead, MinVer: "5.5.19",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/radiusprofile", nil)
			},
		},
		{
			ToolName: "system_radius_account_list", ToolDesc: "List RADIUS accounts",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionRead, MinVer: "5.5.19",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/account", nil)
			},
		},
		{
			ToolName: "system_radiusprofile_create", ToolDesc: "Create a RADIUS profile",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionCreate, Mutating: true, MinVer: "5.5.19",
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"RADIUS profile configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/rest/radiusprofile", p.Config)
			},
		},
		{
			ToolName: "system_radiusprofile_update", ToolDesc: "Update a RADIUS profile",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionUpdate, Mutating: true, MinVer: "5.5.19",
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string","description":"RADIUS profile _id"},"config":{"type":"object","description":"RADIUS profile configuration"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "PUT", sp()+"/rest/radiusprofile/"+p.ID, p.Config)
			},
		},
		{
			ToolName: "system_radiusprofile_delete", ToolDesc: "Delete a RADIUS profile",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionDelete, Mutating: true, MinVer: "5.5.19",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "DELETE", sp()+"/rest/radiusprofile/"+p.ID, nil)
			},
		},
		{
			ToolName: "system_radius_account_create", ToolDesc: "Create a RADIUS account",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionCreate, Mutating: true, MinVer: "5.5.19",
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"RADIUS account configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/rest/account", p.Config)
			},
		},
		{
			ToolName: "system_radius_account_update", ToolDesc: "Update a RADIUS account",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionUpdate, Mutating: true, MinVer: "5.5.19",
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string","description":"RADIUS account _id"},"config":{"type":"object","description":"RADIUS account configuration"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "PUT", sp()+"/rest/account/"+p.ID, p.Config)
			},
		},
		{
			ToolName: "system_radius_account_delete", ToolDesc: "Delete a RADIUS account",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionDelete, Mutating: true, MinVer: "5.5.19",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "DELETE", sp()+"/rest/account/"+p.ID, nil)
			},
		},
		{
			ToolName: "system_tag_list", ToolDesc: "List device tags",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionRead, MinVer: "5.5.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/tag", nil)
			},
		},
		{
			ToolName: "system_tag_create", ToolDesc: "Create a device tag",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionCreate, Mutating: true, MinVer: "5.5.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"Tag configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Config map[string]interface{} `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/rest/tag", p.Config)
			},
		},
		{
			ToolName: "system_tag_update", ToolDesc: "Update a device tag",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionUpdate, Mutating: true, MinVer: "5.5.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string","description":"Tag _id"},"config":{"type":"object","description":"Tag configuration"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "PUT", sp()+"/rest/tag/"+p.ID, p.Config)
			},
		},
		{
			ToolName: "system_tag_delete", ToolDesc: "Delete a device tag",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionDelete, Mutating: true, MinVer: "5.5.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "DELETE", sp()+"/rest/tag/"+p.ID, nil)
			},
		},
		{
			ToolName: "system_backup_delete", ToolDesc: "Delete a backup file",
			ToolCategory: permissions.CatBackup, ToolAction: permissions.ActionDelete, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"filename":{"type":"string","description":"Backup filename to delete"}},"required":["filename"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Filename string `json:"filename"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/cmd/backup", map[string]interface{}{
					"cmd": "delete-backup", "filename": p.Filename,
				})
			},
		},
		{
			ToolName: "system_network_references", ToolDesc: "Get references to a network (which devices/clients use it)",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"network_id":{"type":"string","description":"Network _id"}},"required":["network_id"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					NetworkID string `json:"network_id"`
				}
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/networks/%s/references", base, c.Site(), p.NetworkID), nil)
			},
		},
		{
			ToolName: "system_device_tags", ToolDesc: "List device tags (Integration API)",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/device-tags", base, c.Site()), nil)
			},
		},
		{
			ToolName: "system_radius_profiles_v2", ToolDesc: "List RADIUS profiles (Integration API)",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/radius/profiles", base, c.Site()), nil)
			},
		},
	}
}
