package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
)

// MetaToolIndex is the lazy-mode tool that returns the tool catalog.
type MetaToolIndex struct {
	registry *Registry
}

func NewMetaToolIndex(r *Registry) *MetaToolIndex {
	return &MetaToolIndex{registry: r}
}

func (m *MetaToolIndex) Name() string        { return "tool_index" }
func (m *MetaToolIndex) Description() string {
	return "List all available UniFi tools. Optionally filter by category (devices, clients, networks, wlan, firewall, vpn, routing, qos, stats, events, system, hotspot, poe, dpi, backup, settings)."
}
func (m *MetaToolIndex) Category() permissions.Category { return CatSystem }
func (m *MetaToolIndex) Action() permissions.Action     { return permissions.ActionRead }
func (m *MetaToolIndex) IsMutating() bool               { return false }
func (m *MetaToolIndex) MinVersion() string              { return "" }
func (m *MetaToolIndex) IsUndocumented() bool            { return false }

func (m *MetaToolIndex) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"category": {
				"type": "string",
				"description": "Filter by category name"
			}
		}
	}`)
}

func (m *MetaToolIndex) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Category string `json:"category"`
	}
	if len(input) > 0 {
		json.Unmarshal(input, &params)
	}
	index := m.registry.Index(params.Category)
	return json.Marshal(index)
}

// MetaToolExecute dispatches a tool call by name (lazy mode).
type MetaToolExecute struct {
	registry *Registry
}

func NewMetaToolExecute(r *Registry) *MetaToolExecute {
	return &MetaToolExecute{registry: r}
}

func (m *MetaToolExecute) Name() string        { return "tool_execute" }
func (m *MetaToolExecute) Description() string {
	return "Execute a UniFi tool by name. Use tool_index first to discover available tools and their input schemas."
}
func (m *MetaToolExecute) Category() permissions.Category { return CatSystem }
func (m *MetaToolExecute) Action() permissions.Action     { return permissions.ActionRead }
func (m *MetaToolExecute) IsMutating() bool               { return false }
func (m *MetaToolExecute) MinVersion() string              { return "" }
func (m *MetaToolExecute) IsUndocumented() bool            { return false }

func (m *MetaToolExecute) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"tool_name": {
				"type": "string",
				"description": "The name of the tool to execute"
			},
			"input": {
				"type": "object",
				"description": "The input parameters for the tool"
			}
		},
		"required": ["tool_name"]
	}`)
}

func (m *MetaToolExecute) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		ToolName string          `json:"tool_name"`
		Input    json.RawMessage `json:"input"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("parsing input: %w", err)
	}
	if params.Input == nil {
		params.Input = json.RawMessage(`{}`)
	}
	return m.registry.Execute(ctx, params.ToolName, params.Input)
}

// MetaToolBatch executes multiple tools in parallel (lazy mode).
type MetaToolBatch struct {
	registry *Registry
}

func NewMetaToolBatch(r *Registry) *MetaToolBatch {
	return &MetaToolBatch{registry: r}
}

func (m *MetaToolBatch) Name() string        { return "tool_batch" }
func (m *MetaToolBatch) Description() string {
	return "Execute multiple UniFi tools in parallel. Returns results for each tool call."
}
func (m *MetaToolBatch) Category() permissions.Category { return CatSystem }
func (m *MetaToolBatch) Action() permissions.Action     { return permissions.ActionRead }
func (m *MetaToolBatch) IsMutating() bool               { return false }
func (m *MetaToolBatch) MinVersion() string              { return "" }
func (m *MetaToolBatch) IsUndocumented() bool            { return false }

func (m *MetaToolBatch) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"calls": {
				"type": "array",
				"description": "Array of tool calls to execute in parallel",
				"items": {
					"type": "object",
					"properties": {
						"name": { "type": "string", "description": "Tool name" },
						"input": { "type": "object", "description": "Tool input" }
					},
					"required": ["name"]
				}
			}
		},
		"required": ["calls"]
	}`)
}

func (m *MetaToolBatch) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Calls []BatchCall `json:"calls"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("parsing input: %w", err)
	}
	results := m.registry.Batch(ctx, params.Calls)
	return json.Marshal(results)
}

// CatSystem is used for meta-tools that don't belong to a specific resource category.
const CatSystem permissions.Category = "system"
