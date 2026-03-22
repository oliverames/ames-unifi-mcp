#!/usr/bin/env node

// Postinstall script: copies the correct platform binary to bin/ames-unifi-mcp

const fs = require("fs");
const path = require("path");
const os = require("os");

const BINARY_NAME = "ames-unifi-mcp";

const PLATFORM_MAP = {
  "darwin-arm64": `${BINARY_NAME}-darwin-arm64`,
  "darwin-x64": `${BINARY_NAME}-darwin-amd64`,
  "linux-arm64": `${BINARY_NAME}-linux-arm64`,
  "linux-x64": `${BINARY_NAME}-linux-amd64`,
};

function install() {
  const platform = os.platform();
  const arch = os.arch();
  const key = `${platform}-${arch}`;

  const binaryFile = PLATFORM_MAP[key];
  if (!binaryFile) {
    console.error(
      `Unsupported platform: ${platform}-${arch}. ` +
        `Supported: ${Object.keys(PLATFORM_MAP).join(", ")}`
    );
    process.exit(1);
  }

  const distDir = path.join(__dirname, "dist");
  const binDir = path.join(__dirname, "bin");
  const src = path.join(distDir, binaryFile);
  const dest = path.join(binDir, BINARY_NAME);

  if (!fs.existsSync(src)) {
    console.error(`Binary not found: ${src}`);
    console.error("Try rebuilding: make build-all");
    process.exit(1);
  }

  // Ensure bin/ exists
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }

  // Copy binary
  fs.copyFileSync(src, dest);

  // Make executable
  fs.chmodSync(dest, 0o755);

  console.log(`Installed ${BINARY_NAME} for ${platform}/${arch}`);
}

install();
