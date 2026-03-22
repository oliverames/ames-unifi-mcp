package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
)

type mockTool struct {
	name     string
	mutating bool
	executed bool
}

func (m *mockTool) Name() string                    { return m.name }
func (m *mockTool) Description() string             { return "test tool" }
func (m *mockTool) Category() permissions.Category  { return permissions.CatDevices }
func (m *mockTool) Action() permissions.Action      { return permissions.ActionExecute }
func (m *mockTool) InputSchema() json.RawMessage    { return json.RawMessage(`{"type":"object","properties":{"mac":{"type":"string"}},"required":["mac"]}`) }
func (m *mockTool) IsMutating() bool                { return m.mutating }
func (m *mockTool) MinVersion() string              { return "" }
func (m *mockTool) IsUndocumented() bool            { return false }

func (m *mockTool) Execute(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
	m.executed = true
	return json.Marshal(map[string]string{"status": "ok"})
}

func TestConfirmGate_DryRun(t *testing.T) {
	inner := &mockTool{name: "test_restart", mutating: true}
	gated := WithConfirm(inner)

	input := json.RawMessage(`{"mac": "aa:bb:cc:dd:ee:ff"}`)
	result, err := gated.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var preview map[string]interface{}
	json.Unmarshal(result, &preview)

	if preview["requires_confirmation"] != true {
		t.Error("expected requires_confirmation=true in dry-run")
	}
	if inner.executed {
		t.Error("inner tool should NOT have been executed without confirm=true")
	}
}

func TestConfirmGate_Confirmed(t *testing.T) {
	inner := &mockTool{name: "test_restart", mutating: true}
	gated := WithConfirm(inner)

	input := json.RawMessage(`{"mac": "aa:bb:cc:dd:ee:ff", "confirm": true}`)
	_, err := gated.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !inner.executed {
		t.Error("inner tool should have been executed with confirm=true")
	}
}

func TestConfirmGate_NonMutating(t *testing.T) {
	inner := &mockTool{name: "test_list", mutating: false}
	gated := WithConfirm(inner)

	// Should be the same object (no wrapping for non-mutating)
	if gated != inner {
		t.Error("WithConfirm should return the same tool for non-mutating tools")
	}
}
