package extended

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliveames/ames-unifi-mcp/internal/client"
	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
	"github.com/oliveames/ames-unifi-mcp/internal/tools/core"
)

func BuildPoETools(c *client.Client) []*core.BaseTool {
	sp := func() string { return fmt.Sprintf("api/s/%s", c.Site()) }

	return []*core.BaseTool{
		{
			ToolName:     "poe_power_cycle",
			ToolDesc:     "Power-cycle a PoE switch port (remotely reboot PoE-powered devices) [undocumented]",
			ToolCategory: permissions.CatPoE,
			ToolAction:   permissions.ActionExecute,
			Mutating:     true,
			Undocumented: true,
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"mac": {"type": "string", "description": "Switch MAC address"},
					"port_idx": {"type": "integer", "description": "Port index (1-based)"}
				},
				"required": ["mac", "port_idx"]
			}`),
			Client: c,
			Handler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
				var p struct {
					Mac     string `json:"mac"`
					PortIdx int    `json:"port_idx"`
				}
				json.Unmarshal(input, &p)
				return c.Do(ctx, "POST", sp()+"/cmd/devmgr", map[string]interface{}{
					"cmd": "power-cycle", "mac": p.Mac, "port_idx": p.PortIdx,
				})
			},
		},
	}
}
