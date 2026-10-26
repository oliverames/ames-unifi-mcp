package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/oliverames/ames-unifi-mcp/internal/config"
	"github.com/oliverames/ames-unifi-mcp/internal/permissions"
	"github.com/oliverames/ames-unifi-mcp/internal/version"
)

type panicTool struct{ mockTool }

func (p *panicTool) Execute(context.Context, json.RawMessage) (json.RawMessage, error) {
	panic("boom")
}

// A panicking tool inside tool_batch runs in its own goroutine, where an
// unrecovered panic would kill the whole server.
func TestBatchRecoversToolPanic(t *testing.T) {
	r := NewRegistry(permissions.NewChecker(config.PermAdmin), version.Info{Raw: "10.0.0", Major: 10})
	if err := r.Register(&panicTool{mockTool{name: "panics"}}); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(&mockTool{name: "works"}); err != nil {
		t.Fatal(err)
	}
	results := r.Batch(context.Background(), []BatchCall{
		{Name: "panics", Input: json.RawMessage(`{"mac":"aa:bb:cc:dd:ee:ff"}`)},
		{Name: "works", Input: json.RawMessage(`{"mac":"aa:bb:cc:dd:ee:ff"}`)},
	})
	if results[0].Error != "internal error in tool panics" {
		t.Errorf("panicking tool error = %q", results[0].Error)
	}
	if results[1].Error != "" {
		t.Errorf("healthy tool error = %q", results[1].Error)
	}
}
