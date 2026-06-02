#!/usr/bin/env node
import { execFileSync } from "node:child_process";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const distDir = path.join(root, "dist");
const stageDir = path.join(distDir, "mcpb-staging");
const pkg = JSON.parse(fs.readFileSync(path.join(root, "package.json"), "utf8"));
const manifestPath = path.join(root, "mcpb", "manifest.json");
const outputPath = path.join(distDir, `${pkg.name}-${pkg.version}.mcpb`);
const mcpbPackage = "@anthropic-ai/mcpb@2.1.2";

const binaries = [
  "ames-unifi-mcp-darwin-arm64",
  "ames-unifi-mcp-darwin-amd64",
  "ames-unifi-mcp-linux-arm64",
  "ames-unifi-mcp-linux-amd64",
];

function copyFile(src, dest) {
  fs.mkdirSync(path.dirname(dest), { recursive: true });
  fs.copyFileSync(src, dest);
}

fs.rmSync(stageDir, { recursive: true, force: true });
fs.mkdirSync(stageDir, { recursive: true });

copyFile(path.join(root, "README.md"), path.join(stageDir, "README.md"));
copyFile(path.join(root, "LICENSE"), path.join(stageDir, "LICENSE"));
copyFile(path.join(root, "icon.png"), path.join(stageDir, "icon.png"));
copyFile(path.join(root, "assets", "icon.png"), path.join(stageDir, "assets", "icon.png"));
copyFile(path.join(root, "mcpb", "server", "launch.js"), path.join(stageDir, "server", "launch.js"));

for (const binary of binaries) {
  const src = path.join(distDir, binary);
  if (!fs.existsSync(src)) {
    throw new Error(`Missing release binary: ${src}. Run make build-all first.`);
  }
  copyFile(src, path.join(stageDir, "server", "dist", binary));
  fs.chmodSync(path.join(stageDir, "server", "dist", binary), 0o755);
}

const manifest = JSON.parse(fs.readFileSync(manifestPath, "utf8"));
manifest.version = pkg.version;
manifest.description = pkg.description;
fs.writeFileSync(path.join(stageDir, "manifest.json"), `${JSON.stringify(manifest, null, 2)}\n`);

execFileSync("npx", ["-y", mcpbPackage, "validate", path.join(stageDir, "manifest.json")], {
  cwd: root,
  stdio: "inherit",
});

fs.rmSync(outputPath, { force: true });
execFileSync("npx", ["-y", mcpbPackage, "pack", stageDir, outputPath], {
  cwd: root,
  stdio: "inherit",
});

execFileSync("npx", ["-y", mcpbPackage, "info", outputPath], {
  cwd: root,
  stdio: "inherit",
});

console.log(`Built ${outputPath}`);
