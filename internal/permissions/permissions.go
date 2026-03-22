package permissions

import "github.com/oliveames/ames-unifi-mcp/internal/config"

// Action represents what a tool does.
type Action int

const (
	ActionRead Action = iota
	ActionCreate
	ActionUpdate
	ActionDelete
	ActionExecute // for cmd/ operations
)

// Category represents a resource domain.
type Category string

const (
	CatDevices  Category = "devices"
	CatClients  Category = "clients"
	CatNetworks Category = "networks"
	CatWLAN     Category = "wlan"
	CatFirewall Category = "firewall"
	CatVPN      Category = "vpn"
	CatRouting  Category = "routing"
	CatQoS      Category = "qos"
	CatStats    Category = "stats"
	CatEvents   Category = "events"
	CatSystem   Category = "system"
	CatHotspot  Category = "hotspot"
	CatPoE      Category = "poe"
	CatDPI      Category = "dpi"
	CatBackup   Category = "backup"
	CatSettings Category = "settings"
)

// Checker evaluates whether an action is permitted.
type Checker struct {
	profile config.PermissionProfile
}

func NewChecker(profile config.PermissionProfile) *Checker {
	return &Checker{profile: profile}
}

// Allowed returns true if the given action on the given category is permitted.
func (c *Checker) Allowed(cat Category, action Action) bool {
	// Read is always allowed
	if action == ActionRead {
		return true
	}

	switch c.profile {
	case config.PermReadOnly:
		return false
	case config.PermStandard:
		return c.standardAllowed(cat, action)
	case config.PermAdmin:
		return true
	}
	return false
}

func (c *Checker) standardAllowed(cat Category, action Action) bool {
	// Standard profile allows most mutations except destructive system operations
	switch cat {
	case CatBackup:
		return action == ActionRead // backup trigger requires admin
	case CatSystem:
		return false // reboot, poweroff require admin
	case CatFirewall:
		return action != ActionDelete // can create/update but not delete firewall rules
	case CatPoE:
		return false // PoE power cycling requires admin
	default:
		return true
	}
}

// ProfileName returns the human-readable profile name.
func (c *Checker) ProfileName() string {
	switch c.profile {
	case config.PermReadOnly:
		return "read-only"
	case config.PermStandard:
		return "standard"
	case config.PermAdmin:
		return "admin"
	}
	return "unknown"
}
