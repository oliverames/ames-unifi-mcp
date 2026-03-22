package version

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Doer is the minimal interface we need from the client.
type Doer interface {
	Do(ctx context.Context, method, path string, payload interface{}) (json.RawMessage, error)
	SitePath() string
}

// Info holds parsed controller version information.
type Info struct {
	Raw   string // e.g., "9.1.120"
	Major int
	Minor int
	Patch int
}

func (v Info) String() string {
	return v.Raw
}

// AtLeast returns true if the controller version is >= the given version.
func (v Info) AtLeast(major, minor, patch int) bool {
	if v.Major != major {
		return v.Major > major
	}
	if v.Minor != minor {
		return v.Minor > minor
	}
	return v.Patch >= patch
}

// HasZBF returns true if the controller supports Zone-Based Firewall (9.0+).
func (v Info) HasZBF() bool {
	return v.AtLeast(9, 0, 0)
}

// HasIntegrationAPI returns true if the local Integration API is available (9.0+).
func (v Info) HasIntegrationAPI() bool {
	return v.AtLeast(9, 0, 0)
}

// HasAPIKeyAuth returns true if API key auth is available locally (9.1.105+).
func (v Info) HasAPIKeyAuth() bool {
	return v.AtLeast(9, 1, 105)
}

// HasDNSPolicies returns true if DNS policies are available (10.0+).
func (v Info) HasDNSPolicies() bool {
	return v.AtLeast(10, 0, 0)
}

// HasACLRules returns true if ACL rules are available (10.0+).
func (v Info) HasACLRules() bool {
	return v.AtLeast(10, 0, 0)
}

// Detect queries the controller for its version.
func Detect(ctx context.Context, client Doer) (Info, error) {
	data, err := client.Do(ctx, "GET", client.SitePath()+"/stat/sysinfo", nil)
	if err != nil {
		return Info{}, fmt.Errorf("querying sysinfo: %w", err)
	}

	var sysinfo []struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &sysinfo); err != nil {
		return Info{}, fmt.Errorf("parsing sysinfo: %w", err)
	}
	if len(sysinfo) == 0 {
		return Info{}, fmt.Errorf("no sysinfo returned")
	}

	return Parse(sysinfo[0].Version)
}

// Parse parses a version string like "9.1.120" into an Info struct.
func Parse(raw string) (Info, error) {
	info := Info{Raw: raw}
	parts := strings.SplitN(raw, ".", 3)
	if len(parts) < 3 {
		return info, fmt.Errorf("unexpected version format: %s", raw)
	}

	var err error
	info.Major, err = strconv.Atoi(parts[0])
	if err != nil {
		return info, fmt.Errorf("parsing major version: %w", err)
	}
	info.Minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return info, fmt.Errorf("parsing minor version: %w", err)
	}
	info.Patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return info, fmt.Errorf("parsing patch version: %w", err)
	}

	return info, nil
}
