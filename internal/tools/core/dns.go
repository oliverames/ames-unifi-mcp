package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliveames/ames-unifi-mcp/internal/client"
	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
)

func BuildDNSTools(c *client.Client) []*BaseTool {
	base := c.Config().BaseURL() + "/integration"

	return []*BaseTool{
		{
			ToolName: "dns_policy_list", ToolDesc: "List all DNS policies (Network 10.0+)",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionRead, MinVer: "10.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/dns/policies", base, c.Site()), nil)
			},
		},
		{
			ToolName: "dns_policy_create", ToolDesc: "Create a new DNS policy (Network 10.0+)",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionCreate, Mutating: true, MinVer: "10.0.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"DNS policy configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "POST", fmt.Sprintf("%s/v1/sites/%s/dns/policies", base, c.Site()), p.Config)
			},
		},
		{
			ToolName: "dns_policy_get", ToolDesc: "Get a DNS policy by ID (Network 10.0+)",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionRead, MinVer: "10.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/dns/policies/%s", base, c.Site(), p.ID), nil)
			},
		},
		{
			ToolName: "dns_policy_update", ToolDesc: "Update a DNS policy by ID (Network 10.0+)",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionUpdate, Mutating: true, MinVer: "10.0.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"},"config":{"type":"object"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "PUT", fmt.Sprintf("%s/v1/sites/%s/dns/policies/%s", base, c.Site(), p.ID), p.Config)
			},
		},
		{
			ToolName: "dns_policy_delete", ToolDesc: "Delete a DNS policy by ID (Network 10.0+)",
			ToolCategory: permissions.CatNetworks, ToolAction: permissions.ActionDelete, Mutating: true, MinVer: "10.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "DELETE", fmt.Sprintf("%s/v1/sites/%s/dns/policies/%s", base, c.Site(), p.ID), nil)
			},
		},
	}
}
