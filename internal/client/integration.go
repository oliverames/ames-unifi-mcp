package client

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/oliverames/ames-unifi-mcp/internal/config"
)

var siteUUID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// DoIntegrationSite addresses a resource using the Integration site's UUID.
// Legacy tools continue using the configured short name through Site/SitePath.
func (c *Client) DoIntegrationSite(ctx context.Context, method, path string, payload interface{}) (json.RawMessage, error) {
	id, err := c.IntegrationSiteID(ctx)
	if err != nil {
		return nil, err
	}
	return c.DoRaw(ctx, method, c.cfg.BaseURL()+"/integration/v1/sites/"+id+path, payload)
}

// IntegrationSiteID resolves and caches the configured legacy site reference.
// Failed resolutions are not cached, so corrected credentials/sites can recover.
func (c *Client) IntegrationSiteID(ctx context.Context) (string, error) {
	if c.cfg.AuthMethod() != config.AuthAPIKey {
		return "", fmt.Errorf("integration API requires UNIFI_API_KEY; username/password sessions are not supported")
	}
	c.siteMu.Lock()
	defer c.siteMu.Unlock()
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if c.integrationSiteID != "" {
		return c.integrationSiteID, nil
	}
	// Bound discovery even if a malformed controller response never terminates.
	for offset, pages := 0, 0; pages < 100; pages++ {
		body, err := c.DoRaw(ctx, "GET", fmt.Sprintf("%s/integration/v1/sites?offset=%d&limit=200", c.cfg.BaseURL(), offset), nil)
		if err != nil {
			return "", fmt.Errorf("resolving Integration site: %w", err)
		}
		var page struct {
			TotalCount int `json:"totalCount"`
			Data       []struct {
				ID                string `json:"id"`
				InternalReference string `json:"internalReference"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return "", fmt.Errorf("parsing Integration sites: %w", err)
		}
		for _, site := range page.Data {
			if site.InternalReference == c.cfg.Site {
				if !siteUUID.MatchString(site.ID) {
					return "", fmt.Errorf("integration site %q returned an invalid UUID", c.cfg.Site)
				}
				c.integrationSiteID = site.ID
				return site.ID, nil
			}
		}
		offset += len(page.Data)
		if len(page.Data) == 0 || offset >= page.TotalCount {
			return "", fmt.Errorf("integration site matching UNIFI_SITE %q was not found", c.cfg.Site)
		}
	}
	return "", fmt.Errorf("integration site discovery exceeded 100 pages")
}
