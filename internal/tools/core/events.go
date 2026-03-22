package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliveames/ames-unifi-mcp/internal/client"
	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
)

func BuildEventTools(c *client.Client) []*BaseTool {
	sp := func() string { return fmt.Sprintf("api/s/%s", c.Site()) }

	return []*BaseTool{
		{
			ToolName: "event_list", ToolDesc: "List recent events (newest first, max 3000)",
			ToolCategory: permissions.CatEvents, ToolAction: permissions.ActionRead,
			Schema: json.RawMessage(`{"type":"object","properties":{"within":{"type":"integer","description":"Hours of history (default 720)","default":720},"limit":{"type":"integer","description":"Max events (default 100)","default":100}}}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Within int `json:"within"`
					Limit  int `json:"limit"`
				}
				json.Unmarshal(input, &p)
				if p.Within == 0 {
					p.Within = 720
				}
				if p.Limit == 0 {
					p.Limit = 100
				}
				return c.Do(ctx, "POST", sp()+"/stat/event", map[string]interface{}{
					"_sort": "-time", "within": p.Within, "_start": 0, "_limit": p.Limit,
				})
			},
		},
		{
			ToolName: "alarm_list", ToolDesc: "List alarms. Set archived=true to include archived.",
			ToolCategory: permissions.CatEvents, ToolAction: permissions.ActionRead,
			Schema: json.RawMessage(`{"type":"object","properties":{"archived":{"type":"boolean","description":"Include archived (default false)","default":false}}}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Archived *bool `json:"archived"` }
				json.Unmarshal(input, &p)
				path := sp() + "/rest/alarm"
				if p.Archived == nil || !*p.Archived {
					path += "?archived=false"
				}
				return c.Do(ctx, "GET", path, nil)
			},
		},
		{
			ToolName: "alarm_count", ToolDesc: "Count active (unarchived) alarms",
			ToolCategory: permissions.CatEvents, ToolAction: permissions.ActionRead,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "GET", sp()+"/cnt/alarm?archived=false", nil)
			},
		},
		{
			ToolName: "alarm_archive", ToolDesc: "Archive a single alarm by ID",
			ToolCategory: permissions.CatEvents, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: idSchema(), Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ ID string `json:"id"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/cmd/evtmgr", map[string]interface{}{"cmd": "archive-alarm", "_id": p.ID})
			},
		},
		{
			ToolName: "alarm_archive_all", ToolDesc: "Archive all active alarms",
			ToolCategory: permissions.CatEvents, ToolAction: permissions.ActionUpdate, Mutating: true,
			Schema: noInputSchema(), Client: c,
			Handler: func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return c.Do(ctx, "POST", sp()+"/cmd/evtmgr", map[string]interface{}{"cmd": "archive-all-alarms"})
			},
		},
	}
}
