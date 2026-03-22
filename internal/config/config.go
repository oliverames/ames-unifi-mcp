package config

import (
	"fmt"
	"os"
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

	if cfg.Host == "" {
		return nil, fmt.Errorf("UNIFI_HOST is required")
	}
	if cfg.APIKey == "" && (cfg.Username == "" || cfg.Password == "") {
		return nil, fmt.Errorf("either UNIFI_API_KEY or UNIFI_USERNAME+UNIFI_PASSWORD is required")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
