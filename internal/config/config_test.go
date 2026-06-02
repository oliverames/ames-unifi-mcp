package config

import (
	"strings"
	"testing"
)

func TestLoadDefaultsInvalidVerifySSLToTrue(t *testing.T) {
	t.Setenv("UNIFI_HOST", "https://192.168.1.1")
	t.Setenv("UNIFI_API_KEY", "test-key")
	t.Setenv("UNIFI_VERIFY_SSL", "definitely")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.VerifySSL {
		t.Fatal("invalid UNIFI_VERIFY_SSL should default to true")
	}
}

func TestLoadRejectsUnsafeHost(t *testing.T) {
	t.Setenv("UNIFI_HOST", "https://user:pass@example.com/proxy/network")
	t.Setenv("UNIFI_API_KEY", "test-key")

	_, err := Load()
	if err == nil {
		t.Fatal("expected Load() to reject host with embedded credentials and path")
	}
	if !strings.Contains(err.Error(), "credentials") {
		t.Fatalf("expected credentials error, got %v", err)
	}
}

func TestLoadRejectsInvalidSiteID(t *testing.T) {
	t.Setenv("UNIFI_HOST", "https://192.168.1.1")
	t.Setenv("UNIFI_API_KEY", "test-key")
	t.Setenv("UNIFI_SITE", "../default")

	_, err := Load()
	if err == nil {
		t.Fatal("expected Load() to reject invalid site identifier")
	}
	if !strings.Contains(err.Error(), "UNIFI_SITE") {
		t.Fatalf("expected site error, got %v", err)
	}
}
