package client

import (
	"context"
	"strings"
	"testing"

	"github.com/oliveames/ames-unifi-mcp/internal/config"
)

func TestDoRawRejectsUnexpectedHost(t *testing.T) {
	cfg := &config.Config{
		Host:      "https://controller.example",
		APIKey:    "test-key",
		Site:      "default",
		VerifySSL: true,
	}
	c, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = c.DoRaw(context.Background(), "GET", "https://evil.example/v1/info", nil)
	if err == nil {
		t.Fatal("expected unexpected host rejection")
	}
	if !strings.Contains(err.Error(), "unexpected host") {
		t.Fatalf("expected unexpected host error, got %v", err)
	}
}
