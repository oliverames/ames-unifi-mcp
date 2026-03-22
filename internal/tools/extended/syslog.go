package extended

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliveames/ames-unifi-mcp/internal/client"
	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
	"github.com/oliveames/ames-unifi-mcp/internal/tools/core"
)

func BuildSyslogTools(c *client.Client) []*core.BaseTool {
	return []*core.BaseTool{
		{
			ToolName: "syslog_query", ToolDesc: "Query system logs by class (device-alert, admin-activity, threat-alert, etc.)",
			ToolCategory: permissions.CatEvents, ToolAction: permissions.ActionRead,
			MinVer: "6.0.0",
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"class": {"type": "string", "enum": ["device-alert","next-ai-alert","vpn-alert","admin-activity","update-alert","client-alert","threat-alert","triggers"], "description": "Log class to query"},
					"config": {"type": "object", "description": "Optional filter parameters"}
				},
				"required": ["class"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Class  string          `json:"class"`
					Config json.RawMessage `json:"config"`
				}
				json.Unmarshal(input, &p)
				path := fmt.Sprintf("v2/api/site/%s/system-log/%s", c.Site(), p.Class)
				var payload interface{}
				if len(p.Config) > 0 && string(p.Config) != "null" {
					payload = p.Config
				}
				return c.Do(ctx, "POST", path, payload)
			},
		},
		{
			ToolName: "syslog_fingerprints", ToolDesc: "Get client device fingerprint data [undocumented]",
			ToolCategory: permissions.CatClients, ToolAction: permissions.ActionRead,
			Undocumented: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"source": {"type": "string", "description": "Fingerprint source identifier"}
				},
				"required": ["source"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct{ Source string `json:"source"` }
				json.Unmarshal(input, &p)
				return c.Do(ctx, "GET", "v2/api/fingerprint_devices/"+p.Source, nil)
			},
		},
	}
}
