# Worklog

## 2026-03-22 — Deep review: 7 bugs fixed, 39 tools added (310 total), README rewritten

**What changed**: Ran a comprehensive Ralph Loop review (10+ iterations) covering bug hunting and API coverage auditing. Fixed 7 bugs in client.go: DoRaw missing 401 re-login (all Integration/Cloud API calls would fail on session expiry), csrfToken data race across concurrent batch goroutines (added sync.RWMutex), HTTP response body leak on 401 retry (defer captured variable by reference), thundering herd on concurrent 401s (added single-flight re-login with generation counter), stats_authorization using string timestamps instead of integers, duplicate system_vpn_list, dead code. Enhanced controller error messages to include data array details. Added 39 new tools: traffic routes CRUD (v2 API), VPN server/tunnel CRUD (Integration API 10.1+), switching stacks/LAGs/MC-LAGs CRUD (10.0+), device_upgrade_all, wlan_enable/disable, backup restore, hotspot operator/2.0/package CRUD, scheduled task CRUD. Rewrote README with full 310-tool coverage breakdown, architecture docs, and usage examples.

**Decisions made**:
- Used generation-counter single-flight pattern for concurrent re-login rather than sync.Once (generation counters allow re-login after the next session expiry, while Once would only login once ever)
- Changed from defer resp.Body.Close() to immediate close-after-read pattern to eliminate body leaks on 401 retry path
- Added WLAN enable/disable as convenience tools rather than forcing LLM to construct full wlan_update config objects
- Included data array in controller error messages (field-level validation details were being silently discarded)

**Left off at**: The codebase is in good shape. Next session should:
1. Publish as an NPM package for easy distribution (user's next request)
2. Test against a real Dream Machine — still untested against live hardware
3. Add integration tests or mock-based unit tests for tool handlers and client retry logic
4. Consider whether Cloud API needs separate auth config (local vs cloud API keys)
5. Fix the go.mod module path if needed (oliveames vs oliverames)

**Open questions**:
- Does the same API key work for both local controller and api.ui.com cloud endpoints?
- Should we add WebSocket support for real-time event streaming (would require a different MCP pattern — resources or notifications)?
- The go.mod module path uses `github.com/oliveames/ames-unifi-mcp` — needs verification against actual repo

---

## 2026-03-22 — Initial build: 197-tool UniFi MCP server from scratch

**What changed**: Built the entire ames-unifi-mcp server from zero. Phase 0 research scraped 10 sources (official API, community wiki, OpenAPI specs, 4 existing MCP servers) into a 600+ line docs/api-research.md. Implemented the full Go architecture: HTTP client with API key + session auth, lazy/eager tool registration via mcp-go SDK, confirm gate on mutations, 3-tier permission system, and controller version detection. Wrote 197 tools across 20 files covering every documented endpoint — devices, clients, networks, WLANs, firewall (legacy + ZBF), ACL rules, DNS policies, traffic rules, stats/DPI/reports, events/alarms, system/settings/backups, WAN, switching, PoE, hotspot/vouchers, Cloud Site Manager API, admin/site management, v2 system logs, AP groups, and misc endpoints.

**Decisions made**:
- Scoped to UniFi OS only (Dream Machine) — no standalone controller support. Always uses /proxy/network prefix.
- Used BaseTool struct with inline Handler func pattern rather than code generation (simpler to extend, unlike claytono's reflection-based approach).
- Chose mcp-go SDK (mark3labs) for MCP protocol, same as claytono's server.
- Put all 42 settings categories behind a single generic `system_setting_update` tool + convenience wrappers (LED toggle, IPS) rather than 42 individual tools.
- Cloud API tools point directly at api.ui.com — relies on same API key working for both local and cloud (may need separate auth path).

**Left off at**: Two background review agents were running (bug review + coverage audit) but session ended before they reported. Next session should:
1. Run `go vet ./...` and `golangci-lint run` to catch any static analysis issues
2. Fix the known bug in client.go: request body is consumed on first attempt and can't be replayed on retry/re-login (need to buffer the body bytes)
3. Test against a real Dream Machine — the client, auth flow, and UniFi OS path detection are all untested
4. Consider whether Cloud API needs separate auth config (cloud API key vs local API key may differ)
5. Add integration tests or at least mock-based unit tests for the tool handlers

**Open questions**:
- Does the same API key work for both local controller and api.ui.com cloud endpoints?
- Should we use go-unifi library instead of raw HTTP for the legacy endpoints (more type safety, but adds dependency complexity)?
- The go.mod module path uses `github.com/oliveames/ames-unifi-mcp` but the actual repo is `github.com/oliverames/ames-unifi-mcp` — need to verify and potentially fix the module path

---
