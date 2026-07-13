# Security Policy

## Supported versions

Security fixes are made on the current release line. Upgrade to the latest npm or GitHub release before reporting a problem that may already be fixed.

## Report a vulnerability

Use [GitHub private vulnerability reporting](https://github.com/oliverames/ames-unifi-mcp/security/advisories/new). Do not open a public issue for credential exposure, authentication bypasses, confirmation-gate failures, request forgery, or controller access flaws.

Include the affected version, controller and Network application versions, permission profile, reproduction steps, and expected impact. Remove API keys, passwords, cookies, hostnames, device identifiers, and response bodies that identify a real network.

## Credential handling

The server reads credentials from environment variables or caller-supplied 1Password references. It does not ship with a vault name, item name, credential, or telemetry service. Start with `UNIFI_PERMISSION_PROFILE=read-only`, keep TLS verification enabled when the controller has a trusted certificate, and use a dedicated UniFi API key with the least access the workflow needs.
