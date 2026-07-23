#!/usr/bin/env bash
#
# Package fitter_mcp release binaries into per-platform .mcpb bundles
# (zip with a manifest.json) for the MCP registry and Claude Desktop
# one-click install.
#
# Usage: ./scripts/build_mcpb.bash <version-tag>       e.g. v1.0.19
#
# Expects the binaries already built by:
#   ./scripts/build.bash mcp fitter_mcp_<version-tag>

set -euo pipefail

version=${1:-}
if [[ -z "$version" ]]; then
  echo "usage: $0 <version-tag>"
  exit 1
fi
version_num=${version#v}

# goos/goarch/mcpb-platform-id
platforms=(
  "darwin/arm64/darwin"
  "darwin/amd64/darwin"
  "linux/arm64/linux"
  "linux/amd64/linux"
  "windows/amd64/win32"
)

for entry in "${platforms[@]}"; do
  IFS=/ read -r goos goarch mcpb_platform <<<"$entry"

  bin_src="bin/fitter_mcp_${version}-${goos}-${goarch}"
  bin_name="fitter_mcp"
  if [[ "$goos" == "windows" ]]; then
    bin_src+=".exe"
    bin_name+=".exe"
  fi

  if [[ ! -f "$bin_src" ]]; then
    echo "missing $bin_src — run ./scripts/build.bash mcp fitter_mcp_${version} first"
    exit 1
  fi

  stage=$(mktemp -d)
  mkdir -p "$stage/server"
  cp "$bin_src" "$stage/server/$bin_name"
  chmod +x "$stage/server/$bin_name"

  cat >"$stage/manifest.json" <<EOF
{
  "manifest_version": "0.3",
  "name": "fitter",
  "display_name": "Fitter — web data for AI agents",
  "version": "${version_num}",
  "description": "Turn any website or API into structured JSON with declarative scraping configs the LLM authors itself. HTTP, headless browser, JSON/HTML/XML/XPath — local-first, no scraping API keys.",
  "author": {
    "name": "PxyUp",
    "url": "https://github.com/PxyUp"
  },
  "repository": {
    "type": "git",
    "url": "https://github.com/PxyUp/fitter"
  },
  "homepage": "https://github.com/PxyUp/fitter",
  "license": "MIT",
  "keywords": ["scraping", "web-data", "json", "browser", "extraction"],
  "server": {
    "type": "binary",
    "entry_point": "server/${bin_name}",
    "mcp_config": {
      "command": "\${__dirname}/server/${bin_name}",
      "args": [],
      "env": {}
    }
  },
  "compatibility": {
    "platforms": ["${mcpb_platform}"]
  }
}
EOF

  out="bin/fitter-mcp-${goos}-${goarch}.mcpb"
  rm -f "$out"
  (cd "$stage" && zip -qr "$OLDPWD/$out" .)
  rm -rf "$stage"
  echo "built $out"
done
