package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliveames/ames-unifi-mcp/internal/client"
	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
)

func BuildTrafficTools(c *client.Client) []*BaseTool {
	base := c.Config().BaseURL() + "/integration"

	return []*BaseTool{
		// v2 API traffic rules
		{
			ToolName: "traffic_rule_list", ToolDesc: "List all traffic rules (v2 API)",
			ToolCategory: permissions.CatQoS, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", fmt.Sprintf("v2/api/site/%s/trafficrules", c.Site()), nil)
			},
		},
		{
			ToolName: "traffic_rule_create", ToolDesc: "Create a traffic rule (v2 API)",
			ToolCategory: permissions.CatQoS, ToolAction: permissions.ActionCreate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"Traffic rule configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", fmt.Sprintf("v2/api/site/%s/trafficrules", c.Site()), p.Config)
			},
		},
		{
			ToolName: "traffic_rule_get", ToolDesc: "Get a traffic rule by ID (v2 API)",
			ToolCategory: permissions.CatQoS, ToolAction: permissions.ActionRead,
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "GET", fmt.Sprintf("v2/api/site/%s/trafficrules/%s", c.Site(), p.ID), nil)
			},
		},
		{
			ToolName: "traffic_rule_update", ToolDesc: "Update a traffic rule by ID (v2 API)",
			ToolCategory: permissions.CatQoS, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"},"config":{"type":"object"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "PUT", fmt.Sprintf("v2/api/site/%s/trafficrules/%s", c.Site(), p.ID), p.Config)
			},
		},
		{
			ToolName: "traffic_rule_delete", ToolDesc: "Delete a traffic rule by ID (v2 API)",
			ToolCategory: permissions.CatQoS, ToolAction: permissions.ActionDelete, Mutating: true,
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "DELETE", fmt.Sprintf("v2/api/site/%s/trafficrules/%s", c.Site(), p.ID), nil)
			},
		},

		// v2 API traffic routes (policy-based routing)
		{
			ToolName: "traffic_route_list", ToolDesc: "List all traffic/policy-based routes (v2 API)",
			ToolCategory: permissions.CatRouting, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", fmt.Sprintf("v2/api/site/%s/trafficroutes", c.Site()), nil)
			},
		},
		{
			ToolName: "traffic_route_get", ToolDesc: "Get a traffic route by ID (v2 API)",
			ToolCategory: permissions.CatRouting, ToolAction: permissions.ActionRead,
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "GET", fmt.Sprintf("v2/api/site/%s/trafficroutes/%s", c.Site(), p.ID), nil)
			},
		},
		{
			ToolName: "traffic_route_create", ToolDesc: "Create a traffic/policy-based route (v2 API)",
			ToolCategory: permissions.CatRouting, ToolAction: permissions.ActionCreate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"Traffic route configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", fmt.Sprintf("v2/api/site/%s/trafficroutes", c.Site()), p.Config)
			},
		},
		{
			ToolName: "traffic_route_update", ToolDesc: "Update a traffic route by ID (v2 API)",
			ToolCategory: permissions.CatRouting, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"},"config":{"type":"object"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "PUT", fmt.Sprintf("v2/api/site/%s/trafficroutes/%s", c.Site(), p.ID), p.Config)
			},
		},
		{
			ToolName: "traffic_route_delete", ToolDesc: "Delete a traffic route by ID (v2 API)",
			ToolCategory: permissions.CatRouting, ToolAction: permissions.ActionDelete, Mutating: true,
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "DELETE", fmt.Sprintf("v2/api/site/%s/trafficroutes/%s", c.Site(), p.ID), nil)
			},
		},

		// Integration API traffic matching lists
		{
			ToolName: "traffic_matching_list_list", ToolDesc: "List traffic matching lists (Network 10.0+)",
			ToolCategory: permissions.CatQoS, ToolAction: permissions.ActionRead, MinVer: "10.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/traffic-matching-lists", base, c.Site()), nil)
			},
		},
		{
			ToolName: "traffic_matching_list_create", ToolDesc: "Create a traffic matching list (Network 10.0+)",
			ToolCategory: permissions.CatQoS, ToolAction: permissions.ActionCreate, Mutating: true, MinVer: "10.0.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"Traffic matching list configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "POST", fmt.Sprintf("%s/v1/sites/%s/traffic-matching-lists", base, c.Site()), p.Config)
			},
		},
		{
			ToolName: "traffic_matching_list_get", ToolDesc: "Get a traffic matching list by ID (Network 10.0+)",
			ToolCategory: permissions.CatQoS, ToolAction: permissions.ActionRead, MinVer: "10.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/traffic-matching-lists/%s", base, c.Site(), p.ID), nil)
			},
		},
		{
			ToolName: "traffic_matching_list_update", ToolDesc: "Update a traffic matching list by ID (Network 10.0+)",
			ToolCategory: permissions.CatQoS, ToolAction: permissions.ActionUpdate, Mutating: true, MinVer: "10.0.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"},"config":{"type":"object"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "PUT", fmt.Sprintf("%s/v1/sites/%s/traffic-matching-lists/%s", base, c.Site(), p.ID), p.Config)
			},
		},
		{
			ToolName: "traffic_matching_list_delete", ToolDesc: "Delete a traffic matching list by ID (Network 10.0+)",
			ToolCategory: permissions.CatQoS, ToolAction: permissions.ActionDelete, Mutating: true, MinVer: "10.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "DELETE", fmt.Sprintf("%s/v1/sites/%s/traffic-matching-lists/%s", base, c.Site(), p.ID), nil)
			},
		},
	}
}
