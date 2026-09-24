package tools_test

import (
	"context"
	"encoding/json"
	"github.com/oliverames/ames-unifi-mcp/internal/inputvalidation"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/oliverames/ames-unifi-mcp/internal/client"
	"github.com/oliverames/ames-unifi-mcp/internal/config"
	"github.com/oliverames/ames-unifi-mcp/internal/permissions"
	"github.com/oliverames/ames-unifi-mcp/internal/tools"
	"github.com/oliverames/ames-unifi-mcp/internal/tools/core"
	"github.com/oliverames/ames-unifi-mcp/internal/tools/extended"
	"github.com/oliverames/ames-unifi-mcp/internal/version"
)

func TestInvalidInputsNeverReachController(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"meta":{"rc":"ok"},"data":[]}`))
	}))
	defer server.Close()
	c, err := client.New(&config.Config{Host: server.URL, APIKey: "test", Site: "default"})
	if err != nil {
		t.Fatal(err)
	}
	registry := tools.NewRegistry(permissions.NewChecker(config.PermAdmin), version.Info{Major: 10})
	all := append(core.BuildDeviceTools(c), extended.BuildAdminTools(c)...)
	var get, restart tools.Tool
	for _, tool := range all {
		if err := registry.Register(tool); err != nil {
			t.Fatal(err)
		}
		if tool.Name() == "device_get" {
			get = tool
		}
		if tool.Name() == "device_restart" {
			restart = tool
		}
	}
	ctx := context.Background()
	dispatch := map[string]func(json.RawMessage) error{
		"direct":   func(input json.RawMessage) error { _, err := get.Execute(ctx, input); return err },
		"registry": func(input json.RawMessage) error { _, err := registry.Execute(ctx, "device_get", input); return err },
		"lazy": func(input json.RawMessage) error {
			outer, _ := json.Marshal(map[string]any{"tool_name": "device_get", "input": input})
			if !json.Valid(input) {
				outer = []byte(`{"tool_name":"device_get","input":` + string(input) + `}`)
			}
			_, err := tools.NewMetaToolExecute(registry).Execute(ctx, outer)
			return err
		},
		"batch": func(input json.RawMessage) error {
			result := registry.Batch(ctx, []tools.BatchCall{{Name: "device_get", Input: input}})
			if result[0].Error == "" {
				return nil
			}
			return &inputError{}
		},
	}
	for name, call := range dispatch {
		t.Run(name, func(t *testing.T) {
			for _, input := range []string{`{`, `null`, `[]`, `{}`, `{"mac":null}`, `{"mac":4}`, `{"mac":""}`, `{"mac":"../bad"}`, `{"mac":"00:11:22:33:44:55:66:77"}`} {
				before := requests.Load()
				if call([]byte(input)) == nil {
					t.Errorf("accepted %s", input)
				}
				if requests.Load() != before {
					t.Fatalf("controller contacted for %s", input)
				}
			}
			before := requests.Load()
			if err := call([]byte(`{"mac":"00:11:22:33:44:55"}`)); err != nil {
				t.Fatal(err)
			}
			if requests.Load() != before+1 {
				t.Fatal("valid call not dispatched")
			}
		})
	}
	for _, input := range []string{`{"mac":"00:11:22:33:44:55","reboot_type":"invalid","confirm":true}`, `{"mac":null,"confirm":true}`, `{"mac":"00:11:22:33:44:55","confirm":"true"}`} {
		before := requests.Load()
		if _, err := registry.Execute(ctx, "device_restart", []byte(input)); err == nil {
			t.Errorf("accepted %s", input)
		}
		if requests.Load() != before {
			t.Fatal("invalid mutation contacted controller")
		}
	}
	for _, tc := range []struct{ name, input string }{
		{"device_adv_adopt", `{"mac":"00:11:22:33:44:55","ip":"192.0.2.1","username":"test","password":"test","url":"http://example.test/inform","port":999999999999999999999999999999,"confirm":true}`},
		{"admin_revoke_super", `{"admin_id":"../admin","confirm":true}`},
		{"admin_revoke_super", `{"admin_id":null,"confirm":true}`},
		{"admin_revoke_super", `{"admin_id":"","confirm":true}`},
	} {
		if _, ok := registry.Get(tc.name); !ok {
			t.Fatalf("missing tool %s", tc.name)
		}
		before := requests.Load()
		if _, err := registry.Execute(ctx, tc.name, []byte(tc.input)); err == nil {
			t.Errorf("accepted %s", tc.input)
		}
		if requests.Load() != before {
			t.Fatal("invalid input contacted controller")
		}
	}
	before := requests.Load()
	if _, err := restart.Execute(ctx, []byte(`{"mac":"00:11:22:33:44:55","reboot_type":"soft"}`)); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != before+1 {
		t.Fatal("valid mutation not dispatched")
	}
}

type inputError struct{}

func (*inputError) Error() string { return "invalid input" }

func TestAllToolSchemasAreValid(t *testing.T) {
	builders := []func(*client.Client) []*core.BaseTool{
		core.BuildFirewallLegacyTools,
		core.BuildFirewallZBFTools,
		core.BuildDNSTools,
		core.BuildSystemTools,
		core.BuildWiFiTools,
		core.BuildWLANTools,
		core.BuildSwitchingTools,
		core.BuildEventTools,
		core.BuildClientTools,
		core.BuildStatsTools,
		core.BuildTrafficTools,
		core.BuildWANTools,
		core.BuildACLTools,
		core.BuildNetworkTools,
		core.BuildDeviceTools,
		extended.BuildCloudTools,
		extended.BuildMiscTools,
		extended.BuildPoETools,
		extended.BuildSyslogTools,
		extended.BuildAdminTools,
		extended.BuildAPGroupTools,
		extended.BuildHotspotTools,
	}
	c, err := client.New(&config.Config{Host: "http://127.0.0.1:1", APIKey: "test", Site: "default"})
	if err != nil {
		t.Fatal(err)
	}
	for _, build := range builders {
		for _, tool := range build(c) {
			if err := inputvalidation.Validate(tool.InputSchema(), []byte(`{}`)); err != nil && strings.Contains(err.Error(), "invalid tool schema") {
				t.Errorf("%s: %v", tool.Name(), err)
			}
		}
	}
}
