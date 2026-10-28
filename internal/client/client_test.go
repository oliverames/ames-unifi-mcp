package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/oliverames/ames-unifi-mcp/internal/config"
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

func TestDotSegmentPathsAreRefused(t *testing.T) {
	requests := 0
	c := &Client{cfg: &config.Config{Host: "https://controller.example", APIKey: "test-key", Site: "default"}}
	c.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		requests++
		return response(200, `{"meta":{"rc":"ok"},"data":[]}`), nil
	})}

	for _, path := range []string{
		"api/s/default/set/setting/../../cmd/devmgr",
		"api/s/default/set/setting/%2e%2e/%2e%2e/%2e%2e/%2e%2e/api/system/reboot",
		"api/s/default/get/setting/./mgmt",
	} {
		_, err := c.Do(context.Background(), "POST", path, map[string]string{"cmd": "restart"})
		if err == nil || !strings.Contains(err.Error(), "dot segment") {
			t.Fatalf("%s: expected dot-segment refusal, got %v", path, err)
		}
	}
	if _, err := c.DoRaw(context.Background(), "POST", c.cfg.BaseURL()+"/v2/api/site/default/../../../../api/system/reboot", nil); err == nil || !strings.Contains(err.Error(), "dot segment") {
		t.Fatalf("expected raw dot-segment refusal, got %v", err)
	}
	if requests != 0 {
		t.Fatalf("dispatched %d dot-segment requests", requests)
	}

	// Dots inside a segment, such as report interval and type, stay valid.
	if _, err := c.Do(context.Background(), "POST", "api/s/default/stat/report/5minutes.site", nil); err != nil {
		t.Fatalf("ordinary dotted segment refused: %v", err)
	}
}

func TestRedirectsAreHeldToTheControllerHost(t *testing.T) {
	var foreignHits atomic.Int32
	foreign := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		foreignHits.Add(1)
		fmt.Fprint(w, `{}`)
	}))
	defer foreign.Close()

	controller := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/proxy/network/v2/api/away":
			http.Redirect(w, r, foreign.URL+"/collect", http.StatusFound)
		case "/proxy/network/v2/api/moved":
			http.Redirect(w, r, "/proxy/network/v2/api/here", http.StatusFound)
		default:
			fmt.Fprint(w, `{"ok":true}`)
		}
	}))
	defer controller.Close()

	c, err := New(&config.Config{Host: controller.URL, APIKey: "test-key", Site: "default", VerifySSL: true})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = c.DoRaw(context.Background(), "GET", c.cfg.BaseURL()+"/v2/api/away", nil)
	if err == nil || !strings.Contains(err.Error(), "unexpected host") {
		t.Fatalf("expected cross-host redirect refusal, got %v", err)
	}
	if got := foreignHits.Load(); got != 0 {
		t.Fatalf("API key followed a redirect to another host %d times", got)
	}

	body, err := c.DoRaw(context.Background(), "GET", c.cfg.BaseURL()+"/v2/api/moved", nil)
	if err != nil || string(body) != `{"ok":true}` {
		t.Fatalf("same-host redirect failed: %s, %v", body, err)
	}
}
