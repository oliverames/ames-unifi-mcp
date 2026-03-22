# ames-unifi-mcp — Agent Guide

You are interacting with a UniFi Network controller through an MCP server. The controller is a Dream Machine running UniFi OS.

## How to use tools

This server uses **lazy mode** by default. You have 3 meta-tools:

1. **`tool_index`** — List all available tools. Filter by category: `devices`, `clients`, `networks`, `wlan`, `firewall`, `stats`, `events`, `system`, `hotspot`, `poe`, `dpi`, `backup`, `settings`, `routing`, `vpn`, `qos`.
2. **`tool_execute`** — Execute a tool by name with input parameters.
3. **`tool_batch`** — Execute multiple tools in parallel (e.g., get device list + client list simultaneously).

### Workflow

1. Call `tool_index` to see available tools (optionally filter by category)
2. Call `tool_execute` with the tool name and its parameters
3. For multi-step operations, use `tool_batch` to parallelize independent queries

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

If a tool isn't in the index, the controller may be too old.

## Common Multi-Step Operations

**Find and restart a specific device:**
1. `device_list_basic` to find the device by name/type
2. `device_restart` with the MAC address (will dry-run first)

**Check network health:**
1. Use `tool_batch` with `stats_site_health` + `client_list_active` + `alarm_count`

**Investigate a client issue:**
1. `client_get` with the MAC to check connection status
2. `stats_dpi_client` to see bandwidth usage by app
3. `client_reconnect` to force reconnect if needed

**Manage guest access:**
1. `hotspot_create_voucher` to generate access codes
2. `hotspot_authorize_guest` for direct MAC-based authorization
3. `hotspot_list_guests` to see active sessions

## MAC Address Format

Always use lowercase MACs (e.g., `aa:bb:cc:dd:ee:ff`). The controller expects this format.

## Undocumented Endpoints

Some tools use undocumented endpoints (marked in their descriptions). These work reliably on current firmware but behavior may change across versions. Examples: device locate (LED blink), per-client DPI stats, PoE power cycling.
