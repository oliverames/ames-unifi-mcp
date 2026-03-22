# ames-unifi-mcp

A Go-based MCP server for UniFi Network management. Designed for personal use with a Dream Machine (UniFi OS) and the unifi.ui.com cloud interface.

## Features

- **60+ tools** covering devices, clients, networks, WLANs, firewall (legacy + ZBF), stats, DPI, events, hotspot, PoE, backups, and more
- **Lazy mode** (default): 3 meta-tools (~200 tokens of context). The LLM queries `tool_index`, then dispatches via `tool_execute` or `tool_batch`.
- **Confirm gate**: All mutating operations require `confirm: true` or return a dry-run preview
- **Permission profiles**: `read-only`, `standard`, `admin`
- **Version detection**: Automatically detects controller version and gates features (ZBF requires 9.0+, DNS policies require 10.0+)
- **UniFi OS native**: Always uses `/proxy/network` prefix — built for Dream Machine

## Quick Start

```json
{
  "mcpServers": {
    "ames-unifi": {
      "command": "ames-unifi-mcp",
      "env": {
        "UNIFI_HOST": "https://192.168.1.1",
        "UNIFI_API_KEY": "your-api-key",
        "UNIFI_SITE": "default",
        "UNIFI_VERIFY_SSL": "false",
        "UNIFI_PERMISSION_PROFILE": "standard"
      }
    }
  }
}
```

## Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `UNIFI_HOST` | Yes | — | Controller URL (e.g., `https://192.168.1.1`) |
| `UNIFI_API_KEY` | Yes* | — | API key (preferred). Generate at Settings > Control Plane > Integrations |
| `UNIFI_USERNAME` | Yes* | — | Username (fallback if no API key) |
| `UNIFI_PASSWORD` | Yes* | — | Password (fallback if no API key) |
| `UNIFI_SITE` | No | `default` | Site name |
| `UNIFI_VERIFY_SSL` | No | `true` | Set `false` for self-signed certs |
| `UNIFI_TOOL_MODE` | No | `lazy` | `lazy` (3 meta-tools) or `eager` (all tools) |
| `UNIFI_PERMISSION_PROFILE` | No | `standard` | `read-only`, `standard`, or `admin` |
| `UNIFI_LOG_LEVEL` | No | `error` | `debug`, `info`, `warn`, `error` |

*Either `UNIFI_API_KEY` or both `UNIFI_USERNAME`/`UNIFI_PASSWORD` are required.

## Building

```bash
make build         # Build for current platform
make build-all     # Build for darwin/linux, amd64/arm64
make test          # Run tests
make lint          # golangci-lint
make docker        # Build multi-arch Docker image
```

## Tool Categories

| Category | Tools | Description |
|----------|-------|-------------|
| devices | 9 | List, get, restart, adopt, locate, upgrade, provision |
| clients | 7 | List active/all, get, block, unblock, reconnect, forget |
| networks | 5 | CRUD for network configurations |
| wlan | 5 | CRUD for wireless networks |
| firewall | 9 | Legacy rules/groups + ZBF zones/policies (9.0+) |
| stats | 8 | Health, sysinfo, dashboard, DPI, speedtest, routing |
| events | 5 | Events, alarms (list, count, archive) |
| system | 10 | Sites, settings, admins, backups, firmware, VPN, DNS |
| hotspot | 4 | Guest auth, vouchers, session management |
| poe | 1 | PoE port power cycling (undocumented) |

## Adding New Tools

See `internal/tools/extended/README.md`. The extension layer makes it trivial to add tools as new endpoints are discovered.

## Architecture

```
cmd/ames-unifi-mcp/main.go     Entry point, wires everything together
internal/config/                Environment config loading
internal/client/                HTTP client (auth, retry, UniFi OS paths)
internal/version/               Controller version detection
internal/permissions/           Permission profiles (read-only/standard/admin)
internal/tools/                 Tool interface, registry, confirm gate, meta-tools
internal/tools/core/            Core tools (devices, clients, networks, etc.)
internal/tools/extended/        Extension tools (PoE, hotspot, etc.)
docs/api-research.md            Living API reference (600+ lines)
```
