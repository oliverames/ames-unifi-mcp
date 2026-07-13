package main

import (
	"testing"

	"github.com/oliverames/ames-unifi-mcp/internal/client"
	"github.com/oliverames/ames-unifi-mcp/internal/config"
)

func TestBuildAllToolsMatchesDocumentedCatalog(t *testing.T) {
	c, err := client.New(&config.Config{NeedsAuth: true, VerifySSL: true})
	if err != nil {
		t.Fatalf("client.New() error = %v", err)
	}

	allTools := buildAllTools(c)
	if got, want := len(allTools), 310; got != want {
		t.Fatalf("tool count = %d, want %d", got, want)
	}

	seen := make(map[string]struct{}, len(allTools))
	for _, tool := range allTools {
		if _, exists := seen[tool.Name()]; exists {
			t.Fatalf("duplicate tool name %q", tool.Name())
		}
		seen[tool.Name()] = struct{}{}
	}
}
