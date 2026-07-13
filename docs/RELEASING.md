# Releasing

The tag, npm package, MCPB bundle, native binaries, checksums, and SBOM must describe the same version.

## Prepare

1. Update `package.json` and the release links in `README.md`.
2. Run the full local verification suite from `CONTRIBUTING.md`.
3. Build the bundle with `make clean && make mcpb`.
4. Inspect the npm payload with `npm pack --dry-run`.
5. Confirm the current tree and full history pass secret scanning.
6. Commit and push the verified release preparation.

The GitHub repository needs an Actions secret named `NPM_TOKEN`. Store the token in the project credential system and send it to GitHub without writing it to the repository or a shell history file.

## Publish

Create and push an annotated tag that matches `package.json`:

```bash
version=$(node -p "require('./package.json').version")
git tag -a "v${version}" -m "UniFi MCP Server ${version}"
git push origin "v${version}"
```

The release workflow reruns tests, vet, staticcheck, and `govulncheck`; builds four native binaries and the MCPB bundle; generates a CycloneDX SBOM and SHA-256 checksums; creates the GitHub release; and publishes the npm package with provenance. It skips the npm publish step when that exact version already exists, which makes a failed workflow safe to rerun.

## Verify

- Confirm the GitHub release, npm package, and MCPB manifest show the same version.
- Download the release assets and run `shasum -a 256 -c SHA256SUMS`.
- Run the downloaded binary with no credentials and confirm it starts in needs-auth mode.
- Follow `docs/INTEGRATION_TESTING.md` for an optional live controller smoke test.
