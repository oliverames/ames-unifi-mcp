package permissions

import (
	"testing"

	"github.com/oliveames/ames-unifi-mcp/internal/config"
)

func TestReadOnly(t *testing.T) {
	c := NewChecker(config.PermReadOnly)

	if !c.Allowed(CatDevices, ActionRead) {
		t.Error("read-only should allow reads")
	}
	if c.Allowed(CatDevices, ActionCreate) {
		t.Error("read-only should not allow creates")
	}
	if c.Allowed(CatDevices, ActionDelete) {
		t.Error("read-only should not allow deletes")
	}
}

func TestStandard(t *testing.T) {
	c := NewChecker(config.PermStandard)

	if !c.Allowed(CatDevices, ActionRead) {
		t.Error("standard should allow reads")
	}
	if !c.Allowed(CatClients, ActionExecute) {
		t.Error("standard should allow client mutations")
	}
	if c.Allowed(CatPoE, ActionExecute) {
		t.Error("standard should not allow PoE")
	}
	if c.Allowed(CatSystem, ActionExecute) {
		t.Error("standard should not allow system mutations")
	}
	if c.Allowed(CatFirewall, ActionDelete) {
		t.Error("standard should not allow firewall deletes")
	}
	if !c.Allowed(CatFirewall, ActionCreate) {
		t.Error("standard should allow firewall creates")
	}
}

func TestAdmin(t *testing.T) {
	c := NewChecker(config.PermAdmin)

	if !c.Allowed(CatPoE, ActionExecute) {
		t.Error("admin should allow PoE")
	}
	if !c.Allowed(CatSystem, ActionExecute) {
		t.Error("admin should allow system mutations")
	}
	if !c.Allowed(CatFirewall, ActionDelete) {
		t.Error("admin should allow firewall deletes")
	}
}
