<p align="center">
  <img src="assets/icon.png" width="80" height="80" alt="UniFi">
</p>

<h1 align="center">UniFi MCP Server</h1>

<p align="center">
  <strong>Complete UniFi Network controller management through the Model Context Protocol</strong>
</p>

<p align="center">
  <code>310 tools</code> &nbsp;&bull;&nbsp;
  <code>16 categories</code> &nbsp;&bull;&nbsp;
  <code>3 API layers</code> &nbsp;&bull;&nbsp;
  <code>Network 5.x &ndash; 10.x</code>
</p>

<p align="center">
  <a href="https://www.npmjs.com/package/ames-unifi-mcp"><img src="https://img.shields.io/npm/v/ames-unifi-mcp?style=flat-square&color=f5a542" alt="npm"></a>
  <a href="https://github.com/oliverames/ames-unifi-mcp/releases/tag/v1.0.3"><img src="https://img.shields.io/github/v/release/oliverames/ames-unifi-mcp?style=flat-square&color=f5a542&label=MCPB" alt="MCPB release"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-f5a542?style=flat-square" alt="License"></a>
  <a href="https://www.buymeacoffee.com/oliverames"><img src="https://img.shields.io/badge/Buy_Me_a_Coffee-support-f5a542?style=flat-square&logo=buy-me-a-coffee&logoColor=white" alt="Buy Me a Coffee"></a>
</p>

<p align="center">
  <a href="#quick-start">Quick Start</a> &nbsp;&bull;&nbsp;
  <a href="#install-with-mcpb">MCPB Download</a> &nbsp;&bull;&nbsp;
  <a href="#tool-coverage">Tool Coverage</a> &nbsp;&bull;&nbsp;
  <a href="#architecture">Architecture</a> &nbsp;&bull;&nbsp;
  <a href="#configuration">Configuration</a>
</p>

---

A Go-based MCP server that gives AI assistants deep, safe access to UniFi Network controllers. Built for UniFi OS devices (Dream Machine, Cloud Gateway) and the `unifi.ui.com` cloud interface.

## Why This Exists

Managing a UniFi network through natural language means your AI assistant needs to understand every corner of the controller API. This server exposes **310 tools** spanning the complete API surface &mdash; from basic device listing to zone-based firewall policy ordering, from hotspot voucher generation to MC-LAG domain management.

Every mutating operation passes through a **confirm gate** that returns a dry-run preview before execution. The assistant sees exactly what will change and asks you before proceeding.

## Quick Start

### Install with MCPB

For Claude Desktop and other MCPB-compatible clients, download the local bundle from the [v1.0.3 release](https://github.com/oliverames/ames-unifi-mcp/releases/tag/v1.0.3):

[Download `ames-unifi-mcp-1.0.3.mcpb`](https://github.com/oliverames/ames-unifi-mcp/releases/download/v1.0.3/ames-unifi-mcp-1.0.3.mcpb)

The bundle includes the UniFi favicon, production runtime binaries for macOS and Linux, and setup prompts for host, authentication, site, SSL, tool mode, and permission profile.

Add to your MCP client configuration:

```json
{
  "mcpServers": {
    "unifi": {
      "command": "ames-unifi-mcp",
      "env": {
        "UNIFI_HOST": "https://192.168.1.1",
        "UNIFI_API_KEY": "your-api-key-here",
        "UNIFI_SITE": "default",
        "UNIFI_VERIFY_SSL": "false"
      }
    }
  }
}
```

Generate an API key at **Settings &rarr; Control Plane &rarr; Integrations** (requires Network 9.1.105+). For older firmware, use username/password authentication instead.

## How It Works

### Lazy Mode (Default)

Instead of flooding the context window with 310 tool definitions, the server exposes just **3 meta-tools**:

| Meta-Tool | Purpose |
|-----------|---------|
| `tool_index` | Browse the tool catalog, optionally filtered by category |
| `tool_execute` | Run any tool by name with input parameters |
| `tool_batch` | Run multiple tools in parallel for efficiency |

The assistant discovers tools on demand, keeping context lean (~200 tokens vs ~30,000 in eager mode).

### Confirm Gate

Every write operation requires explicit confirmation. Without `confirm: true`, the tool returns a dry-run preview showing what *would* happen:

```
You:       "Restart the office access point"
Assistant:  Let me preview that first.
            → device_restart {"mac": "aa:bb:cc:dd:ee:ff"}
            ← Preview: This would restart "Office AP" (soft reboot). Confirm?
You:       "Yes"
Assistant:  → device_restart {"mac": "aa:bb:cc:dd:ee:ff", "confirm": true}
            ← Device restarting.
```

### Permission Profiles

Control what the assistant can do:

| Profile | Capabilities |
|---------|-------------|
| **read-only** | Query everything, change nothing |
| **standard** | Reads + safe mutations (WLAN, clients, networks). No PoE cycling, system restarts, or firewall deletes |
| **admin** | Full access including destructive and system-level operations |

Permission-denied tools still appear in `tool_index` (marked `[PERMISSION DENIED]`) so the assistant can explain what's unavailable and why.

### Version Detection

On startup, the server queries the controller version and automatically gates tools by firmware requirements:

| Feature | Minimum Version |
|---------|----------------|
| Legacy API (full) | 5.x+ |
| Zone-Based Firewall | Network 9.0+ |
| Integration API | Network 9.0+ |
| API Key Authentication | Network 9.1.105+ |
| DNS Policies, ACL Rules | Network 10.0+ |
| Switch Stacks, LAGs, MC-LAGs | Network 10.0+ |
| VPN Server/Tunnel CRUD | Network 10.1+ |

If a tool isn't in the index, the controller firmware is too old to support it.

---

## Tool Coverage

### 310 tools across 16 categories

<table>
<tr>
<td width="50%" valign="top">

**Devices** &mdash; 29 tools
```
device_list          device_restart
device_get           device_adopt
device_upgrade       device_upgrade_all
device_locate_on/off device_force_provision
device_spectrum_scan device_rolling_upgrade_*
device_migrate       device_port_action
device_list_v2       device_stats_latest
device_unadopt       device_pending_list
```

**Clients** &mdash; 15 tools
```
client_list_active   client_block/unblock
client_get           client_reconnect
client_forget        client_sessions
client_rename        client_update
client_list_v2       client_action
```

**Networks** &mdash; 10 tools
```
network_list    network_create
network_get     network_update
network_delete  network_*_v2 (Integration API)
```

**Wireless (WLAN + WiFi)** &mdash; 12 tools
```
wlan_list       wlan_create/update/delete
wlan_enable     wlan_disable
wifi_broadcast_list/get/create/update/delete
```

**Firewall** &mdash; 23 tools
```
firewall_rule_*      (legacy CRUD)
firewall_group_*     (address/port groups)
firewall_zone_*      (ZBF zones, 9.0+)
firewall_policy_*    (ZBF policies + ordering)
```

**ACL Rules** &mdash; 7 tools
```
acl_rule_list/get/create/update/delete
acl_rule_ordering_get/set
```

**DNS Policies** &mdash; 5 tools
```
dns_policy_list/get/create/update/delete
```

**Traffic & QoS** &mdash; 15 tools
```
traffic_rule_*           (v2 API rules)
traffic_route_*          (v2 API routes)
traffic_matching_list_*  (10.0+)
```

</td>
<td width="50%" valign="top">

**VPN** &mdash; 10 tools
```
vpn_server_list/get/create/update/delete
vpn_tunnel_list/get/create/update/delete
```

**Switching** &mdash; 15 tools
```
switching_stack_*    (switch stacks)
switching_lag_*      (link aggregation)
switching_mclag_*    (MC-LAG domains)
```

**Stats & DPI** &mdash; 24 tools
```
stats_site_health    stats_sysinfo
stats_dashboard      stats_report
stats_speedtest_*    stats_ips_events
stats_rogueap        stats_spectrumscan
stats_dpi_site       stats_dpi_client
stats_dpi_apps       stats_dpi_categories
```

**Events & Alarms** &mdash; 5 tools
```
event_list           alarm_list
alarm_count          alarm_archive
alarm_archive_all
```

**Hotspot & Guests** &mdash; 13 tools
```
hotspot_authorize/unauthorize_guest
hotspot_create_voucher  hotspot_extend
hotspot_voucher_*_v2    (Integration API)
hotspot_config          hotspot_packages
```

**System & Settings** &mdash; 53 tools
```
system_reboot/poweroff  system_settings
system_backup_*         system_firmware_*
system_port_forward_*   system_static_route_*
system_usergroup_*      system_portprofile_*
system_dhcpoption_*     system_radiusprofile_*
system_tag_*            system_led_toggle
```

**Cloud API** &mdash; 9 tools
```
cloud_host_list/get     cloud_site_list
cloud_device_list       cloud_isp_metrics
cloud_sdwan_*
```

**Admin & Misc** &mdash; 56 tools
```
admin_site_create/delete/rename
admin_invite/revoke/grant/update
poe_power_cycle         syslog_query
apgroup_*               misc_rogueknown_*
misc_scheduletask_*     misc_hotspotop_*
misc_dpigroup_*         misc_cnt_resource
```

</td>
</tr>
</table>

### API Layer Coverage

The server covers all three UniFi API layers:

| API Layer | Path Prefix | Auth | Coverage |
|-----------|-------------|------|----------|
| **Legacy API** | `/api/s/{site}/...` | Session cookie | Full &mdash; stat, cmd, rest, set, get, upd, list, cnt, guest, dl |
| **v2 API** | `/v2/api/site/{site}/...` | Session cookie | Full &mdash; traffic rules, traffic routes, AP groups, system logs |
| **Integration API** | `/integration/v1/...` | API key or session | Full &mdash; devices, clients, networks, WiFi, firewall, VPN, switching, hotspot, DPI |
| **Cloud Site Manager** | `api.ui.com/v1/...` | API key | Full &mdash; hosts, sites, devices, ISP metrics, SD-WAN |

---

## Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `UNIFI_HOST` | * | &mdash; | Controller URL (`https://192.168.1.1`) |
| `UNIFI_API_KEY` | * | &mdash; | API key (preferred, requires 9.1.105+) |
| `UNIFI_USERNAME` | * | &mdash; | Username (if no API key) |
| `UNIFI_PASSWORD` | * | &mdash; | Password (if no API key) |
| `UNIFI_SITE` | No | `default` | Site name |
| `UNIFI_VERIFY_SSL` | No | `true` | `false` for self-signed certs |
| `UNIFI_TOOL_MODE` | No | `lazy` | `lazy` (3 meta-tools) or `eager` (all 310 tools) |
| `UNIFI_PERMISSION_PROFILE` | No | `standard` | `read-only`, `standard`, or `admin` |

<sub>* `UNIFI_HOST` plus either `UNIFI_API_KEY` or both `UNIFI_USERNAME` + `UNIFI_PASSWORD` are required for the server to actually call the controller. If they are absent, the server still starts and registers all tools (so the plugin appears installed), but every tool call returns a structured "credentials not configured" error pointing the user at the env vars or 1Password fallback. This lets the connector live in a clean "needs authentication" state instead of a hard startup error.</sub>

### 1Password fallback

If env vars are unset, the server falls back to 1Password CLI (`op read`) using these references:

- `op://Development/UniFi Controller/host`
- `op://Development/UniFi Controller/api_key`
- `op://Development/UniFi Controller/username`
- `op://Development/UniFi Controller/password`

Create a `UniFi Controller` item in your `Development` vault with those fields and the server will resolve credentials at startup with no env vars required. Resolution happens in `internal/config/config.go:opRead` and requires either an interactive `op` session or `OP_SERVICE_ACCOUNT_TOKEN` in the environment.

### Authentication Methods

**API Key** (recommended) &mdash; Generate at Settings &rarr; Control Plane &rarr; Integrations. Supports both legacy and Integration API endpoints. No session management overhead.

**Username/Password** &mdash; Uses session cookies with automatic re-login on expiry. The server includes a single-flight re-login mechanism that prevents thundering-herd issues when batch operations encounter session timeouts simultaneously.

### 1Password Integration

If credentials are not set in the environment, the server automatically attempts to resolve them from [1Password CLI](https://developer.1password.com/docs/cli/):

| Variable | 1Password Reference |
|----------|-------------------|
| `UNIFI_HOST` | `op://Development/UniFi Controller/host` |
| `UNIFI_API_KEY` | `op://Development/UniFi Controller/api_key` |
| `UNIFI_USERNAME` | `op://Development/UniFi Controller/username` |
| `UNIFI_PASSWORD` | `op://Development/UniFi Controller/password` |

This means you can skip setting env vars entirely if you have `op` installed and a service account or session active. The fallback adds ~1-2s to startup and is silently skipped if 1Password is unavailable.

---

## Architecture

```
cmd/ames-unifi-mcp/main.go          Entry point, server wiring
internal/
  config/                            Environment config loading
  client/                            HTTP client
    - Session auth with auto re-login (single-flight)
    - API key auth (X-API-Key header)
    - CSRF token management (thread-safe)
    - Retry with backoff (429, 5xx)
    - Legacy envelope parsing (meta.rc/data)
    - Raw response passthrough (Integration/v2 APIs)
  version/                           Controller version detection
  permissions/                       Permission profiles
  tools/
    tool.go                          Tool interface
    registry.go                      Tool registry with version/permission gating
    confirm.go                       Confirm gate (dry-run preview pattern)
    metatools.go                     tool_index, tool_execute, tool_batch
    core/                            Core tool implementations
      devices.go    clients.go    networks.go
      wlan.go       wifi.go       firewall.go
      acl.go        dns.go        traffic.go
      wan.go        switching.go  stats.go
      events.go     system.go
    extended/                        Extended tool implementations
      poe.go        hotspot.go    cloud.go
      admin.go      syslog.go     apgroups.go
      misc.go
```

### Key Design Decisions

**Lazy mode by default.** An LLM calling 310 tools directly wastes context and confuses tool selection. The 3-meta-tool pattern lets the assistant discover tools on demand, typically using < 200 tokens of context for the tool definitions.

**Confirm gate on all mutations.** Rather than relying on the MCP client to prevent unintended actions, every mutating tool returns a preview by default. The `confirm: true` parameter is an explicit opt-in. This is baked into the tool schema &mdash; the LLM sees it as a required step, not an optional flag.

**Version gating at registration.** Tools for newer API features aren't hidden or errored &mdash; they simply don't register if the controller is too old. The tool index only shows what's actually available.

**Permission gating with visibility.** Denied tools appear in the index with a `[PERMISSION DENIED]` suffix. This lets the assistant explain to the user why something isn't available, rather than returning a cryptic "unknown tool" error.

---

## Building

```bash
go build -o ames-unifi-mcp ./cmd/ames-unifi-mcp/
```

### Cross-compilation

```bash
# Linux ARM64 (e.g., Raspberry Pi, Docker on NAS)
GOOS=linux GOARCH=arm64 go build -o ames-unifi-mcp ./cmd/ames-unifi-mcp/

# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o ames-unifi-mcp ./cmd/ames-unifi-mcp/
```

### Running Tests

```bash
go test ./...
```

---

## Common Operations

Here's what natural-language network management looks like:

**"How's my network doing?"**
```
→ tool_batch: stats_site_health + client_list_active + alarm_count
← WAN: healthy, 47 clients connected, 0 active alarms
```

**"Block that sketchy device"**
```
→ client_get {"mac": "aa:bb:cc:dd:ee:ff"}
← Device "Unknown-IoT" on VLAN 30, 2.3 GB today
→ client_block {"mac": "aa:bb:cc:dd:ee:ff"}
← Preview: Would block Unknown-IoT. Confirm?
→ client_block {"mac": "aa:bb:cc:dd:ee:ff", "confirm": true}
← Blocked.
```

**"Create a guest voucher for 24 hours"**
```
→ hotspot_create_voucher {"expire_minutes": 1440, "quota": 1}
← Preview: Would create 1 single-use voucher, 24h validity. Confirm?
→ hotspot_create_voucher {"expire_minutes": 1440, "quota": 1, "confirm": true}
← Voucher created: 83927-10458
```

**"Which APs are on old firmware?"**
```
→ device_list_basic
← 12 devices. Filtering by upgrade_available...
  - Lobby AP (U6-Pro) — current: 6.6.55, available: 7.0.83
  - Garage AP (U6-Lite) — current: 6.6.55, available: 7.0.83
```

---

## License

MIT &mdash; Not affiliated with Ubiquiti Inc.

---

<p align="center">
  <a href="https://www.buymeacoffee.com/oliverames">
    <img src="https://img.shields.io/badge/Buy_Me_a_Coffee-support-f5a542?style=for-the-badge&logo=buy-me-a-coffee&logoColor=white" alt="Buy Me a Coffee">
  </a>
</p>

<p align="center">
  <sub>
    Built by <a href="https://ames.consulting">Oliver Ames</a> in Vermont
    &bull; <a href="https://github.com/oliverames">GitHub</a>
    &bull; <a href="https://linkedin.com/in/oliverames">LinkedIn</a>
    &bull; <a href="https://bsky.app/profile/oliverames.bsky.social">Bluesky</a>
  </sub>
</p>
