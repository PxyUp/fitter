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
| `fitter_validate_config` | Validate a config without executing it |
| `fitter_config_reference` | Condensed format reference so the model can author configs without external docs |

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
- **Transport:** stdio
- **Platforms:** macOS (amd64/arm64), Linux (amd64/arm64), Windows (amd64)
- **Install:** [GitHub releases](https://github.com/PxyUp/fitter/releases) or `go build -o fitter_mcp ./cmd/mcp`
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

- [ ] Official MCP registry (`server.json` in repo root; requires publishing via `mcp-publisher` CLI — needs an mcpb bundle or OCI image, see note below)
- [ ] [Glama](https://glama.ai/mcp/servers) — auto-indexes GitHub; claim the listing
- [ ] [PulseMCP](https://www.pulsemcp.com) — submit form
- [ ] [mcp.so](https://mcp.so) — submit form
- [ ] Awesome MCP servers lists on GitHub (PR)
- [ ] Add `mcp` + `mcp-server` GitHub topics to the repo

> **Note:** the official registry accepts packages as npm/pypi/nuget/oci/mcpb. Fitter ships raw Go binaries, so the cleanest path is publishing a small **OCI image** (scratch + fitter_mcp binary) to ghcr.io, or packaging an `.mcpb` bundle per release. Verify the current schema at modelcontextprotocol.io/registry before submitting — `server.json` in the repo root is a draft.
