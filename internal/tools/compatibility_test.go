package tools_test

import (
	"context"
	"fmt"
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

func TestRemovedEventEndpointVersionBoundary(t *testing.T) {
	for _, major := range []int{0, 8, 9, 10} {
		t.Run(fmt.Sprint(major), func(t *testing.T) {
			var calls atomic.Int32
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if r.URL.Path != "/proxy/network/api/s/default/stat/event" {
					t.Errorf("wrong endpoint %s", r.URL.Path)
				}
				fmt.Fprint(w, `{"meta":{"rc":"ok"},"data":[]}`)
			}))
			defer s.Close()
			c, _ := client.New(&config.Config{Host: s.URL, APIKey: "test", Site: "default"})
			reg := tools.NewRegistry(permissions.NewChecker(config.PermAdmin), version.Info{Major: major})
			for _, tool := range core.BuildEventTools(c) {
				if err := reg.Register(tool); err != nil {
					t.Fatal(err)
				}
			}
			// Registry/lazy, eager Get, and batch all reach the compatibility gate.
			_, err := reg.Execute(context.Background(), "event_list", []byte(`{}`))
			if major < 9 {
				if err != nil || calls.Load() != 1 {
					t.Fatalf("legacy event failed: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), "not available") {
				t.Fatalf("expected compatibility error: %v", err)
			}
			tool, ok := reg.Get("event_list")
			if !ok {
				t.Fatal("tool unavailable to eager dispatch")
			}
			if _, err := tool.Execute(context.Background(), []byte(`{}`)); err == nil {
				t.Fatal("eager bypassed gate")
			}
			batch := reg.Batch(context.Background(), []tools.BatchCall{{Name: "event_list", Input: []byte(`{}`)}})
			if batch[0].Error == "" {
				t.Fatal("batch bypassed gate")
			}
			if calls.Load() != 0 {
				t.Fatal("removed endpoint contacted")
			}
			for _, meta := range reg.Index("") {
				if meta.Name == "event_list" && meta.RemovedIn != "9.0.0" {
					t.Fatal("removal metadata missing")
				}
			}
		})
	}
}

func TestMiscSelfUsesNetworkEndpoint(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/proxy/network/api/self" {
			t.Errorf("wrong self endpoint: %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"meta":{"rc":"ok"},"data":[]}`)
	}))
	defer s.Close()
	c, _ := client.New(&config.Config{Host: s.URL, APIKey: "test", Site: "default"})
	for _, tool := range extended.BuildMiscTools(c) {
		if tool.Name() == "misc_self" {
			if _, err := tool.Execute(context.Background(), []byte(`{}`)); err != nil {
				t.Fatal(err)
			}
			return
		}
	}
	t.Fatal("misc_self missing")
}

func TestIntegrationHandlersResolveSiteAndPreservePreview(t *testing.T) {
	const id = "88f7af54-98f8-306a-a1c7-c9349722b1f6"
	var discovery, devices, writes atomic.Int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/proxy/network/integration/v1/sites":
			discovery.Add(1)
			fmt.Fprintf(w, `{"totalCount":1,"data":[{"id":%q,"internalReference":"default"}]}`, id)
		case "/proxy/network/integration/v1/sites/" + id + "/devices":
			devices.Add(1)
			fmt.Fprint(w, `{"data":[]}`)
		case "/proxy/network/integration/v1/sites/" + id + "/hotspot/vouchers":
			writes.Add(1)
			if r.Method != "POST" {
				t.Error("wrong method")
			}
			fmt.Fprint(w, `{"data":[]}`)
		default:
			t.Errorf("wrong Integration path %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer s.Close()
	c, _ := client.New(&config.Config{Host: s.URL, APIKey: "test", Site: "default"})
	reg := tools.NewRegistry(permissions.NewChecker(config.PermAdmin), version.Info{Major: 10})
	for _, tool := range append(core.BuildDeviceTools(c), extended.BuildHotspotTools(c)...) {
		if err := reg.Register(tool); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := reg.Execute(context.Background(), "hotspot_voucher_create_v2", []byte(`{"config":{}}`)); err != nil {
		t.Fatal(err)
	}
	if discovery.Load() != 0 || writes.Load() != 0 {
		t.Fatal("preview made requests")
	}
	if _, err := reg.Execute(context.Background(), "device_list_v2", []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := reg.Execute(context.Background(), "hotspot_voucher_create_v2", []byte(`{"config":{},"confirm":true}`)); err != nil {
		t.Fatal(err)
	}
	if discovery.Load() != 1 || devices.Load() != 1 || writes.Load() != 1 {
		t.Fatal("wrong discovery/read/write counts")
	}
}
