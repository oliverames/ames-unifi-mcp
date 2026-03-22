package extended

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliveames/ames-unifi-mcp/internal/client"
	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
	"github.com/oliveames/ames-unifi-mcp/internal/tools/core"
)

func BuildAdminTools(c *client.Client) []*core.BaseTool {
	sp := func() string { return fmt.Sprintf("api/s/%s", c.Site()) }

	return []*core.BaseTool{
		{
			ToolName: "admin_site_create", ToolDesc: "Create a new site",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionCreate, Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"desc": {"type": "string", "description": "Site description/display name"},
					"name": {"type": "string", "description": "Site short name (optional)"}
				},
				"required": ["desc"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Desc string `json:"desc"`
					Name string `json:"name"`
				}
				json.Unmarshal(input, &p)
				payload := map[string]interface{}{"cmd": "add-site", "desc": p.Desc}
				if p.Name != "" {
					payload["name"] = p.Name
				}
				return c.Do(ctx, "POST", sp()+"/cmd/sitemgr", payload)
			},
		},
		{
			ToolName: "admin_site_delete", ToolDesc: "Delete a site",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionDelete, Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"site_id": {"type": "string", "description": "Site _id to delete"}
				},
				"required": ["site_id"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ SiteID string `json:"site_id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/cmd/sitemgr", map[string]interface{}{
					"cmd": "delete-site", "site": p.SiteID,
				})
			},
		},
		{
			ToolName: "admin_site_rename", ToolDesc: "Rename the current site",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"desc": {"type": "string", "description": "New site description/name"}
				},
				"required": ["desc"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Desc string `json:"desc"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/cmd/sitemgr", map[string]interface{}{
					"cmd": "update-site", "desc": p.Desc,
				})
			},
		},
		{
			ToolName: "admin_device_move", ToolDesc: "Move a device to another site",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"mac": {"type": "string", "description": "Device MAC address"},
					"site_id": {"type": "string", "description": "Target site _id"}
				},
				"required": ["mac", "site_id"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Mac    string `json:"mac"`
					SiteID string `json:"site_id"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/cmd/sitemgr", map[string]interface{}{
					"cmd": "move-device", "mac": p.Mac, "site": p.SiteID,
				})
			},
		},
		{
			ToolName: "admin_invite", ToolDesc: "Invite a new admin to this site",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionCreate, Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"name": {"type": "string", "description": "Admin display name"},
					"email": {"type": "string", "description": "Admin email address"},
					"role": {"type": "string", "description": "Role (default: admin)", "default": "admin"}
				},
				"required": ["name", "email"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Name  string `json:"name"`
					Email string `json:"email"`
					Role  string `json:"role"`
				}
				json.Unmarshal(input, &p)
				if p.Role == "" {
					p.Role = "admin"
				}
				return c.Do(ctx, "POST", sp()+"/cmd/sitemgr", map[string]interface{}{
					"cmd": "invite-admin", "name": p.Name, "email": p.Email, "role": p.Role,
				})
			},
		},
		{
			ToolName: "admin_revoke", ToolDesc: "Revoke admin access from this site",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionDelete, Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"admin_id": {"type": "string", "description": "Admin _id to revoke"}
				},
				"required": ["admin_id"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ AdminID string `json:"admin_id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/cmd/sitemgr", map[string]interface{}{
					"cmd": "revoke-admin", "admin": p.AdminID,
				})
			},
		},
		{
			ToolName: "admin_list", ToolDesc: "List all administrators and their permissions across sites",
			ToolCategory: permissions.CatSystem, ToolAction: permissions.ActionRead,
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", "api/stat/admin", nil)
			},
		},
	}
}
