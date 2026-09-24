package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/oliverames/ames-unifi-mcp/internal/config"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func response(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}

func TestRetryPolicy(t *testing.T) {
	for _, method := range []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE"} {
		for _, failure := range []string{"network", "429", "503"} {
			t.Run(method+"/"+failure, func(t *testing.T) {
				calls := 0
				c := &Client{cfg: &config.Config{Host: "https://controller.example"}}
				c.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
					calls++
					if calls > 1 {
						return response(200, `{}`), nil
					}
					switch failure {
					case "network":
						return nil, errors.New("response lost after server accepted request")
					case "429":
						return response(429, `rate limited`), nil
					default:
						return response(503, `response unavailable`), nil
					}
				})}
				resp, err := c.doWithRetry(func() (*http.Request, error) {
					return http.NewRequestWithContext(context.Background(), method, "https://controller.example/api", strings.NewReader(`{"create":true}`))
				}, 1)
				if resp != nil {
					defer resp.Body.Close()
				}
				if method == "GET" || method == "HEAD" {
					if calls != 2 || err != nil || resp.StatusCode != 200 {
						t.Fatalf("read: calls=%d, err=%v, response=%v", calls, err, resp)
					}
				} else {
					if calls != 1 {
						t.Fatalf("mutation replayed: %d requests", calls)
					}
					if failure == "network" && err == nil {
						t.Fatal("lost response must be reported")
					}
					if failure != "network" && (err != nil || resp.StatusCode < 400) {
						t.Fatalf("server failure lost: %v, %v", resp, err)
					}
				}
			})
		}
	}
}

func TestRetryBackoffHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	c := &Client{cfg: &config.Config{Host: "https://controller.example"}}
	c.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		time.AfterFunc(25*time.Millisecond, cancel)
		return response(503, "unavailable"), nil
	})}
	start := time.Now()
	_, err := c.doWithRetry(func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, "GET", "https://controller.example/api", nil)
	}, 1)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("sent %d requests after cancellation", calls)
	}
	if time.Since(start) > 250*time.Millisecond {
		t.Fatal("cancellation waited for retry backoff")
	}
}

type brokenBody struct{}

func (brokenBody) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
func (brokenBody) Close() error             { return nil }

func TestReloginPreservesBodyReadErrors(t *testing.T) {
	for _, raw := range []bool{false, true} {
		t.Run(map[bool]string{false: "legacy", true: "raw"}[raw], func(t *testing.T) {
			requests := 0
			c := &Client{cfg: &config.Config{Host: "https://controller.example", Username: "test", Password: "test"}}
			c.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path == "/api/auth/login" {
					return response(200, `{}`), nil
				}
				requests++
				if requests == 1 {
					return response(401, `expired`), nil
				}
				r2 := response(200, "")
				r2.Body = brokenBody{}
				return r2, nil
			})}
			var err error
			if raw {
				_, err = c.DoRaw(context.Background(), "GET", "https://controller.example/v1/data", nil)
			} else {
				_, err = c.Do(context.Background(), "GET", "api/s/default/data", nil)
			}
			if !errors.Is(err, io.ErrUnexpectedEOF) {
				t.Fatalf("body read error discarded: %v", err)
			}
		})
	}
}

func TestReloginRevalidatesRequestHost(t *testing.T) {
	for _, raw := range []bool{false, true} {
		t.Run(map[bool]string{false: "legacy", true: "raw"}[raw], func(t *testing.T) {
			requests := 0
			c := &Client{cfg: &config.Config{Host: "https://controller.example", Username: "test", Password: "test"}}
			c.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path == "/api/auth/login" {
					// Simulate an allowed-origin change while the session is refreshed.
					c.cfg.Host = "https://replacement.example"
					return response(200, `{}`), nil
				}
				requests++
				return response(401, `expired`), nil
			})}
			var err error
			if raw {
				_, err = c.DoRaw(context.Background(), "GET", "https://controller.example/v1/data", nil)
			} else {
				_, err = c.Do(context.Background(), "GET", "api/s/default/data", nil)
			}
			if err == nil || !strings.Contains(err.Error(), "unexpected host") {
				t.Fatalf("retry host was not validated: %v", err)
			}
			if requests != 1 {
				t.Fatalf("dispatched to stale host: %d requests", requests)
			}
		})
	}
}
