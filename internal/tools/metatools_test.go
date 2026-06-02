package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/oliveames/ames-unifi-mcp/internal/config"
	"github.com/oliveames/ames-unifi-mcp/internal/permissions"
	"github.com/oliveames/ames-unifi-mcp/internal/version"
)

func TestMetaToolBatchRejectsOversizedBatch(t *testing.T) {
	registry := NewRegistry(permissions.NewChecker(config.PermAdmin), version.Info{Raw: "10.3.58", Major: 10, Minor: 3, Patch: 58})
	batch := NewMetaToolBatch(registry)

	calls := make([]BatchCall, maxBatchCalls+1)
	for i := range calls {
		calls[i] = BatchCall{Name: "missing_tool", Input: json.RawMessage(`{}`)}
	}
	input, err := json.Marshal(map[string]interface{}{"calls": calls})
	if err != nil {
		t.Fatalf("marshal input: %v", err)
	}

	_, err = batch.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected oversized batch to be rejected")
	}
	if !strings.Contains(err.Error(), "too many batch calls") {
		t.Fatalf("expected batch-size error, got %v", err)
	}
}
