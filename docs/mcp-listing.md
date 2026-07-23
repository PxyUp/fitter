# MCP directory listing copy

Paste-ready copy for MCP directories (official registry, Glama, PulseMCP, mcp.so, Smithery, Awesome MCP lists). Keep the one-liner and description in sync everywhere — consistent wording helps search.

## Name

`fitter` — Web data for AI agents

## One-liner (≤120 chars)

> Turn any website or API into structured JSON with declarative scraping configs the LLM authors itself.

## Short description (≤350 chars)

> Fitter is a local-first scraping engine for AI agents. The LLM writes a small JSON/YAML config (HTTP or headless browser + JSON/HTML/XML/XPath selectors), validates it, and runs it on your machine — no scraping API, no keys, no code. Configs are reusable: save them, re-run them, schedule them.

## Long description

Fitter exposes a declarative web-extraction engine over MCP. Instead of generating throwaway scraping code, the model authors a **config** — where the data lives (HTTP request, Playwright/Chromium browser, static value, file) and what to extract (gjson paths, CSS selectors, XPath) — then executes it locally and gets clean JSON back.

**Tools:**

| Tool | Purpose |
|------|---------|
| `fitter_run` | Run an inline JSON/YAML config, return extracted JSON |
| `fitter_run_file` | Run a config from a local file |
| `fitter_run_url` | Run a config downloaded from an HTTP(S) URL |
| `fitter_validate_config` | Validate a config without executing it |
| `fitter_config_reference` | Condensed format reference so the model can author configs without external docs (also exposed as MCP resource `fitter://config-reference`) |

**Why it's different:**

- **Auditable** — you can read exactly what the agent fetched and how
- **Local-first** — data never goes through a third-party scraping service
- **Reusable** — configs the agent writes become cron jobs via fitter's service mode
- **Complete** — pagination, nested per-item requests, cached auth references, host rate limits, browser rendering in one static Go binary

**Example prompts:**

- "Get the top 5 HackerNews stories with titles and scores"
- "Scrape headlines and links from this news site"
- "Check the Bitcoin price and 24h range"
- "Parse this RSS feed and give me the latest articles"

## Metadata

- **Categories:** web scraping, data extraction, automation, research
- **Transport:** stdio (default) + streamable HTTP (`--http :8080`, optional bearer auth)
- **Platforms:** macOS (amd64/arm64), Linux (amd64/arm64), Windows (amd64)
- **Install:** `.mcpb` one-click bundle or raw binary from [GitHub releases](https://github.com/PxyUp/fitter/releases), Docker image `ghcr.io/pxyup/fitter-mcp`, or `go build -o fitter_mcp ./cmd/mcp`
- **License:** MIT

## Claude Code registration

```bash
claude mcp add fitter -s user -- /path/to/fitter_mcp
```

## Claude Desktop registration

```json
{
  "mcpServers": {
    "fitter": { "command": "/path/to/fitter_mcp" }
  }
}
```

## Submission checklist

- [x] Official MCP registry — **automated**: the release workflow packs per-platform `.mcpb` bundles (`scripts/build_mcpb.bash`), uploads them as release assets, fills `server.json` (version + `fileSha256` per bundle) and publishes via `mcp-publisher` with GitHub OIDC. Runs on the next `v*.*.*` tag.
- [x] Add `mcp` + `mcp-server` + `ai-agents` GitHub topics to the repo
- [ ] [Glama](https://glama.ai/mcp/servers) — auto-indexes GitHub (topics help); claim the listing
- [ ] [PulseMCP](https://www.pulsemcp.com) — submit form
- [ ] [mcp.so](https://mcp.so) — submit form
- [ ] Awesome MCP servers lists on GitHub (PR)

## How the registry publishing works

- `server.json` (repo root) is a template: `__TAG__`, `__VERSION__` and `__SHA_*__` placeholders are substituted in CI — do not put real values in the committed file.
- `scripts/build_mcpb.bash <tag>` zips each `fitter_mcp` release binary with a generated MCPB `manifest.json` into `bin/fitter-mcp-<os>-<arch>.mcpb`.
- The registry entry is `io.github.PxyUp/fitter` — the namespace is case-sensitive and must match the GitHub username exactly; GitHub OIDC from this repo authorizes it (workflow permission `id-token: write`).
- MCP clients verify `fileSha256` before install; the hashes are computed from the exact uploaded bundles in the same job.
- The `oci` package points at `ghcr.io/pxyup/fitter-mcp:<tag>` (built from `Dockerfile.mcp`, multi-arch). The registry verifies it via the `io.modelcontextprotocol.server.name` image label, so the registry publish step must run after the docker push — keep that step order in `release.yaml`.
