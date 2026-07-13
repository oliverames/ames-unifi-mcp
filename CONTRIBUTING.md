# Contributing

Thanks for helping improve the UniFi MCP server.

## Before you start

- Use Go 1.26.1 or newer and Node.js 22 or newer.
- Open an issue before making a large API or architecture change.
- Read `docs/api-research.md` before adding a tool. Update that file when an endpoint, version requirement, or controller behavior changes.
- Keep mutating tools behind the existing confirmation gate.

## Local checks

Run the same checks used in CI:

```bash
test -z "$(gofmt -l .)"
go test -race ./...
go vet ./...
go run honnef.co/go/tools/cmd/staticcheck@2026.1 ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.6.0 ./...
node --check install.js scripts/build-mcpb.mjs mcpb/server/launch.js
make build-all
```

Live controller checks are opt-in because they require real network access. Follow [the integration testing guide](docs/INTEGRATION_TESTING.md), begin with a dedicated test site and the `read-only` permission profile, and never commit controller credentials or captured responses.

## Pull requests

Keep each pull request focused. Describe the controller version you tested, list any new or changed tools, and call out mutations or permission changes. Include tests for behavior that can run without a controller.
