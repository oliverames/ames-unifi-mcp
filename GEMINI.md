# ames-unifi-mcp

Go-based MCP server for UniFi Network controller management. 310 tools across 16 categories, lazy mode by default. Distributed via npm (`ames-unifi-mcp`), installed by `npx`. Published with `npm publish` from repo root (requires `package.json` version bump).

## Commands

| Task | Command |
|------|---------|
| Build | `make build` |
| Build all platforms | `make build-all` (darwin/linux × arm64/amd64 → `dist/`) |
| Test | `make test` (`go test -race -cover ./...`) |
| Lint | `make lint` (`golangci-lint run ./...`) |
| Clean | `make clean` |
| Docker | `make docker` (multi-arch buildx) |
| Research | `make research` (print API research sources before adding tools) |
| Release | `make build-all && npm version patch && npm publish` (bump version, cross-compile, publish) |

## Source Structure

```
cmd/ames-unifi-mcp/main.go         Entry point, server wiring
internal/
  config/                           Environment config loading
  client/                           HTTP client (session auth, auto re-login)
  permissions/                      Permission profile enforcement
  tools/
    metatools.go                    tool_index, tool_execute, tool_batch
    registry.go                     Tool registration and lazy/eager mode
    confirm.go                      Confirm gate (dry-run previews)
    core/                           Core API tools (devices, clients, firewall, stats, etc.)
    extended/                       Extended tools (hotspot, PoE, admin, syslog, etc.)
  version/                          Controller version auto-detection
```

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `UNIFI_HOST` | * | — | Controller URL (`https://192.168.1.1`) |
| `UNIFI_API_KEY` | * | — | API key (preferred, requires 9.1.105+) |
| `UNIFI_USERNAME` | * | — | Username (if no API key) |
| `UNIFI_PASSWORD` | * | — | Password (if no API key) |
| `UNIFI_SITE` | No | `default` | Site name |
| `UNIFI_VERIFY_SSL` | No | `true` | `false` for self-signed certs |
| `UNIFI_TOOL_MODE` | No | `lazy` | `lazy` (3 meta-tools) or `eager` (all 310) |
| `UNIFI_PERMISSION_PROFILE` | No | `standard` | `read-only`, `standard`, or `admin` |

\* `UNIFI_HOST` plus either `UNIFI_API_KEY` or `UNIFI_USERNAME`+`UNIFI_PASSWORD` are needed to actually call the controller. If absent, `config.Load()` sets `cfg.NeedsAuth = true` instead of hard-failing — server starts, registers tools, and every tool dispatch in `cmd/ames-unifi-mcp/main.go` short-circuits with `cfg.AuthHint()`. Plugin appears installed-but-inactive instead of erroring at startup. Same applies if the 1Password fallback (`op://Development/UniFi Controller/...`) returns nothing.

## Lazy Mode Architecture

Default mode exposes 3 meta-tools (`tool_index`, `tool_execute`, `tool_batch`) instead of all 310. Implementation in `internal/tools/metatools.go`. The registry in `registry.go` handles both modes — `UNIFI_TOOL_MODE=eager` registers all tools directly. Categories: `devices`, `clients`, `networks`, `wlan`, `firewall`, `stats`, `events`, `system`, `hotspot`, `poe`, `dpi`, `backup`, `settings`, `routing`, `vpn`, `qos`.

## Safety: Confirm Gate

All mutating tools require `"confirm": true` in the input. Without it, you get a **dry-run preview** showing what would happen. Always:

1. First call without confirm to see the preview
2. Explain the preview to the user
3. Call again with `"confirm": true` only after user approval

## Permission Awareness

Check which permission profile is active before attempting mutations. The profile determines what's available:

- **read-only**: Only read operations
- **standard**: Reads + safe mutations (WLAN, client management, network config). No PoE, system restarts, backup triggers, or firewall deletes.
- **admin**: Full access including PoE power cycling, system operations, and destructive actions

## Version-Dependent Features

The server auto-detects the controller version. Some tools are only available on newer firmware:

- **Zone-Based Firewall** (zones, policies): Network 9.0+
- **Integration API endpoints** (VPN servers, WiFi broadcasts): Network 9.0+
- **DNS Policies**: Network 10.0+
- **ACL Rules**: Network 10.0+
- **Switch Stacks, LAGs, MC-LAGs**: Network 10.0+
- **Traffic Matching Lists**: Network 10.0+
- **VPN Server/Tunnel CRUD** (create, update, delete): Network 10.1+

If a tool isn't in the index, the controller may be too old.

## MAC Address Format

Always use lowercase MACs (e.g., `aa:bb:cc:dd:ee:ff`). The controller expects this format.

## Adding New Tools

Before implementing a new tool, check `docs/api-research.md` and update it with findings from:
- https://developer.ui.com/
- https://ubntwiki.com/products/software/unifi-controller/api
- https://beez.ly/unifi-apis/

Tool handlers follow the pattern in `internal/tools/core/*.go` — each file covers one category.

## Undocumented Endpoints

Some tools use undocumented endpoints (marked in their descriptions). These work reliably on current firmware but behavior may change across versions. Examples: device locate (LED blink), per-client DPI stats, PoE power cycling.

## Gotchas

- **Self-signed certs**: Most UniFi controllers use self-signed TLS. Set `UNIFI_VERIFY_SSL=false` or requests will fail silently.
- **Makefile module path**: The `MODULE` var in the Makefile references `oliveames` (missing `r`). This doesn't affect builds but matters if Go module tooling uses it.
- **Session re-login thundering herd**: The client uses single-flight re-login to prevent parallel batch operations from all hitting the login endpoint simultaneously. If you see auth errors during batch calls, this is already handled — don't add extra retry logic.
- **API key vs session auth scope**: API key auth works with Legacy + Integration API. But some v2 API endpoints only accept session cookies — if a v2 tool fails with an API key, try username/password auth.
- **tool_batch parallelism**: `tool_batch` executes tools concurrently. Don't batch operations that depend on each other's results (e.g., create network → assign device to that network).
