# Worklog

## 2026-04-07 — Soft-fail on missing creds: needs-auth state replaces hard startup error

**What changed**: Refactored `internal/config/config.go` to detect missing credentials and set `cfg.NeedsAuth = true` instead of returning an error from `Load()`. Added `Config.AuthHint()` returning a user-facing configure-me message. Updated `internal/client/client.go` to skip the `login()` call when `NeedsAuth` (would otherwise crash on empty `cfg.Host`). Updated `cmd/ames-unifi-mcp/main.go` to skip controller version detection when `NeedsAuth`, and to wrap every tool dispatch (lazy meta-tools and eager all-tools paths) with a `cfg.NeedsAuth` check that short-circuits to `cfg.AuthHint()` as a structured `isError: true` MCP result before any client method runs. Added `authGate()` helper. Verified end-to-end via raw JSON-RPC stdio smoke test: server initializes cleanly, `tools/list` returns the three meta-tools, `tools/call` returns the auth hint as `isError`. All existing tests still pass. Updated `README.md` "Configuration" section and `CLAUDE.md` env vars table to document the new joint-optional credential contract and the soft-fail behavior. Bumped to v1.0.3, built all four platform binaries via `make build-all`, published `ames-unifi-mcp@1.0.3` to npm. Postpublish hook auto-bumped `ames-original-connectors` to 1.2.13 in the ames-claude marketplace.

**Decisions made**: Used the `NeedsAuth` flag pattern rather than introducing a separate "stub client" type — keeps the surface area minimal (4 files touched) and means existing code paths only need a single boolean check at dispatch time. Returned the auth hint as `isError: true` in the tool result envelope rather than as a JSON-RPC error: from Claude Code's perspective the server is healthy, the tool exists and ran, it just produced an error result — far cleaner UX than "MCP server crashed". For stdio MCPs there is no equivalent of the GitHub-style "△ needs authentication" indicator (that's reserved for HTTP/OAuth MCPs that emit a specific auth signal); "starts cleanly + tools error with actionable message" is the closest stdio analog. Bumped as patch (1.0.2 → 1.0.3) because the change is backward-compatible: existing properly-configured installs behave identically.

**Left off at**: Resolution of previous entry: the previous "1Password items still need to be created for Meta, Threads, Sprout, UniFi" left-off item is now PARTIALLY resolved for UniFi specifically — the server tolerates the missing item gracefully (needs-auth state) instead of erroring at startup. The 1Password item still needs to be created (`UniFi Controller` in Development vault, fields: host, api_key OR username + password) before any UniFi tool actually works, but it's no longer urgent. Other connectors (Meta, Threads, Sprout) have not been similarly upgraded yet. New: should the same soft-fail-with-NeedsAuth pattern be applied to imagerelay-mcp-server, meta-mcp-server, and sprout-mcp-server for consistency? They currently rely on the simpler "spawn fails fast if creds absent" pattern which is OK now that the connectors `.mcp.json` no longer pre-validates env vars (they'll spawn and error at first API call instead of at startup), but the UniFi pattern is more user-friendly.

**Open questions**: For new MCP servers gaining self-resolution capabilities, should the soft-fail/needs-auth pattern be the default? It's a few extra lines per server but the UX is meaningfully better. Part of a broader session that also touched ames-claude.

---

## 2026-04-06 — 1Password CLI fallback for credential resolution

**What changed**: Added automatic 1Password CLI fallback to credential resolution at startup. When environment variables are not set, the server attempts to resolve them via `op read` from the Development vault before failing. Uses `execFileSync` (Node) or `exec.Command` (Go) for shell-safe execution with a 10s timeout. Silent no-op if 1Password CLI is unavailable. Updated README to document the integration with `op://` reference paths. Part of a broader session that also touched ynab-mcp-server, imagerelay-mcp-server, meta-mcp-server, sprout-mcp-server, and ames-unifi-mcp.

**Decisions made**: Used `execFileSync` instead of `execSync` to avoid shell injection surface (even though inputs are hardcoded string literals). Added the fallback as a separate `op-fallback.ts` module (TS servers) or inline helper (Go) rather than modifying the existing auth flow, keeping the env var path as primary (zero overhead) and 1Password as fallback only. Chose `op://Development/` vault paths matching existing 1Password item names where items exist; for servers without items yet (Meta, Sprout, UniFi), chose conventional names so items can be created later.

**Left off at**: Published and pushed. 1Password items still need to be created for Meta Access Token, Threads Access Token, Sprout API Token/OAuth Client, and UniFi Controller credentials. YNAB and ImageRelay items already exist. Also: 20 uncategorized YNAB transactions from this session's review were identified but not yet categorized.

**Open questions**: None.

---



## 2026-03-22 — Security audit across all public repos, Mistral key rotation

**What changed**: Ran parallel Opus agents to audit all 8 public repos for leaked secrets. Found and fixed: Mistral API key hardcoded in ames-consulting (merged PR #11), `.claude/settings.json` committed in sunshine-trail and ping-warden (removed), hardcoded YNAB budget UUID in test.js (replaced with env var), credential storage paths leaked in ynab WORKLOG (redacted), missing LICENSE in ynab (added). Rotated compromised Mistral API key and updated it in settings.json, credentials/env.json, voicemate xcconfig, and 60+ backup archive files. Also created `readme-style` skill in ames-skills codifying Oliver's README conventions.

**Decisions made**:
- Kept `.claude/rules/` in ames-consulting repo (useful coding conventions for contributors) but removed `settings.json` (IDE config that shouldn't be public)
- Updated old Mistral key in backup archive files too — even though they're private, the old key appearing anywhere is a liability
- Created the readme-style skill as a standalone skill rather than adding it to wrap-up — it's a creative/design concern, not a session-end checklist item

**Left off at**: All 8 public repos are audited and clean. Next session should:
1. Run `gh secret-scanning` or BFG Repo-Cleaner on repos with secrets in git history (ames-consulting Mistral key, ynab budget UUID, sunshine-trail old password)
2. Audit meta-mcp-server git history (current HEAD is clean but history wasn't scanned due to Bash sandbox limitations)
3. Package ames-unifi-mcp as DXT extension for Claude Desktop/Cowork
4. Consider adding a `repo-audit` skill that automates the secret scanning workflow

**Open questions**:
- Should we use BFG Repo-Cleaner to scrub old secrets from git history, or accept that rotating the keys is sufficient?
- The protect-secrets hook blocks legitimate audit commands — should it have an override for audit workflows?

---

## 2026-03-22 — Branding, SEO, GitHub housekeeping

**What changed**: Applied branded README styling across all 8 public repos — amber badge colors (#f5a542), Buy Me a Coffee links (badge row + footer), ames.consulting footer with social links. Maxed all repos to 20 GitHub topics for SEO. Set homepage URLs (npm links for MCP servers, live sites for web projects). Created `readme-style` skill in ames-skills to codify these conventions. Fixed GitHub auth: removed stale `GITHUB_TOKEN` from settings.json so `gh` uses its keyring token (which has `delete_repo` scope). Made archived repos (geforce-now-launcher, get-filenames) private. Deleted nanoclaw. Made MCP servers (imagerelay, sprout, ynab, unifi) public. Updated wrap-up skill to v2.2.0 with README staleness checks and npm publish detection.

**Decisions made**:
- Used amber #f5a542 (from ames.consulting brand) as the unifying color for all GitHub badge shields across repos
- Chose `flat-square` style for header badges, `for-the-badge` for the larger footer Buy Me a Coffee button
- Applied footer consistently: "Built by Oliver Ames in Vermont" + GitHub/LinkedIn/Bluesky links
- Secret audit agents couldn't run due to Bash sandbox restrictions on subagents — needs different approach next session

**Left off at**: Secret audit across all 8 public repos was attempted but blocked by subagent sandbox permissions. Next session should:
1. Run the secret/cleanup audit manually (not via subagents) — scan git history for leaked keys, .env files, internal IPs
2. Package as DXT extension for Claude Desktop/Cowork (icon.png is ready)
3. Consider GitHub Actions for auto-publishing npm on git tags

**Open questions**:
- Should subagent audit tasks use a different permission model? The `block-dangerous-commands` hook prevented git history scanning
- Are there secrets in git history that need BFG Repo-Cleaner treatment?

---

## 2026-03-22 — NPM publish + icon

**What changed**: Published `ames-unifi-mcp@1.0.0` to npm. Created package.json with postinstall script that copies the correct platform binary (darwin/linux, amd64/arm64) to `bin/`. Added .npmignore to exclude Go source from the npm tarball. Added MIT LICENSE. Added UniFi icon (icon.png) for future DXT packaging.

**Decisions made**:
- Used the postinstall-copies-binary pattern (like esbuild/turbo) rather than per-platform npm packages (@ames-unifi-mcp/darwin-arm64 etc.) — simpler, single package, 11.8 MB compressed for all 4 platforms
- Kept `dist/` in .gitignore but ship it in the npm package via the `files` field — binaries are build artifacts, not source
- Named the npm package `ames-unifi-mcp` matching the binary name for simplicity

**Left off at**: Package is live on npm. Next session should:
1. Package as a DXT extension for Claude Desktop/Cowork (icon.png is ready)
2. Test `npm install -g ames-unifi-mcp` on a clean machine to verify the postinstall flow
3. Consider adding Windows support (GOOS=windows) if there's demand
4. Test against a real Dream Machine to validate the full tool set

**Open questions**:
- Should the npm package include a `--version` flag in the binary for debugging?
- DXT packaging: does it need a different manifest format than the MCP server config block?
- Should we set up GitHub Actions to auto-publish new npm versions on git tags?

---

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
