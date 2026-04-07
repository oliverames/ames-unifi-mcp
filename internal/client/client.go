package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"

	"github.com/oliveames/ames-unifi-mcp/internal/config"
)

// Client handles HTTP communication with the UniFi controller.
// Targets UniFi OS (Dream Machine) — always uses /proxy/network prefix.
type Client struct {
	cfg        *config.Config
	httpClient *http.Client
	mu         sync.RWMutex // protects csrfToken and loginGen
	csrfToken  string
	loginGen   uint64    // incremented on each successful login; used to coalesce concurrent re-logins
	loginMu    sync.Mutex // serializes re-login attempts
}

// LegacyResponse is the envelope for legacy API responses.
type LegacyResponse struct {
	Meta struct {
		RC  string `json:"rc"`
		Msg string `json:"msg,omitempty"`
	} `json:"meta"`
	Data json.RawMessage `json:"data"`
}

func New(cfg *config.Config) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("creating cookie jar: %w", err)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !cfg.VerifySSL,
		},
	}

	httpClient := &http.Client{
		Jar:       jar,
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	c := &Client{
		cfg:        cfg,
		httpClient: httpClient,
	}

	// If unauthenticated (no creds), skip login. The server starts in a
	// degraded "needs auth" state — tool handlers in main.go short-circuit
	// before any client method gets called.
	if cfg.NeedsAuth {
		return c, nil
	}

	// If using username/password, login immediately to establish session.
	if cfg.AuthMethod() == config.AuthUserPass {
		if err := c.login(context.Background()); err != nil {
			return nil, fmt.Errorf("login failed: %w", err)
		}
	}

	return c, nil
}

func (c *Client) login(ctx context.Context) error {
	// UniFi OS login endpoint
	loginURL := strings.TrimRight(c.cfg.Host, "/") + "/api/auth/login"

	body, _ := json.Marshal(map[string]string{
		"username": c.cfg.Username,
		"password": c.cfg.Password,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login returned %d: %s", resp.StatusCode, string(respBody))
	}

	// Capture CSRF token from response header (required for mutations on UniFi OS)
	c.mu.Lock()
	if token := resp.Header.Get("X-Csrf-Token"); token != "" {
		c.csrfToken = token
	}
	c.loginGen++
	c.mu.Unlock()

	return nil
}

// Do executes an HTTP request against the UniFi controller's legacy API.
// The path should be relative (e.g., "api/s/default/stat/device").
// The /proxy/network prefix is added automatically for UniFi OS.
func (c *Client) Do(ctx context.Context, method, path string, payload interface{}) (json.RawMessage, error) {
	url := c.cfg.BaseURL() + "/" + strings.TrimLeft(path, "/")

	var bodyBytes []byte
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshaling payload: %w", err)
		}
		bodyBytes = data
	}

	newReq := func() (*http.Request, error) {
		var body io.Reader
		if bodyBytes != nil {
			body = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return nil, err
		}
		c.setHeaders(req)
		return req, nil
	}

	// Snapshot login generation before the request so we can detect concurrent re-logins.
	c.mu.RLock()
	genBefore := c.loginGen
	c.mu.RUnlock()

	resp, err := c.doWithRetry(newReq, 3)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized && c.cfg.AuthMethod() == config.AuthUserPass {
		// Session expired — single-flight re-login (coalesces concurrent 401s)
		if err := c.reloginIfNeeded(ctx, genBefore); err != nil {
			return nil, fmt.Errorf("re-login failed: %w", err)
		}
		req2, err := newReq()
		if err != nil {
			return nil, fmt.Errorf("creating retry request: %w", err)
		}
		resp2, err := c.httpClient.Do(req2)
		if err != nil {
			return nil, err
		}
		respBody, _ = io.ReadAll(resp2.Body)
		resp2.Body.Close()
		if resp2.StatusCode >= 400 {
			return nil, fmt.Errorf("request failed after re-login (%d): %s", resp2.StatusCode, string(respBody))
		}
	} else if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed (%d): %s", resp.StatusCode, string(respBody))
	}

	// Parse legacy API envelope.
	// Only extract .Data if the response actually has the legacy envelope
	// structure (meta.rc is present). Otherwise return the raw body to avoid
	// silently discarding non-legacy JSON responses.
	var envelope LegacyResponse
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		// Not valid JSON — return raw body
		return respBody, nil
	}
	if envelope.Meta.RC == "" {
		// No meta.rc field — not a legacy envelope, return full body
		return respBody, nil
	}
	if envelope.Meta.RC == "error" {
		// Include data array in error when available (often contains field-level validation details)
		if len(envelope.Data) > 0 && string(envelope.Data) != "[]" {
			return nil, fmt.Errorf("controller error: %s — %s", envelope.Meta.Msg, string(envelope.Data))
		}
		return nil, fmt.Errorf("controller error: %s", envelope.Meta.Msg)
	}

	return envelope.Data, nil
}

// DoRaw executes a request and returns the raw response body without envelope parsing.
// Use for Integration API or v2 endpoints. Handles 401 re-login for session auth.
func (c *Client) DoRaw(ctx context.Context, method, fullURL string, payload interface{}) (json.RawMessage, error) {
	var bodyBytes []byte
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshaling payload: %w", err)
		}
		bodyBytes = data
	}

	newReq := func() (*http.Request, error) {
		var body io.Reader
		if bodyBytes != nil {
			body = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
		if err != nil {
			return nil, err
		}
		c.setHeaders(req)
		return req, nil
	}

	// Snapshot login generation before the request so we can detect concurrent re-logins.
	c.mu.RLock()
	genBefore := c.loginGen
	c.mu.RUnlock()

	resp, err := c.doWithRetry(newReq, 3)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized && c.cfg.AuthMethod() == config.AuthUserPass {
		// Session expired — single-flight re-login (coalesces concurrent 401s)
		if err := c.reloginIfNeeded(ctx, genBefore); err != nil {
			return nil, fmt.Errorf("re-login failed: %w", err)
		}
		req2, err := newReq()
		if err != nil {
			return nil, fmt.Errorf("creating retry request: %w", err)
		}
		resp2, err := c.httpClient.Do(req2)
		if err != nil {
			return nil, err
		}
		body, _ = io.ReadAll(resp2.Body)
		resp2.Body.Close()
		if resp2.StatusCode >= 400 {
			return nil, fmt.Errorf("request failed after re-login (%d): %s", resp2.StatusCode, string(body))
		}
		return body, nil
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed (%d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// reloginIfNeeded performs a single-flight re-login. If another goroutine
// already refreshed the session (loginGen changed), the login is skipped.
// Returns nil if the session is now valid (either refreshed or already refreshed by another goroutine).
func (c *Client) reloginIfNeeded(ctx context.Context, genBefore uint64) error {
	c.loginMu.Lock()
	defer c.loginMu.Unlock()

	// Check if another goroutine already refreshed
	c.mu.RLock()
	currentGen := c.loginGen
	c.mu.RUnlock()

	if currentGen != genBefore {
		return nil // another goroutine already re-logged in
	}

	return c.login(ctx)
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.cfg.AuthMethod() == config.AuthAPIKey {
		req.Header.Set("X-API-Key", c.cfg.APIKey)
	}
	c.mu.RLock()
	token := c.csrfToken
	c.mu.RUnlock()
	if token != "" {
		req.Header.Set("X-Csrf-Token", token)
	}
}

func (c *Client) doWithRetry(newReq func() (*http.Request, error), maxRetries int) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*attempt) * 500 * time.Millisecond)
		}
		req, err := newReq()
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		// Retry on 5xx or 429
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			continue
		}
		return resp, nil
	}
	return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
}

// Site returns the configured site name.
func (c *Client) Site() string {
	return c.cfg.Site
}

// SitePath returns the legacy API site prefix.
func (c *Client) SitePath() string {
	return fmt.Sprintf("api/s/%s", c.cfg.Site)
}

// Config returns the underlying config.
func (c *Client) Config() *config.Config {
	return c.cfg
}
