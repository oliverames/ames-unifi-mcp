package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliveames/ames-unifi-mcp/internal/client"
	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
)

func BuildSwitchingTools(c *client.Client) []*BaseTool {
	base := c.Config().BaseURL() + "/integration"

	return []*BaseTool{
		{
			ToolName: "switching_stack_list", ToolDesc: "List switch stacks (Network 10.0+)",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionRead, MinVer: "10.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/switching/switch-stacks", base, c.Site()), nil)
			},
		},
		{
			ToolName: "switching_stack_get", ToolDesc: "Get a switch stack by ID (Network 10.0+)",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionRead, MinVer: "10.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/switching/switch-stacks/%s", base, c.Site(), p.ID), nil)
			},
		},
		{
			ToolName: "switching_lag_list", ToolDesc: "List LAGs (Link Aggregation Groups) (Network 10.0+)",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionRead, MinVer: "10.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/switching/lags", base, c.Site()), nil)
			},
		},
		{
			ToolName: "switching_lag_get", ToolDesc: "Get a LAG by ID (Network 10.0+)",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionRead, MinVer: "10.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/switching/lags/%s", base, c.Site(), p.ID), nil)
			},
		},
		{
			ToolName: "switching_mclag_list", ToolDesc: "List MC-LAG domains (Network 10.0+)",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionRead, MinVer: "10.0.0",
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/switching/mc-lag-domains", base, c.Site()), nil)
			},
		},
		{
			ToolName: "switching_mclag_get", ToolDesc: "Get an MC-LAG domain by ID (Network 10.0+)",
			ToolCategory: permissions.CatDevices, ToolAction: permissions.ActionRead, MinVer: "10.0.0",
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.DoRaw(ctx, "GET", fmt.Sprintf("%s/v1/sites/%s/switching/mc-lag-domains/%s", base, c.Site(), p.ID), nil)
			},
		},
	}
}
