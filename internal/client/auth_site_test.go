package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/oliverames/ames-unifi-mcp/internal/config"
)

func TestPermanent401DoesNotStormLogin(t *testing.T) {
	for _, raw := range []bool{false, true} {
		for _, loginOK := range []bool{false, true} {
			t.Run(fmt.Sprintf("raw=%v/loginOK=%v", raw, loginOK), func(t *testing.T) {
				var logins atomic.Int32
				c := &Client{cfg: &config.Config{Host: "https://controller.example", Username: "test", Password: "test"}}
				c.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
					if r.URL.Path == "/api/auth/login" {
						logins.Add(1)
						if loginOK {
							return response(200, `{}`), nil
						}
						return response(403, `denied`), nil
					}
					return response(401, `denied`), nil
				})}
				var wg sync.WaitGroup
				for range 20 {
					wg.Go(func() {
						var err error
						if raw {
							_, err = c.DoRaw(context.Background(), "GET", c.cfg.BaseURL()+"/v2/api/data", nil)
						} else {
							_, err = c.Do(context.Background(), "GET", "api/self", nil)
						}
						if err == nil {
							t.Error("expected authentication error")
						}
					})
				}
				wg.Wait()
				if got := logins.Load(); got != 1 {
					t.Fatalf("login attempts=%d, want 1", got)
				}
			})
		}
	}
}

func TestFreshSessionAndIntegrationAuth(t *testing.T) {
	var logins, requests atomic.Int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/login" {
			logins.Add(1)
			fmt.Fprint(w, `{}`)
			return
		}
		requests.Add(1)
		w.WriteHeader(401)
	}))
	defer s.Close()
	c, err := New(&config.Config{Host: s.URL, Username: "test", Password: "test", Site: "default"})
	if err != nil {
		t.Fatal(err)
	}
	for range 5 {
		if _, err := c.Do(context.Background(), "GET", "api/self", nil); err == nil {
			t.Fatal("expected 401")
		}
	}
	if logins.Load() != 1 {
		t.Fatal("fresh session caused another login")
	}
	before := requests.Load()
	for _, path := range []string{"/v1/sites", "/v1/info", "/v1/pending-devices"} {
		_, err := c.DoRaw(context.Background(), "GET", c.cfg.BaseURL()+"/integration"+path, nil)
		if err == nil || !strings.Contains(err.Error(), "UNIFI_API_KEY") {
			t.Fatalf("wrong auth error: %v", err)
		}
	}
	if requests.Load() != before {
		t.Fatal("unsupported auth reached controller")
	}
	// An actually aged session still gets one refresh attempt.
	c.mu.Lock()
	c.lastAuthSuccess = time.Now().Add(-time.Hour)
	c.mu.Unlock()
	c.Do(context.Background(), "GET", "api/self", nil)
	if logins.Load() != 2 {
		t.Fatal("expired session did not refresh")
	}
}

func TestIntegrationSiteResolutionPaginationAndCache(t *testing.T) {
	const id = "88f7af54-98f8-306a-a1c7-c9349722b1f6"
	var lookups atomic.Int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "test" {
			t.Error("API key missing")
		}
		switch r.URL.Path {
		case "/proxy/network/integration/v1/sites":
			lookups.Add(1)
			if r.URL.Query().Get("offset") == "0" {
				fmt.Fprint(w, `{"totalCount":2,"data":[{"id":"00000000-0000-0000-0000-000000000000","internalReference":"other","name":"default"}]}`)
			} else {
				fmt.Fprintf(w, `{"totalCount":2,"data":[{"id":%q,"internalReference":"default","name":"Office"}]}`, id)
			}
		case "/proxy/network/integration/v1/sites/" + id + "/devices":
			fmt.Fprint(w, `{"data":[]}`)
		case "/proxy/network/api/s/default/stat/device":
			fmt.Fprint(w, `{"meta":{"rc":"ok"},"data":[]}`)
		default:
			t.Errorf("wrong site path: %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer s.Close()
	c, err := New(&config.Config{Host: s.URL, APIKey: "test", Site: "default"})
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			if _, err := c.DoIntegrationSite(context.Background(), "GET", "/devices", nil); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if lookups.Load() != 2 {
		t.Fatalf("lookup requests=%d", lookups.Load())
	}
	if c.Site() != "default" {
		t.Fatal("legacy site changed")
	}
	if _, err := c.Do(context.Background(), "GET", c.SitePath()+"/stat/device", nil); err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationSiteFailuresDoNotDispatchOrPoisonCache(t *testing.T) {
	for _, body := range []string{`{`, `{"data":[]}`, `{"data":[{"id":"../../other","internalReference":"default"}]}`} {
		t.Run(body, func(t *testing.T) {
			var calls atomic.Int32
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); fmt.Fprint(w, body) }))
			defer s.Close()
			c, _ := New(&config.Config{Host: s.URL, APIKey: "test", Site: "default"})
			for range 2 {
				if _, err := c.DoIntegrationSite(context.Background(), "DELETE", "/devices/device", nil); err == nil {
					t.Fatal("expected lookup error")
				}
			}
			if calls.Load() != 2 || c.integrationSiteID != "" {
				t.Fatal("failed lookup cached or mutation dispatched")
			}
		})
	}
}
