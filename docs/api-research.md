# UniFi API Research

> Living reference document for ames-unifi-mcp. Updated 2026-03-22.
> Every new tool's endpoint must be documented here before implementation.

## Sources

| Source | URL | Type | Coverage | Last Scraped |
|--------|-----|------|----------|-------------|
| Ubiquiti Developer Portal | https://developer.ui.com/ | Official | Cloud Site Manager API v1.0.0 (hosts, sites, devices, ISP metrics, SD-WAN) | 2026-03-22 |
| Getting Started Guide | https://help.ui.com/hc/en-us/articles/30076656117655 | Official | API key setup, local Integration API access | 2026-03-22 |
| beez.ly OpenAPI Specs | https://beez.ly/unifi-apis/ | Community (extracted from binaries) | Network Integration API (44 paths, 375 schemas, 18 versions 9.0.99–10.2.97), Protect API (25 paths, 158 schemas) | 2026-03-22 |
| ubiquiti-community/unifi-api | https://github.com/ubiquiti-community/unifi-api | Community OpenAPI | Legacy controller API (148 paths, 325 schemas, 239 operations at v9.1.120) | 2026-03-22 |
| UBNT Community Wiki | https://ubntwiki.com/products/software/unifi-controller/api | Community reverse-engineered | Most complete legacy API catalog including all cmd/ managers | 2026-03-22 |
| Art-of-WiFi UniFi-API-client | https://github.com/Art-of-WiFi/UniFi-API-browser | Community PHP | Deepest undocumented endpoint coverage (v1.1.101) | 2026-03-22 |
| claytono/go-unifi-mcp | https://github.com/claytono/go-unifi-mcp | MCP server (Go) | 242 generated operations, lazy/eager modes, ID resolution, query engine | 2026-03-22 |
| ubiquiti-community/go-unifi | https://github.com/ubiquiti-community/go-unifi | Go library | 69 resources with Go types, code-gen pipeline from controller field schemas | 2026-03-22 |
| sirkirby/unifi-mcp | https://github.com/sirkirby/unifi-mcp | MCP server (Python) | 144 tools (Network 88 + Protect 29 + Access 27), permissions, confirm gate | 2026-03-22 |
| gilberth/mcp-unifi-network | https://github.com/gilberth/mcp-unifi-network | MCP server (Python) | 64 tools, ZBF detection, original upstream of sirkirby | 2026-03-22 |

---

## API Architecture Overview

Ubiquiti exposes **three distinct API surfaces**:

| API | Base URL | Auth | Status | Use Case |
|-----|----------|------|--------|----------|
| **Cloud Site Manager** | `https://api.ui.com/v1/` | API Key (`X-API-Key`) | Official v1.0.0 GA | Multi-site management from cloud |
| **Local Integration API** | `https://{host}/integration/v1/` | API Key (`X-API-Key`) | Official (Network 9.0+) | Local controller, structured endpoints |
| **Legacy Controller API** | `https://{host}/api/s/{site}/` | Session cookie (login) | Unofficial, community-documented | Full feature access, all versions |

### Rate Limits

| Track | Limit |
|-------|-------|
| Cloud EA | 100 req/min |
| Cloud v1 GA | 10,000 req/min |
| Local controller | No documented limits (large queries can be slow, up to 5 min) |

---

## Authentication

### API Key (Preferred)

- **Header**: `X-API-Key: YOUR_KEY`
- **Stateless**: No login/logout needed
- **Scope**: Tied to the UI account that created it
- **Generation**: UniFi Network > Settings > Control Plane > Integrations
- **Key shown once** at creation — must be stored securely
- Available on Network 9.1.105+ (local) and via unifi.ui.com (cloud)

### Username/Password (Legacy)

- **Standard Controller**: `POST /api/login` with `{"username": "...", "password": "..."}`
- **UniFi OS (UDM/UCG/UDR)**: `POST /api/auth/login` with same body
- Returns session cookie for subsequent requests
- Logout: `POST /api/logout` (classic) or `POST /api/auth/logout` (UniFi OS)
- UniFi OS requires `X-CSRF-Token` header for mutating operations (extracted from login response)

### UniFi OS Detection

The PHP client detects UniFi OS by checking if `GET /` returns HTTP 200 (classic returns 302). sirkirby's implementation probes for `x-csrf-token` in response headers and tries `/proxy/network/api/self/sites` vs `/api/self/sites`.

---

## UniFi OS vs Standalone Controller Differences

| Aspect | Standalone Controller | UniFi OS (UDM Pro, UCG Max, etc.) |
|--------|----------------------|-----------------------------------|
| Login endpoint | `POST /api/login` | `POST /api/auth/login` |
| API path prefix | None | `/proxy/network` before all `/api/` endpoints |
| User self endpoint | `GET /api/self` | `GET /api/users/self` |
| CSRF token | Not required | Required for `system/reboot`, `system/poweroff`, `stat/authorization` |
| Device by MAC | Not available | `GET /stat/device/{mac}` available |
| Port | `:8443` | `:443` (standard HTTPS), `:11443` for UniFi OS Server |
| Multi-application | Network only | Hosts Network, Protect, Access, Talk on same console |

---

## Official Cloud Site Manager API (`api.ui.com/v1/`)

All endpoints require `X-API-Key` header. Currently **read-only**.

### Hosts

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/hosts` | List all hosts (supports pagination: `pageSize`, `nextToken`) |
| GET | `/v1/hosts/{id}` | Get host by ID |

Response fields: `id`, `hardwareId`, `type`, `ipAddress`, `owner`, `isBlocked`, `registrationTime`, `lastConnectionStateChange`, `latestBackupTime`, `userData`, `reportedState`

### Sites

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/sites` | List all sites from Network hosts (paginated) |

Response fields: `siteId`, `hostId`, `meta` (desc, gatewayMac, name, timezone), `statistics` (device/client counts, performance), `permission`, `isOwner`

### Devices

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/devices` | List devices grouped by host (filter: `hostIds[]`, `time`) |

Response: grouped by host. Device fields: `id`, `mac`, `name`, `model`, `shortname`, `ip`, `productLine`, `status`, `version`, `firmwareStatus`, `updateAvailable`, `isConsole`, `isManaged`, `startupTime`, `adoptionTime`, `note`, `uidb`

### ISP Metrics

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/isp-metrics/{type}` | Get ISP metrics (`5m` or `1h`). Query: `beginTimestamp`, `endTimestamp`, or `duration` (24h/7d/30d) |
| POST | `/v1/isp-metrics/{type}/query` | Query ISP metrics for specific sites with per-site time ranges |

Metrics: `avgLatency`, `download_kbps`, `upload_kbps`, `downtime`, `uptime`, `ispAsn`, `ispName`, `maxLatency`, `packetLoss`

### SD-WAN

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/sd-wan-configs` | List all SD-WAN configurations |
| GET | `/v1/sd-wan-configs/{id}` | Get config details (variant, settings, hubs, spokes) |
| GET | `/v1/sd-wan-configs/{id}/status` | Get deployment status and tunnel health |

---

## Local Integration API (`/integration/v1/` or `/proxy/network/integration/v1/`)

Official API available locally on controllers running Network 9.0+. Auth via API key. Has grown from 7 paths (v9.0.99) to 44 paths (v10.2.97).

### Application Info

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/info` | Get application info (version, runtime) |

### Sites

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/sites` | List local sites |

### Devices

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/pending-devices` | List devices pending adoption |
| GET | `/v1/sites/{siteId}/devices` | List adopted devices |
| POST | `/v1/sites/{siteId}/devices` | Adopt devices |
| GET | `/v1/sites/{siteId}/devices/{deviceId}` | Get device details |
| DELETE | `/v1/sites/{siteId}/devices/{deviceId}` | Remove (unadopt) device |
| POST | `/v1/sites/{siteId}/devices/{deviceId}/actions` | Execute device action |
| POST | `/v1/sites/{siteId}/devices/{deviceId}/interfaces/ports/{portIdx}/actions` | Execute port action |
| GET | `/v1/sites/{siteId}/devices/{deviceId}/statistics/latest` | Get latest device statistics |

### Clients

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/sites/{siteId}/clients` | List connected clients |
| GET | `/v1/sites/{siteId}/clients/{clientId}` | Get client details |
| POST | `/v1/sites/{siteId}/clients/{clientId}/actions` | Execute client action |

### Networks (Full CRUD)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/sites/{siteId}/networks` | List networks |
| POST | `/v1/sites/{siteId}/networks` | Create network |
| GET | `/v1/sites/{siteId}/networks/{networkId}` | Get network details |
| PUT | `/v1/sites/{siteId}/networks/{networkId}` | Update network |
| DELETE | `/v1/sites/{siteId}/networks/{networkId}` | Delete network |
| GET | `/v1/sites/{siteId}/networks/{networkId}/references` | Get network references |

### WiFi Broadcasts (Full CRUD)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/sites/{siteId}/wifi/broadcasts` | List WiFi broadcasts |
| POST | `/v1/sites/{siteId}/wifi/broadcasts` | Create WiFi broadcast |
| GET | `/v1/sites/{siteId}/wifi/broadcasts/{id}` | Get broadcast details |
| PUT | `/v1/sites/{siteId}/wifi/broadcasts/{id}` | Update broadcast |
| DELETE | `/v1/sites/{siteId}/wifi/broadcasts/{id}` | Delete broadcast |

### Hotspot Vouchers

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/sites/{siteId}/hotspot/vouchers` | List vouchers |
| POST | `/v1/sites/{siteId}/hotspot/vouchers` | Generate vouchers |
| DELETE | `/v1/sites/{siteId}/hotspot/vouchers` | Delete vouchers (bulk) |
| GET | `/v1/sites/{siteId}/hotspot/vouchers/{id}` | Get voucher details |
| DELETE | `/v1/sites/{siteId}/hotspot/vouchers/{id}` | Delete voucher |

### Firewall (Zone-Based, Network 9.0+)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/sites/{siteId}/firewall/policies` | List firewall policies |
| POST | `/v1/sites/{siteId}/firewall/policies` | Create firewall policy |
| GET | `/v1/sites/{siteId}/firewall/policies/ordering` | Get policy ordering |
| PUT | `/v1/sites/{siteId}/firewall/policies/ordering` | Reorder policies |
| GET | `/v1/sites/{siteId}/firewall/policies/{id}` | Get policy |
| PUT | `/v1/sites/{siteId}/firewall/policies/{id}` | Update policy |
| DELETE | `/v1/sites/{siteId}/firewall/policies/{id}` | Delete policy |
| PATCH | `/v1/sites/{siteId}/firewall/policies/{id}` | Patch policy |
| GET | `/v1/sites/{siteId}/firewall/zones` | List zones |
| POST | `/v1/sites/{siteId}/firewall/zones` | Create custom zone |
| GET | `/v1/sites/{siteId}/firewall/zones/{id}` | Get zone |
| PUT | `/v1/sites/{siteId}/firewall/zones/{id}` | Update zone |
| DELETE | `/v1/sites/{siteId}/firewall/zones/{id}` | Delete custom zone |

### ACL Rules (Full CRUD + Ordering)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/sites/{siteId}/acl-rules` | List ACL rules |
| POST | `/v1/sites/{siteId}/acl-rules` | Create ACL rule |
| GET | `/v1/sites/{siteId}/acl-rules/ordering` | Get ACL ordering |
| PUT | `/v1/sites/{siteId}/acl-rules/ordering` | Reorder ACL rules |
| GET | `/v1/sites/{siteId}/acl-rules/{id}` | Get ACL rule |
| PUT | `/v1/sites/{siteId}/acl-rules/{id}` | Update ACL rule |
| DELETE | `/v1/sites/{siteId}/acl-rules/{id}` | Delete ACL rule |

### DNS Policies (Full CRUD)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/sites/{siteId}/dns/policies` | List DNS policies |
| POST | `/v1/sites/{siteId}/dns/policies` | Create DNS policy |
| GET | `/v1/sites/{siteId}/dns/policies/{id}` | Get DNS policy |
| PUT | `/v1/sites/{siteId}/dns/policies/{id}` | Update DNS policy |
| DELETE | `/v1/sites/{siteId}/dns/policies/{id}` | Delete DNS policy |

### Switching (Read-Only)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/sites/{siteId}/switching/switch-stacks` | List switch stacks |
| GET | `/v1/sites/{siteId}/switching/switch-stacks/{id}` | Get switch stack |
| GET | `/v1/sites/{siteId}/switching/mc-lag-domains` | List MC-LAG domains |
| GET | `/v1/sites/{siteId}/switching/mc-lag-domains/{id}` | Get MC-LAG domain |
| GET | `/v1/sites/{siteId}/switching/lags` | List LAGs |
| GET | `/v1/sites/{siteId}/switching/lags/{id}` | Get LAG details |

### Traffic Matching Lists (Full CRUD)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/sites/{siteId}/traffic-matching-lists` | List traffic matching lists |
| POST | `/v1/sites/{siteId}/traffic-matching-lists` | Create traffic matching list |
| GET | `/v1/sites/{siteId}/traffic-matching-lists/{id}` | Get traffic matching list |
| PUT | `/v1/sites/{siteId}/traffic-matching-lists/{id}` | Update traffic matching list |
| DELETE | `/v1/sites/{siteId}/traffic-matching-lists/{id}` | Delete traffic matching list |

### Supporting Resources (Read-Only)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/countries` | List countries |
| GET | `/v1/dpi/applications` | List DPI applications |
| GET | `/v1/dpi/categories` | List DPI categories |
| GET | `/v1/sites/{siteId}/device-tags` | List device tags |
| GET | `/v1/sites/{siteId}/radius/profiles` | List RADIUS profiles |
| GET | `/v1/sites/{siteId}/vpn/servers` | List VPN servers |
| GET | `/v1/sites/{siteId}/vpn/site-to-site-tunnels` | List site-to-site VPN tunnels |
| GET | `/v1/sites/{siteId}/wans` | List WAN interfaces |

### Filtering System

The Integration API supports advanced filtering via `filter` query parameter:
- Property filters: `filter=name eq "MyNetwork"`
- Compound expressions: AND/OR/NOT
- Pattern matching: contains, startsWith, endsWith
- Negation: `not(name eq "default")`

---

## Legacy Controller API (`/api/s/{site}/...`)

The traditional API available on all controller versions. Uses session-based auth. Prefixed with `/proxy/network` on UniFi OS.

### Controller-Level Endpoints (No Site Context)

| Method | Path | Description | Auth Required |
|--------|------|-------------|---------------|
| GET | `/status` | Server info (version, UUID, uptime) | No |
| GET | `/api/self` | Current user info (classic) | Yes |
| GET | `/api/users/self` | Current user info (UniFi OS) | Yes |
| GET | `/api/self/sites` | All sites for current user | Yes |
| GET | `/api/stat/sites` | Sites with health and alert info | Yes |
| GET | `/api/stat/admin` | All admins and permissions | Yes |
| POST | `/api/system/reboot` | Reboot UDM (X-CSRF-Token required) | Yes |
| POST | `/api/system/poweroff` | Power off UDM (X-CSRF-Token required) | Yes |
| GET | `/dl/firmware/bundles.json` | Device firmware mappings | No |
| GET | `/dl/autobackup/{filename}` | Download auto-backup file | Yes |

### stat/ Endpoints (Read-Only Statistics)

| Path | Method | Description | Parameters |
|------|--------|-------------|------------|
| `stat/health` | GET | Site health status | — |
| `stat/sysinfo` | GET | Controller info (version, uptime) | — |
| `stat/sta` | GET | All active/online clients | — |
| `stat/sta/{mac}` | GET | Single active client by MAC | — |
| `stat/user/{mac}` | GET | Single configured client details | — |
| `stat/alluser` | POST | All configured clients | `{"type":"all","conn":"all","within":8760}` |
| `stat/guest` | POST | Guest devices | `{"within": hours}` |
| `stat/device-basic` | GET | Devices with minimal keys | — |
| `stat/device` | GET/POST | Full device details (POST filters by MACs on classic) | `{"macs":["mac1",...]}` |
| `stat/device/{mac}` | GET | Single device by MAC (UDM only) | — |
| `stat/event` | GET/POST | Events (newest first, 3000 limit) | `{"_sort":"-time","within":720,"_start":0,"_limit":3000}` |
| `stat/alarm` | GET | Alarms (newest first, 3000 limit) | — |
| `stat/ccode` | GET | Country codes (ISO 3166-1) | — |
| `stat/current-channel` | GET | Available RF channels | — |
| `stat/rogueap` | GET/POST | Neighboring/rogue APs | `{"within": hours}` |
| `stat/sitedpi` | GET/POST | Site DPI stats | `{"type":"by_app"}` or `{"type":"by_cat"}` |
| `stat/stadpi` | GET/POST | Per-client DPI stats | `{"type":"by_app","macs":[...]}` |
| `stat/dpi` | GET | DPI stats | — |
| `stat/dynamicdns` | GET | Dynamic DNS status | — |
| `stat/spectrumscan` | GET | RF scan results | — |
| `stat/spectrumscan/{mac}` | GET | RF scan results for specific AP | — |
| `stat/routing` | GET | Active routes | — |
| `stat/portforward` | GET | Port forward stats | — |
| `stat/voucher` | POST | Hotspot vouchers | `{"create_time": unix_ts}` |
| `stat/payment` | GET | Hotspot payments | `?within=hours` |
| `stat/authorization` | POST | Auth codes used in timeframe | `{"start":"ts","end":"ts"}` (CSRF on UDM) |
| `stat/session` | POST | Login sessions | `{"type":"all","start":ts,"end":ts,"mac":"..."}` |
| `stat/stream` | GET | EDU streams (Insights) | — |
| `stat/sdn` | GET | Cloud/SSO connection status | — |
| `stat/dashboard` | GET | Dashboard metrics (v4.9.1+) | `?scale=5minutes` |
| `stat/fwupdate/latest-version` | GET | Check controller update | — |
| `stat/ips/event` | POST | IPS/IDS events (v5.9+) | `{"start":ms,"end":ms,"_limit":10000}` |

### stat/report/ Endpoints (Time-Series Reports)

All use `POST` with `{"attrs":[...],"start":ms_timestamp,"end":ms_timestamp}`.

| Path Pattern | Intervals | Types | Description |
|-------------|-----------|-------|-------------|
| `stat/report/{interval}.site` | 5minutes, hourly, daily, monthly | site | Site bandwidth, client counts |
| `stat/report/{interval}.ap` | 5minutes, hourly, daily, monthly | ap | AP-level stats (tx attempts, retries, dropped) |
| `stat/report/{interval}.user` | 5minutes, hourly, daily, monthly | user | Client stats (v5.8+: signal, rates, satisfaction) |
| `stat/report/{interval}.gw` | 5minutes, hourly, daily, monthly | gw | Gateway stats (mem, cpu, loadavg, LAN counters) |
| `stat/report/archive.speedtest` | — | speedtest | Speed test history |

**Common attrs by type:**
- **site**: `bytes`, `wan-tx_bytes`, `wan-rx_bytes`, `wlan_bytes`, `num_sta`, `lan-num_sta`, `wlan-num_sta`, `time`
- **ap**: `bytes`, `num_sta`, `time`, `wifi_tx_attempts`, `tx_retries`, `wifi_tx_dropped`, `mac_filter_rejections`
- **user**: `rx_bytes`, `tx_bytes`, `time`, `signal`, `rx_rate`, `tx_rate`, `satisfaction`, `duration`
- **gw**: `mem`, `cpu`, `loadavg_5`, `time`, `lan-rx_bytes`, `lan-tx_bytes`, `lan-rx_packets`, `lan-tx_packets`

### rest/ Endpoints (CRUD Resources)

All support `GET` for listing. `POST` creates, `PUT /{_id}` updates, `DELETE /{_id}` removes.

| Path | Description |
|------|-------------|
| `rest/user` | Configured/known clients |
| `rest/device/{_id}` | Device settings |
| `rest/setting` | All site settings |
| `rest/setting/{key}/{_id}` | Specific setting section |
| `rest/firewallrule` | Firewall rules |
| `rest/firewallgroup` | Firewall groups |
| `rest/wlanconf` | WLAN configurations |
| `rest/routing` | Static routes |
| `rest/tag` | Device tags (v5.5+) |
| `rest/networkconf` | Network configurations |
| `rest/portconf` | Switch port profiles |
| `rest/portforward` | Port forwarding rules |
| `rest/radiusprofile` | RADIUS profiles (v5.5.19+) |
| `rest/account` | RADIUS accounts (v5.5.19+) |
| `rest/dynamicdns` | Dynamic DNS config |
| `rest/usergroup` | User groups (bandwidth settings) |
| `rest/hotspotop` | Hotspot operators |
| `rest/hotspotpackage` | Hotspot packages |
| `rest/rogueknown` | Known rogue APs |
| `rest/alarm` | Alarms (`?archived=false` for active only) |
| `rest/event` | Events (oldest first) |
| `rest/wlangroup` | WLAN groups |
| `rest/broadcastgroup` | Broadcast groups |
| `rest/channelplan` | WiFi channel plans |
| `rest/dashboard` | Custom dashboards |
| `rest/dhcpoption` | Custom DHCP options |
| `rest/dpiapp` | DPI app definitions |
| `rest/dpigroup` | DPI app groups |
| `rest/heatmap` | RF heatmaps |
| `rest/heatmappoint` | RF heatmap points |
| `rest/hotspot2conf` | Hotspot 2.0/Passpoint |
| `rest/map` | Site maps |
| `rest/mediafile` | Media files |
| `rest/scheduletask` | Scheduled tasks |
| `rest/spatialrecord` | Spatial/location records |
| `rest/virtualdevice` | Virtual devices |

### Settings Resources (42 categories)

Read via `GET /get/setting` (all) or `GET /get/setting/{name}`. Update via `PUT /set/setting/{name}/{_id}`.

| Setting Key | Description |
|-------------|-------------|
| `auto_speedtest` | Automatic speed test config |
| `baresip` | SIP/VoIP settings |
| `broadcast` | Broadcast/multicast |
| `connectivity` | Connectivity monitoring |
| `country` | Country/regulatory |
| `dashboard` | Dashboard widgets |
| `doh` | DNS-over-HTTPS |
| `dpi` | Deep Packet Inspection toggle |
| `element_adopt` | Element adoption |
| `ether_lighting` | Ethernet port LED lighting |
| `evaluation_score` | Evaluation scores |
| `global_ap` | Global AP settings |
| `global_nat` | Global NAT config |
| `global_switch` | Global switch (ACL L3 isolation) |
| `guest_access` | Guest portal/access |
| `ips` | IPS/IDS (alerts, suppression, DNS filters, honeypot, ad blocking) |
| `lcm` | LED Controller Module |
| `locale` | Locale/timezone |
| `magic_site_to_site_vpn` | Magic site-to-site VPN |
| `mdns` | mDNS settings |
| `mgmt` | Management (SSH keys) |
| `netflow` | NetFlow export |
| `network_optimization` | Network optimization |
| `ntp` | NTP server |
| `porta` | Portal/hotspot |
| `radio_ai` | Radio AI (channel blacklists) |
| `radius` | RADIUS server |
| `roaming_assistant` | Roaming assistant |
| `rsyslogd` | Remote syslog |
| `snmp` | SNMP configuration |
| `ssl_inspection` | SSL inspection |
| `super_cloudaccess` | Cloud access |
| `super_events` | Event notifications |
| `super_fwupdate` | Firmware updates |
| `super_identity` | Identity settings |
| `super_mail` | Mail notifications |
| `super_mgmt` | Super management |
| `super_sdn` | SDN settings |
| `super_smtp` | SMTP configuration |
| `teleport` | Teleport VPN |
| `traffic_flow` | Traffic flow |
| `usg` | USG/gateway (DNS verification) |
| `usw` | USW/switch settings |

### list/ Endpoints (Read-Only)

| Path | Description |
|------|-------------|
| `list/user` | All known clients |
| `list/usergroup` | User groups |
| `list/wlangroup` | WLAN groups |
| `list/portforward` | Port forwarding |
| `list/portconf` | Port configurations |
| `list/extension` | VoIP extensions |
| `list/alarm` | Alarms with optional filter: `{"archived":false,"key":"EVT_GW_WANTransition"}` |

### cnt/ Endpoints (Counts)

| Path | Description |
|------|-------------|
| `cnt/alarm` | Count all alarms |
| `cnt/alarm?archived=false` | Count active alarms |
| `cnt/{resource}` | Count for any /rest/ resource |

### set/ Endpoints (Apply Settings)

| Path | Description |
|------|-------------|
| `set/setting/mgmt` | Toggle site LEDs |
| `set/setting/guest_access` | Guest login settings |
| `set/setting/ips` | IPS/IDS settings |
| `set/setting/super_mgmt/{_id}` | Super management |
| `set/setting/super_smtp/{_id}` | SMTP settings |
| `set/setting/super_identity/{_id}` | Identity settings |
| `set/setting/element_adopt` | Element adoption |

### upd/ Endpoints

| Path | Description |
|------|-------------|
| `upd/user/{_id}` | Update client name, note, user group |
| `upd/device/{_id}` | Update device name, radio, WLAN group |

### v2/ API Endpoints

| Path | Method | Description |
|------|--------|-------------|
| `v2/api/site/{site}/apgroups` | GET/POST | AP groups (v6.0+) |
| `v2/api/site/{site}/apgroups/{_id}` | PUT/DELETE | AP group management |
| `v2/api/site/{site}/trafficrules` | GET/POST | Traffic rules |
| `v2/api/site/{site}/trafficrules/{_id}` | PUT/DELETE | Traffic rule management (PUT returns 201) |
| `v2/api/site/{site}/system-log/{class}` | POST | System log by class |
| `v2/api/fingerprint_devices/{source}` | GET | Client device fingerprints |

System log classes: `device-alert`, `next-ai-alert`, `vpn-alert`, `admin-activity`, `update-alert`, `client-alert`, `threat-alert`, `triggers`

---

## cmd/ Manager Endpoints

All use `POST /api/s/{site}/cmd/{manager}` with body `{"cmd": "command_name", ...params}`.

### stamgr (Station/Client Manager)

| Command | Parameters | Description |
|---------|-----------|-------------|
| `authorize-guest` | mac, minutes, up?, down?, bytes?, ap_mac? | Authorize guest with limits |
| `unauthorize-guest` | mac | Revoke guest authorization |
| `kick-sta` | mac | Disconnect/reconnect client |
| `block-sta` | mac | Block client |
| `unblock-sta` | mac | Unblock client |
| `forget-sta` | macs[] | Forget client(s) (v5.9+, can be slow) |

### devmgr (Device Manager)

| Command | Parameters | Description |
|---------|-----------|-------------|
| `adopt` | macs[] | Adopt device(s) |
| `adv-adopt` | mac, ip, username, password, url, port?, sshKeyVerify? | Adopt via custom SSH credentials |
| `restart` | macs[], reboot_type?("soft"/"hard") | Reboot device(s). "hard" power-cycles PoE. |
| `force-provision` | macs[] | Force provision device(s) |
| `power-cycle` | mac, port_idx | **Power-cycle PoE switch port** |
| `speedtest` | — | Start speed test |
| `speedtest-status` | — | Get speed test state |
| `set-locate` | mac | **Blink LED to locate device** |
| `unset-locate` | mac | Stop blinking LED |
| `upgrade` | mac | Upgrade to latest stable firmware |
| `upgrade-external` | mac/macs[], url | Upgrade to specific firmware URL |
| `migrate` | macs[], inform_url | Push new inform URL |
| `cancel-migrate` | macs[] | Cancel migration |
| `spectrum-scan` | mac | Trigger RF scan on AP |
| `set-rollupgrade` | [device_types] | Start rolling upgrade (default: uap, usw, ugw, uxg) |
| `unset-rollupgrade` | — | Cancel rolling upgrade |

### sitemgr (Site Manager)

| Command | Parameters | Description |
|---------|-----------|-------------|
| `add-site` | desc, name? | Create new site |
| `delete-site` | site (_id) | Delete site |
| `update-site` | desc | Rename current site |
| `get-admins` | — | List admins for site |
| `move-device` | mac, site (_id) | Move device to another site |
| `delete-device` | mac | Remove device from site |
| `invite-admin` | name, email, for_sso?, role?, permissions[] | Invite admin |
| `grant-admin` | admin (_id), role?, permissions[] | Assign admin to site |
| `update-admin` | admin (_id), name, email, x_password?, role?, permissions[] | Update admin |
| `revoke-admin` | admin (_id) | Revoke admin from site |
| `revoke-super-admin` | admin (_id) | Delete admin entirely |

### evtmgr (Event Manager)

| Command | Parameters | Description |
|---------|-----------|-------------|
| `archive-all-alarms` | — | Archive all alarms |
| `archive-alarm` | _id | Archive single alarm |

### hotspot (Hotspot Manager)

| Command | Parameters | Description |
|---------|-----------|-------------|
| `create-voucher` | expire (min), n (count), quota (0=multi,1=single), note?, up?, down?, bytes? | Create vouchers |
| `delete-voucher` | _id | Revoke voucher |
| `extend` | _id | Extend guest authorization |

### backup (Backup Manager)

| Command | Parameters | Description |
|---------|-----------|-------------|
| `backup` | days? | Generate backup |
| `list-backups` | — | List auto-backup files |
| `delete-backup` | filename | Delete backup file |
| `export-site` | — | Export current site |

### system

| Command | Parameters | Description |
|---------|-----------|-------------|
| `backup` | — | Create backup (alias) |
| `reboot` | — | Reboot CloudKey/UniFi OS |

### stat

| Command | Parameters | Description |
|---------|-----------|-------------|
| `clear-dpi` / `reset-dpi` | — | Reset site DPI counters |

### firmware

| Command | Parameters | Description |
|---------|-----------|-------------|
| `list-available` | — | List available firmware |
| `list-cached` | — | List cached firmware |

### productinfo

| Command | Parameters | Description |
|---------|-----------|-------------|
| `check-firmware-update` | — | Trigger firmware update check |

### Additional Managers (Low Confidence)

| Manager | Confidence | Notes |
|---------|------------|-------|
| `cfgmgr` | Medium | Configuration manager — no documented commands |
| `streammgr` | Low | Stream manager |
| `throughput` | Low | Throughput testing |
| `firewall` | Low | Firewall management |
| `elite` | Low | Elite device management |

---

## Guest/Hotspot Endpoints

| Path | Method | Description |
|------|--------|-------------|
| `guest/s/{site}/hotspotconfig` | GET | Hotspot configuration (auth type, portal design) |
| `guest/s/{site}/hotspotpackages` | GET | Hotspot packages |

---

## WebSocket Event Stream

| Path | Description | Confidence |
|------|-------------|------------|
| `wss://{host}:8443/wss/s/{site}/events` | Classic controller real-time events | High |
| `wss://{host}/proxy/network/wss/s/{site}/events` | UniFi OS real-time events | High |
| `wss://{host}/api/ws/system` | UniFi OS system-level events | Medium |

Events are JSON-formatted. Auth uses same session cookie. Includes device state changes, client connects/disconnects, alerts, speed test results.

---

## Undocumented Endpoints

| Endpoint | Method | Payload | Source | Confidence | Version Range |
|----------|--------|---------|--------|------------|---------------|
| `cmd/devmgr` (power-cycle) | POST | `{"cmd":"power-cycle","mac":"...","port_idx":N}` | Wiki, PHP | Confirmed | All |
| `cmd/devmgr` (set-locate) | POST | `{"cmd":"set-locate","mac":"..."}` | Wiki, PHP | Confirmed | All |
| `cmd/devmgr` (unset-locate) | POST | `{"cmd":"unset-locate","mac":"..."}` | Wiki, PHP | Confirmed | All |
| `stat/stadpi` | POST | `{"type":"by_app","macs":["..."]}` | Wiki | Confirmed | v5.8+ |
| `stat/sitedpi` | POST | `{"type":"by_app"}` or `{"type":"by_cat"}` | Wiki, PHP | Confirmed | v5.8+ |
| `stat/session` | POST | `{"type":"all","start":ts,"end":ts}` | Wiki, PHP | Confirmed | All |
| `stat/authorization` | POST | `{"start":"ts","end":"ts"}` | Wiki | Confirmed | All |
| `stat/stream` | GET | — | Wiki | Medium | Unknown |
| `stat/sdn` | GET | — | Wiki | Medium | Unknown |
| `cmd/cfgmgr` | POST | Unknown | Wiki | Medium | Unknown |
| `cmd/streammgr` | POST | Unknown | Wiki | Low | Unknown |
| `cmd/throughput` | POST | Unknown | Wiki | Low | Unknown |
| `cmd/firewall` | POST | Unknown | Wiki | Low | Unknown |
| `cmd/elite` | POST | Unknown | Wiki | Low | Unknown |
| `wss://{host}/api/ws/system` | WebSocket | — | Community | Medium | UniFi OS |
| `rest/rogueknown` | GET | — | PHP | Confirmed | All |
| `cnt/{resource}` | GET | — | Wiki | Medium | All |
| `dl/firmware/bundles.json` | GET | — | PHP | Confirmed | All |
| `v2/api/site/{site}/system-log/{class}` | POST | — | PHP | Confirmed | v6.0+ |
| `v2/api/fingerprint_devices/{source}` | GET | — | PHP | Confirmed | Unknown |

---

## Version Compatibility Matrix

| Feature | Minimum Version | Notes |
|---------|----------------|-------|
| Legacy API (all endpoints) | All versions | Community-documented |
| API key auth (local) | 9.1.105+ | Settings > Control Plane > Integrations |
| Integration API (local) | 9.0+ | `/integration/v1/` path |
| Zone-Based Firewall | 9.0+ | Replaces legacy firewall rules |
| Traffic Rules v2 | 6.0+ | `v2/api/site/{site}/trafficrules` |
| AP Groups v2 | 6.0+ | `v2/api/site/{site}/apgroups` |
| System Log v2 | 6.0+ | `v2/api/site/{site}/system-log/{class}` |
| DPI stats | 5.8+ | `stat/sitedpi`, `stat/stadpi` |
| IPS/IDS events | 5.9+ | `stat/ips/event` |
| Forget client | 5.9+ | `cmd/stamgr` forget-sta |
| Device tags | 5.5+ | `rest/tag` |
| RADIUS profiles | 5.5.19+ | `rest/radiusprofile`, `rest/account` |
| Dashboard stats | 4.9.1+ | `stat/dashboard` |
| DNS Policies | 10.0+ | `/v1/sites/{siteId}/dns/policies` |
| ACL Rules | 10.0+ | `/v1/sites/{siteId}/acl-rules` |
| Traffic Matching Lists | 10.0+ | `/v1/sites/{siteId}/traffic-matching-lists` |
| Switching (stacks, LAGs) | 10.0+ | `/v1/sites/{siteId}/switching/*` |

---

## Existing MCP Server Comparison

| Feature | claytono/go-unifi-mcp | sirkirby/unifi-mcp | gilberth/mcp-unifi-network |
|---------|----------------------|-------------------|---------------------------|
| Language | Go | Python | Python |
| Total tools | 242 (generated) | 144 (88+29+27) | 64 |
| Products | Network | Network + Protect + Access | Network |
| Tool gen | Code-gen from schemas | Hand-written | Hand-written |
| Lazy loading | 3 meta-tools | 4-5 meta-tools (3 modes) | None (eager only) |
| API key auth | Yes | Yes (dual) | No |
| Confirm gate | None | Yes | Yes |
| Permissions | None | 4-level cascade | Simple YAML |
| ID resolution | Automatic (_id -> _name) | No | No |
| Query/filter | filter, search, fields | No | No |
| Batch | Parallel goroutines | Async | No |
| V2/ZBF | FirewallZone, ZonePolicy | Policies, zones, IP groups | Policies, zones, IP groups |
| Context cost (lazy) | ~200 tokens | ~200 tokens | ~55K+ |

### Resources in go-unifi NOT Exposed by Any MCP Server

- BGPConfig, OSPFRouter (advanced routing)
- DPIApp, DPIGroup (deep packet inspection definitions)
- NAT rules (direct management)
- ClientGroup, ClientInfo
- NetworkMembersGroup
- Most of the 34 settings resources (only go-unifi-mcp covers these)
- Device tags, schedule tasks, spatial records, maps, heat maps, broadcast groups

---

## Known Gaps

### Not Documented Anywhere
- Full WebSocket event message schema and event types
- Complete list of device action payloads for Integration API
- Client action payloads for Integration API
- Port action payloads for Integration API
- UniFi Access API (only sirkirby has reverse-engineered tools)
- UniFi Talk API
- UniFi Connect API
- Per-client bandwidth override via REST (only via `cmd/stamgr` or `rest/user` with `qos_policy_applied`)

### Implementation Notes

1. **PUT vs POST**: Updating existing objects requires `PUT` with `_id` appended — `POST` creates new.
2. **Timestamps**: `stat/report/` uses milliseconds. Session endpoints use seconds.
3. **MAC addresses**: Always lowercase.
4. **DPI compound IDs**: Category-app compound ID uses `(cat << 16) + app`.
5. **Port differences**: Classic = 8443, UniFi OS gateway = 443, UniFi OS Server = 11443.
6. **Response envelope**: Classic API wraps in `{"meta":{"rc":"ok"},"data":[...]}`. Integration API uses `{"data":..., "httpStatusCode":200}`.

---

## Response Formats

### Legacy API
```json
{
  "meta": { "rc": "ok" },
  "data": [ ... ]
}
```
Error: `"meta": { "rc": "error", "msg": "api.err.LoginRequired" }`

### Integration API
```json
{
  "data": [...],
  "httpStatusCode": 200,
  "traceId": "unique-identifier"
}
```
Paginated: adds `"nextToken": "..."`. Error: `{"code": "ERROR_CODE", "httpStatusCode": 404, "message": "..."}`

### Cloud API
Same as Integration API format with pagination via `pageSize` + `nextToken` query params.
