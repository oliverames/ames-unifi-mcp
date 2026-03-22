package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliveames/ames-unifi-mcp/internal/client"
	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
)

// HandlerFunc is a function that executes a tool.
type HandlerFunc func(ctx context.Context, input json.RawMessage) (json.RawMessage, error)

// BaseTool provides common fields and methods for tool implementations.
type BaseTool struct {
	ToolName     string
	ToolDesc     string
	ToolCategory permissions.Category
	ToolAction   permissions.Action
	Schema       json.RawMessage
	Mutating     bool
	MinVer       string
	Undocumented bool
	Client       *client.Client
	Handler      HandlerFunc
}

func (b *BaseTool) Name() string                   { return b.ToolName }
func (b *BaseTool) Description() string            { return b.ToolDesc }
func (b *BaseTool) Category() permissions.Category { return b.ToolCategory }
func (b *BaseTool) Action() permissions.Action     { return b.ToolAction }
func (b *BaseTool) InputSchema() json.RawMessage   { return b.Schema }
func (b *BaseTool) IsMutating() bool               { return b.Mutating }
func (b *BaseTool) MinVersion() string             { return b.MinVer }
func (b *BaseTool) IsUndocumented() bool           { return b.Undocumented }

func (b *BaseTool) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	if b.Handler == nil {
		return nil, fmt.Errorf("no handler for tool %s", b.ToolName)
	}
	return b.Handler(ctx, input)
}

// noInputSchema returns a schema with no required parameters.
func noInputSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{}}`)
}

// macSchema returns a schema requiring a MAC address parameter.
func macSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"mac": {
				"type": "string",
				"description": "Device MAC address (lowercase, colon-separated)"
			}
		},
		"required": ["mac"]
	}`)
}

// idSchema returns a schema requiring an _id parameter.
func idSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"id": {
				"type": "string",
				"description": "The resource _id"
			}
		},
		"required": ["id"]
	}`)
}

// NoInputSchema is the exported version for use in extended/ package.
func NoInputSchema() json.RawMessage { return noInputSchema() }

// MacSchema is the exported version for use in extended/ package.
func MacSchema() json.RawMessage { return macSchema() }

// IDSchema is the exported version for use in extended/ package.
func IDSchema() json.RawMessage { return idSchema() }
