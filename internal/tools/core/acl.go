package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliverames/ames-unifi-mcp/internal/client"
	"github.com/oliverames/ames-unifi-mcp/internal/permissions"
)

func BuildACLTools(c *client.Client) []*BaseTool {

	return []*BaseTool{
		{
			ToolName: "acl_rule_list", ToolDesc: "List all ACL rules (Network 10.0+)",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionRead, MinVer: "10.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoIntegrationSite(ctx, "GET", "/acl-rules", nil)
			},
		},
		{
			ToolName: "acl_rule_create", ToolDesc: "Create a new ACL rule (Network 10.0+)",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionCreate, Mutating: true, MinVer: "10.0.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"ACL rule configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Config json.RawMessage `json:"config"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.DoIntegrationSite(ctx, "POST", "/acl-rules", p.Config)
			},
		},
		{
			ToolName: "acl_rule_get", ToolDesc: "Get an ACL rule by ID (Network 10.0+)",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionRead, MinVer: "10.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.DoIntegrationSite(ctx, "GET", fmt.Sprintf("/acl-rules/%s", p.ID), nil)
			},
		},
		{
			ToolName: "acl_rule_update", ToolDesc: "Update an ACL rule by ID (Network 10.0+)",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionUpdate, Mutating: true, MinVer: "10.0.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"},"config":{"type":"object"}},"required":["id","config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID     string          `json:"id"`
					Config json.RawMessage `json:"config"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.DoIntegrationSite(ctx, "PUT", fmt.Sprintf("/acl-rules/%s", p.ID), p.Config)
			},
		},
		{
			ToolName: "acl_rule_delete", ToolDesc: "Delete an ACL rule by ID (Network 10.0+)",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionDelete, Mutating: true, MinVer: "10.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.DoIntegrationSite(ctx, "DELETE", fmt.Sprintf("/acl-rules/%s", p.ID), nil)
			},
		},
		{
			ToolName: "acl_rule_ordering_get", ToolDesc: "Get ACL rule ordering (Network 10.0+)",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionRead, MinVer: "10.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoIntegrationSite(ctx, "GET", "/acl-rules/ordering", nil)
			},
		},
		{
			ToolName: "acl_rule_ordering_set", ToolDesc: "Set ACL rule ordering (Network 10.0+)",
			ToolCategory: permissions.CatFirewall, ToolAction: permissions.ActionUpdate, Mutating: true, MinVer: "10.0.0",
			Schema: json.RawMessage(`{"type":"object","properties":{"config":{"type":"object","description":"ACL rule ordering configuration"}},"required":["config"]}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Config json.RawMessage `json:"config"`
				}
				if err := json.Unmarshal(input, &p); err != nil {
					return nil, fmt.Errorf("parsing input: %w", err)
				}
				return c.DoIntegrationSite(ctx, "PUT", "/acl-rules/ordering", p.Config)
			},
		},
	}
}
