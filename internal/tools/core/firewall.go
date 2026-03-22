package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliveames/ames-unifi-mcp/internal/client"
	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
)

func BuildFirewallLegacyTools(c *client.Client) []*BaseTool {
	sp := func() string { return fmt.Sprintf("api/s/%s", c.Site()) }

	return []*BaseTool{
		{
			ToolName: "firewall_rule_list", ToolDesc: "List all legacy firewall rules",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/firewallrule", nil)
			},
		},
		{
			ToolName: "firewall_group_list", ToolDesc: "List all firewall groups (address/port groups)",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/rest/firewallgroup", nil)
			},
		},
		{
			ToolName: "firewall_rule_get", ToolDesc: "Get a specific legacy firewall rule by ID",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionRead,
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "GET", sp()+"/rest/firewallrule/"+p.ID, nil)
			},
		},
		{
			ToolName: "firewall_group_get", ToolDesc: "Get a specific firewall group by ID",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionRead,
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "GET", sp()+"/rest/firewallgroup/"+p.ID, nil)
			},
		},
		{
			ToolName: "firewall_rule_create", ToolDesc: "Create a new legacy firewall rule",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionCreate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"Firewall rule config"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/rest/firewallrule", p.Config)
			},
		},
		{
			ToolName: "firewall_rule_update", ToolDesc: "Update a legacy firewall rule by ID",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string","description":"The firewall rule _id"},"config":{"type":"object","description":"Firewall rule config"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "PUT", sp()+"/rest/firewallrule/"+p.ID, p.Config)
			},
		},
		{
			ToolName: "firewall_rule_delete", ToolDesc: "Delete a legacy firewall rule by ID",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionDelete, Mutating: true,
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "DELETE", sp()+"/rest/firewallrule/"+p.ID, nil)
			},
		},
		{
			ToolName: "firewall_group_create", ToolDesc: "Create a new firewall group",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionCreate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"Firewall group config (name, group_type, group_members)"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/rest/firewallgroup", p.Config)
			},
		},
		{
			ToolName: "firewall_group_update", ToolDesc: "Update a firewall group by ID",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string","description":"The firewall group _id"},"config":{"type":"object","description":"Firewall group config"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "PUT", sp()+"/rest/firewallgroup/"+p.ID, p.Config)
			},
		},
		{
			ToolName: "firewall_group_delete", ToolDesc: "Delete a firewall group by ID",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionDelete, Mutating: true,
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "DELETE", sp()+"/rest/firewallgroup/"+p.ID, nil)
			},
		},
	}
}

func BuildFirewallZBFTools(c *client.Client) []*BaseTool {
	base := c.Config().BaseURL() + "/integration"

	return []*BaseTool{
		{
			ToolName: "firewall_zone_list", ToolDesc: "List firewall zones (Zone-Based Firewall, Network 9.0+)",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/firewall/zones", base, c.Site()), nil)
			},
		},
		{
			ToolName: "firewall_zone_create", ToolDesc: "Create a firewall zone (ZBF, Network 9.0+)",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionCreate, Mutating: true, MinVer: "9.0.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"Firewall zone configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "POST", fmt.Sprintf("%s/v1/sites/%s/firewall/zones", base, c.Site()), p.Config)
			},
		},
		{
			ToolName: "firewall_zone_get", ToolDesc: "Get a firewall zone by ID (ZBF, Network 9.0+)",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/firewall/zones/%s", base, c.Site(), p.ID), nil)
			},
		},
		{
			ToolName: "firewall_zone_update", ToolDesc: "Update a firewall zone by ID (ZBF, Network 9.0+)",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionUpdate, Mutating: true, MinVer: "9.0.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"},"config":{"type":"object"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "PUT", fmt.Sprintf("%s/v1/sites/%s/firewall/zones/%s", base, c.Site(), p.ID), p.Config)
			},
		},
		{
			ToolName: "firewall_zone_delete", ToolDesc: "Delete a firewall zone by ID (ZBF, Network 9.0+)",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionDelete, Mutating: true, MinVer: "9.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "DELETE", fmt.Sprintf("%s/v1/sites/%s/firewall/zones/%s", base, c.Site(), p.ID), nil)
			},
		},
		{
			ToolName: "firewall_policy_list", ToolDesc: "List firewall policies (Zone-Based Firewall, Network 9.0+)",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/firewall/policies", base, c.Site()), nil)
			},
		},
		{
			ToolName: "firewall_policy_get", ToolDesc: "Get a firewall policy by ID (ZBF, Network 9.0+)",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/firewall/policies/%s", base, c.Site(), p.ID), nil)
			},
		},
		{
			ToolName: "firewall_policy_create", ToolDesc: "Create a firewall policy (ZBF, Network 9.0+)",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionCreate, Mutating: true, MinVer: "9.0.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"Firewall policy configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "POST", fmt.Sprintf("%s/v1/sites/%s/firewall/policies", base, c.Site()), p.Config)
			},
		},
		{
			ToolName: "firewall_policy_update", ToolDesc: "Update a firewall policy by ID (ZBF, Network 9.0+)",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionUpdate, Mutating: true, MinVer: "9.0.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"},"config":{"type":"object"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "PUT", fmt.Sprintf("%s/v1/sites/%s/firewall/policies/%s", base, c.Site(), p.ID), p.Config)
			},
		},
		{
			ToolName: "firewall_policy_delete", ToolDesc: "Delete a firewall policy (ZBF, Network 9.0+)",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionDelete, Mutating: true, MinVer: "9.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "DELETE", fmt.Sprintf("%s/v1/sites/%s/firewall/policies/%s", base, c.Site(), p.ID), nil)
			},
		},
		{
			ToolName: "firewall_policy_patch", ToolDesc: "Partially update a firewall policy by ID (PATCH, ZBF, Network 9.0+)",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionUpdate, Mutating: true, MinVer: "9.0.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"},"config":{"type":"object"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "PATCH", fmt.Sprintf("%s/v1/sites/%s/firewall/policies/%s", base, c.Site(), p.ID), p.Config)
			},
		},
		{
			ToolName: "firewall_policy_ordering_get", ToolDesc: "Get firewall policy ordering (ZBF, Network 9.0+)",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionRead, MinVer: "9.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/firewall/policies/ordering", base, c.Site()), nil)
			},
		},
		{
			ToolName: "firewall_policy_ordering_set", ToolDesc: "Set firewall policy ordering (ZBF, Network 9.0+)",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionUpdate, Mutating: true, MinVer: "9.0.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"Policy ordering configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Config json.RawMessage `json:"config"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "PUT", fmt.Sprintf("%s/v1/sites/%s/firewall/policies/ordering", base, c.Site()), p.Config)
			},
		},
	}
}
