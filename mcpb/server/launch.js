#!/usr/bin/env node
import { spawn } from "node:child_process";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

const dirname = path.dirname(fileURLToPath(import.meta.url));
const binaryName = "ames-unifi-mcp";
const platformMap = {
  "darwin-arm64": `${binaryName}-darwin-arm64`,
  "darwin-x64": `${binaryName}-darwin-amd64`,
  "linux-arm64": `${binaryName}-linux-arm64`,
  "linux-x64": `${binaryName}-linux-amd64`,
};

const key = `${os.platform()}-${os.arch()}`;
const binary = platformMap[key];

if (!binary) {
  console.error(
    `Unsupported platform: ${key}. Supported: ${Object.keys(platformMap).join(", ")}`
  );
  process.exit(1);
}

const binaryPath = path.join(dirname, "dist", binary);
const child = spawn(binaryPath, process.argv.slice(2), {
  stdio: "inherit",
  env: process.env,
});

child.on("error", (error) => {
  console.error(`Failed to start ${binaryName}: ${error.message}`);
  process.exit(1);
});

child.on("exit", (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
    return;
  }
  process.exit(code ?? 0);
});
