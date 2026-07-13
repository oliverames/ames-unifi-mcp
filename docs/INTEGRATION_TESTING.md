# Integration Testing

Unit tests and CI do not connect to a controller. Live checks are a separate, deliberate step because they can expose network data or change controller state.

## Safe starting point

1. Use a dedicated test site or lab controller.
2. Create a least-privilege API key.
3. Set `UNIFI_PERMISSION_PROFILE=read-only` and `UNIFI_TOOL_MODE=lazy`.
4. Keep `UNIFI_VERIFY_SSL=true` unless the lab controller uses a self-signed certificate you have independently verified.
5. Do not record raw responses from a real controller in fixtures, issues, or CI logs.

Configure credentials directly:

```bash
export UNIFI_HOST="https://controller.example"
export UNIFI_API_KEY="replace-me"
export UNIFI_SITE="default"
export UNIFI_PERMISSION_PROFILE="read-only"
export UNIFI_TOOL_MODE="lazy"
```

Or point the process at your own 1Password fields:

```bash
export UNIFI_HOST_OP_REF="op://Your Vault/Your Item/host"
export UNIFI_API_KEY_OP_REF="op://Your Vault/Your Item/api_key"
export UNIFI_PERMISSION_PROFILE="read-only"
```

## Smoke sequence

Run the server through an MCP client and verify these in order:

1. With credentials removed, the server starts and each tool returns the configuration hint.
2. `tool_index` returns the documented catalog without exposing credentials.
3. A simple read such as `device_list` or `stats_site_health` succeeds against the test site.
4. A mutating tool is denied under the `read-only` profile.
5. Under `standard` on a test site, a mutating tool returns a preview when `confirm` is absent or false.
6. Only on a disposable test object, repeat the operation with `confirm: true` and verify the controller state changed once.

Record the server version, UniFi OS version, Network application version, authentication method, and permission profile with the result. Keep hostnames, site IDs, MAC addresses, IP addresses, usernames, and API responses out of the report.
