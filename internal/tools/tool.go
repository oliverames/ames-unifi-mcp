package tools

import (
	"context"
	"encoding/json"

	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
)

// Tool defines the interface every tool must implement.
type Tool interface {
	// Name returns the tool's unique identifier (e.g., "device_list").
	Name() string

	// Description returns a human-readable description for the LLM.
	Description() string

	// Category returns the tool's resource domain.
	Category() permissions.Category

	// Action returns what kind of operation this tool performs.
	Action() permissions.Action

	// InputSchema returns the JSON Schema for the tool's input parameters.
	InputSchema() json.RawMessage

	// Execute runs the tool with the given input.
	Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error)

	// IsMutating returns true if this tool modifies state (triggers confirm gate).
	IsMutating() bool

	// MinVersion returns the minimum controller version required, or empty string if none.
	MinVersion() string

	// IsUndocumented returns true if this tool uses an undocumented endpoint.
	IsUndocumented() bool
}

// ToolMeta holds metadata for the tool index.
type ToolMeta struct {
	Name           string              `json:"name"`
	Description    string              `json:"description"`
	Category       permissions.Category `json:"category"`
	Mutating       bool                `json:"mutating"`
	MinVersion     string              `json:"min_version,omitempty"`
	Undocumented   bool                `json:"undocumented,omitempty"`
}

// MetaFromTool extracts index metadata from a Tool.
func MetaFromTool(t Tool) ToolMeta {
	return ToolMeta{
		Name:         t.Name(),
		Description:  t.Description(),
		Category:     t.Category(),
		Mutating:     t.IsMutating(),
		MinVersion:   t.MinVersion(),
		Undocumented: t.IsUndocumented(),
	}
}
