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
	"time"

	"github.com/oliveames/ames-unifi-mcp/internal/config"
)

// Client handles HTTP communication with the UniFi controller.
// Targets UniFi OS (Dream Machine) — always uses /proxy/network prefix.
type Client struct {
	cfg        *config.Config
	httpClient *http.Client
	csrfToken  string
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
	if token := resp.Header.Get("X-Csrf-Token"); token != "" {
		c.csrfToken = token
	}

	return nil
}

// Do executes an HTTP request against the UniFi controller's legacy API.
// The path should be relative (e.g., "api/s/default/stat/device").
// The /proxy/network prefix is added automatically for UniFi OS.
func (c *Client) Do(ctx context.Context, method, path string, payload interface{}) (json.RawMessage, error) {
	url := c.cfg.BaseURL() + "/" + strings.TrimLeft(path, "/")

	var bodyReader io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshaling payload: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.doWithRetry(req, 3)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized && c.cfg.AuthMethod() == config.AuthUserPass {
		// Session expired — re-login and retry once
		if err := c.login(ctx); err != nil {
			return nil, fmt.Errorf("re-login failed: %w", err)
		}
		req2, _ := http.NewRequestWithContext(ctx, method, url, bodyReader)
		c.setHeaders(req2)
		resp2, err := c.httpClient.Do(req2)
		if err != nil {
			return nil, err
		}
		defer resp2.Body.Close()
		respBody, _ = io.ReadAll(resp2.Body)
		if resp2.StatusCode >= 400 {
			return nil, fmt.Errorf("request failed after re-login (%d): %s", resp2.StatusCode, string(respBody))
		}
		resp = resp2
	} else if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed (%d): %s", resp.StatusCode, string(respBody))
	}

	// Parse legacy API envelope
	var envelope LegacyResponse
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		// Not a legacy response — return raw body (e.g., v2 API)
		return respBody, nil
	}
	if envelope.Meta.RC == "error" {
		return nil, fmt.Errorf("controller error: %s", envelope.Meta.Msg)
	}

	return envelope.Data, nil
}

// DoRaw executes a request and returns the raw response body without envelope parsing.
// Use for Integration API or v2 endpoints.
func (c *Client) DoRaw(ctx context.Context, method, fullURL string, payload interface{}) (json.RawMessage, error) {
	var bodyReader io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshaling payload: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)

	resp, err := c.doWithRetry(req, 3)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed (%d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.cfg.AuthMethod() == config.AuthAPIKey {
		req.Header.Set("X-API-Key", c.cfg.APIKey)
	}
	if c.csrfToken != "" {
		req.Header.Set("X-Csrf-Token", c.csrfToken)
	}
}

func (c *Client) doWithRetry(req *http.Request, maxRetries int) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*attempt) * 500 * time.Millisecond)
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
