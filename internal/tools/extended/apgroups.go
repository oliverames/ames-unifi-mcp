package extended

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliveames/ames-unifi-mcp/internal/client"
	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
	"github.com/oliveames/ames-unifi-mcp/internal/tools/core"
)

func BuildAPGroupTools(c *client.Client) []*core.BaseTool {
	ap := func() string { return fmt.Sprintf("v2/api/site/%s/apgroups", c.Site()) }

	return []*core.BaseTool{
		{
			ToolName: "apgroup_list", ToolDesc: "List AP groups",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionRead,
			MinVer: "6.0.0",
			Schema: core.NoInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", ap(), nil)
			},
		},
		{
			ToolName: "apgroup_get", ToolDesc: "Get an AP group by ID",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionRead,
			MinVer: "6.0.0",
			Schema: core.IDSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "GET", ap()+"/"+p.ID, nil)
			},
		},
		{
			ToolName: "apgroup_create", ToolDesc: "Create an AP group",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionCreate,
			MinVer: "6.0.0", Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"config": {"type": "object", "description": "AP group configuration"}
				},
				"required": ["config"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", ap(), p.Config)
			},
		},
		{
			ToolName: "apgroup_update", ToolDesc: "Update an AP group",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionUpdate,
			MinVer: "6.0.0", Mutating: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"id": {"type": "string", "description": "AP group _id"},
					"config": {"type": "object", "description": "Updated AP group configuration"}
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
				return c.Do(ctx, "PUT", ap()+"/"+p.ID, p.Config)
			},
		},
		{
			ToolName: "apgroup_delete", ToolDesc: "Delete an AP group",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionDelete,
			MinVer: "6.0.0", Mutating: true,
			Schema: core.IDSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "DELETE", ap()+"/"+p.ID, nil)
			},
		},
	}
}
