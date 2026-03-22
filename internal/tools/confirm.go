package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
)

// ConfirmGate wraps a mutating tool with the confirm=true safety pattern.
// If the input does not include "confirm": true, it returns a dry-run preview
// instead of executing. This prevents accidental mutations.
type ConfirmGate struct {
	inner Tool
}

func WithConfirm(t Tool) Tool {
	if !t.IsMutating() {
		return t // no-op for read-only tools
	}
	return &ConfirmGate{inner: t}
}

func (g *ConfirmGate) Name() string              { return g.inner.Name() }
func (g *ConfirmGate) Category() permissions.Category { return g.inner.Category() }
func (g *ConfirmGate) Action() permissions.Action { return g.inner.Action() }
func (g *ConfirmGate) IsMutating() bool           { return true }
func (g *ConfirmGate) MinVersion() string         { return g.inner.MinVersion() }
func (g *ConfirmGate) IsUndocumented() bool       { return g.inner.IsUndocumented() }

func (g *ConfirmGate) Description() string {
	return g.inner.Description() + " (requires confirm=true to execute; omit for dry-run preview)"
}

func (g *ConfirmGate) InputSchema() json.RawMessage {
	// Inject "confirm" property into the existing schema
	var schema map[string]interface{}
	if err := json.Unmarshal(g.inner.InputSchema(), &schema); err != nil {
		return g.inner.InputSchema()
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		props = make(map[string]interface{})
	}
	props["confirm"] = map[string]interface{}{
		"type":        "boolean",
		"description": "Set to true to execute this operation. Omit or set false for a dry-run preview.",
		"default":     false,
	}
	schema["properties"] = props

	out, _ := json.Marshal(schema)
	return out
}

func (g *ConfirmGate) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("parsing input: %w", err)
	}

	confirmed, _ := params["confirm"].(bool)
	if !confirmed {
		preview := map[string]interface{}{
			"tool":                  g.inner.Name(),
			"action":                "dry-run preview",
			"description":           g.inner.Description(),
			"requires_confirmation": true,
			"message":               fmt.Sprintf("This operation would execute %s. Set confirm=true to proceed.", g.inner.Name()),
			"parameters":            params,
		}
		return json.Marshal(preview)
	}

	// Remove the confirm key before passing to the inner tool
	delete(params, "confirm")
	cleaned, _ := json.Marshal(params)
	return g.inner.Execute(ctx, cleaned)
}
