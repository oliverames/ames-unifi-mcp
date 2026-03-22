package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliveames/ames-unifi-mcp/internal/client"
	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
)

func BuildWiFiTools(c *client.Client) []*BaseTool {
	base := c.Config().BaseURL() + "/integration"

	return []*BaseTool{
		{
			ToolName: "wifi_broadcast_list", ToolDesc: "List WiFi broadcasts/SSIDs (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatWLAN, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/wifi/broadcasts", base, c.Site()), nil)
			},
		},
		{
			ToolName: "wifi_broadcast_get", ToolDesc: "Get WiFi broadcast details by ID (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatWLAN, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/wifi/broadcasts/%s", base, c.Site(), p.ID), nil)
			},
		},
		{
			ToolName: "wifi_broadcast_create", ToolDesc: "Create a WiFi broadcast (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatWLAN, ToolAction: permissions.ActionCreate, Mutating: true, MinVer: "9.0.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"WiFi broadcast configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "POST", fmt.Sprintf("%s/v1/sites/%s/wifi/broadcasts", base, c.Site()), p.Config)
			},
		},
		{
			ToolName: "wifi_broadcast_update", ToolDesc: "Update a WiFi broadcast by ID (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatWLAN, ToolAction: permissions.ActionUpdate, Mutating: true, MinVer: "9.0.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string","description":"Broadcast ID"},"config":{"type":"object","description":"Updated broadcast configuration"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "PUT", fmt.Sprintf("%s/v1/sites/%s/wifi/broadcasts/%s", base, c.Site(), p.ID), p.Config)
			},
		},
		{
			ToolName: "wifi_broadcast_delete", ToolDesc: "Delete a WiFi broadcast by ID (Integration API, Network 9.0+)",
			ToolCategory: permissions.CatWLAN, ToolAction: permissions.ActionDelete, Mutating: true, MinVer: "9.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "DELETE", fmt.Sprintf("%s/v1/sites/%s/wifi/broadcasts/%s", base, c.Site(), p.ID), nil)
			},
		},
	}
}
