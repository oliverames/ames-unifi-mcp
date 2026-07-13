package config

import (
	"os"
	"path/filepath"
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

func TestLoadResolvesOnlyCallerSupplied1PasswordReferences(t *testing.T) {
	binDir := t.TempDir()
	opPath := filepath.Join(binDir, "op")
	script := `#!/bin/sh
if [ "$1" != "read" ]; then
  exit 2
fi
case "$2" in
  op://Test/UniFi/host) printf '%s\n' 'https://192.0.2.10' ;;
  op://Test/UniFi/api_key) printf '%s\n' 'test-api-key' ;;
  *) exit 3 ;;
esac
`
	if err := os.WriteFile(opPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake op: %v", err)
	}
	t.Setenv("PATH", binDir)
	t.Setenv("UNIFI_HOST", "")
	t.Setenv("UNIFI_API_KEY", "")
	t.Setenv("UNIFI_USERNAME", "")
	t.Setenv("UNIFI_PASSWORD", "")
	t.Setenv("UNIFI_HOST_OP_REF", "op://Test/UniFi/host")
	t.Setenv("UNIFI_API_KEY_OP_REF", "op://Test/UniFi/api_key")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Host != "https://192.0.2.10" || cfg.APIKey != "test-api-key" {
		t.Fatalf("unexpected resolved credentials: host=%q api key set=%t", cfg.Host, cfg.APIKey != "")
	}
	if cfg.NeedsAuth {
		t.Fatal("caller-supplied references should provide usable credentials")
	}
}

func TestLoadDoesNotProbe1PasswordWithoutReferences(t *testing.T) {
	binDir := t.TempDir()
	calledPath := filepath.Join(t.TempDir(), "called")
	opPath := filepath.Join(binDir, "op")
	script := "#!/bin/sh\n/usr/bin/touch \"$OP_CALLED_PATH\"\nexit 0\n"
	if err := os.WriteFile(opPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake op: %v", err)
	}
	t.Setenv("PATH", binDir)
	t.Setenv("OP_CALLED_PATH", calledPath)
	t.Setenv("UNIFI_HOST", "")
	t.Setenv("UNIFI_API_KEY", "")
	t.Setenv("UNIFI_USERNAME", "")
	t.Setenv("UNIFI_PASSWORD", "")
	t.Setenv("UNIFI_HOST_OP_REF", "")
	t.Setenv("UNIFI_API_KEY_OP_REF", "")
	t.Setenv("UNIFI_USERNAME_OP_REF", "")
	t.Setenv("UNIFI_PASSWORD_OP_REF", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.NeedsAuth {
		t.Fatal("missing credentials should leave the server in needs-auth state")
	}
	if _, err := os.Stat(calledPath); !os.IsNotExist(err) {
		t.Fatal("Load() invoked op without a caller-supplied reference")
	}
}
