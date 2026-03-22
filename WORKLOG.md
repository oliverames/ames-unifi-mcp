# Worklog

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
