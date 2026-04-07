package config

import (
	"os"
	"os/exec"
	"strings"
)

type AuthMethod int

const (
	AuthAPIKey AuthMethod = iota
	AuthUserPass
)

type ToolMode int

const (
	ToolModeLazy ToolMode = iota
	ToolModeEager
)

type PermissionProfile int

const (
	PermReadOnly PermissionProfile = iota
	PermStandard
	PermAdmin
)

type Config struct {
	Host              string
	APIKey            string
	Username          string
	Password          string
	Site              string
	VerifySSL         bool
	ToolMode          ToolMode
	PermissionProfile PermissionProfile
	LogLevel          string
	// NeedsAuth is set when Load() finds no usable credentials. The server
	// still starts and registers tools, but each tool short-circuits with
	// an "authentication required" error. This lets the plugin appear
	// "installed but inactive" instead of erroring on startup.
	NeedsAuth bool
}

// AuthHint returns a user-facing message explaining how to configure credentials.
// Returned by tool handlers when NeedsAuth is true.
func (c *Config) AuthHint() string {
	return "UniFi Controller credentials not configured. Either set " +
		"UNIFI_HOST and UNIFI_API_KEY (or UNIFI_USERNAME+UNIFI_PASSWORD) in " +
		"the environment, or create a 'UniFi Controller' item in the " +
		"Development 1Password vault with fields: host, api_key (or " +
		"username + password)."
}

func (c *Config) AuthMethod() AuthMethod {
	if c.APIKey != "" {
		return AuthAPIKey
	}
	return AuthUserPass
}

// BaseURL returns the host with /proxy/network prefix for UniFi OS.
func (c *Config) BaseURL() string {
	host := strings.TrimRight(c.Host, "/")
	return host + "/proxy/network"
}

// CloudBaseURL returns the cloud API base URL.
func (c *Config) CloudBaseURL() string {
	return "https://api.ui.com"
}

func Load() (*Config, error) {
	cfg := &Config{
		Host:      os.Getenv("UNIFI_HOST"),
		APIKey:    os.Getenv("UNIFI_API_KEY"),
		Username:  os.Getenv("UNIFI_USERNAME"),
		Password:  os.Getenv("UNIFI_PASSWORD"),
		Site:      envOrDefault("UNIFI_SITE", "default"),
		VerifySSL: envOrDefault("UNIFI_VERIFY_SSL", "true") == "true",
		LogLevel:  envOrDefault("UNIFI_LOG_LEVEL", "error"),
	}

	switch strings.ToLower(envOrDefault("UNIFI_TOOL_MODE", "lazy")) {
	case "eager":
		cfg.ToolMode = ToolModeEager
	default:
		cfg.ToolMode = ToolModeLazy
	}

	switch strings.ToLower(envOrDefault("UNIFI_PERMISSION_PROFILE", "standard")) {
	case "read-only", "readonly":
		cfg.PermissionProfile = PermReadOnly
	case "admin":
		cfg.PermissionProfile = PermAdmin
	default:
		cfg.PermissionProfile = PermStandard
	}

	// 1Password fallback for missing credentials
	if cfg.Host == "" {
		cfg.Host = opRead("op://Development/UniFi Controller/host")
	}
	if cfg.APIKey == "" {
		cfg.APIKey = opRead("op://Development/UniFi Controller/api_key")
	}
	if cfg.Username == "" {
		cfg.Username = opRead("op://Development/UniFi Controller/username")
	}
	if cfg.Password == "" {
		cfg.Password = opRead("op://Development/UniFi Controller/password")
	}

	// Soft-fail: if credentials are missing, mark as NeedsAuth and let the
	// server start anyway. Tool handlers check this flag and return a
	// structured "configure me" error instead of running.
	if cfg.Host == "" || (cfg.APIKey == "" && (cfg.Username == "" || cfg.Password == "")) {
		cfg.NeedsAuth = true
	}

	return cfg, nil
}

// opRead attempts to read a secret from 1Password CLI.
// Returns empty string if op is unavailable or the item doesn't exist.
func opRead(ref string) string {
	cmd := exec.Command("op", "read", ref)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
