# Examples

Ready-to-run Fitter configs. Two formats live here:

- **CLI/MCP format** — a single top-level `item` (+ optional `limits`/`references`). Works with **Fitter_CLI**, **Fitter_MCP** (`fitter_run_file`/`fitter_run_url`) and **Fitter_Agent**.
- **Service format** — a top-level `items` array (+ optional `http_server`, per-item `trigger_config`/`notifier_config`). For the long-running **Fitter** service binary.

## CLI / MCP configs (single `item`)

| Config | What it does |
|---|---|
| [config_morning_briefing.json](config_morning_briefing.json) | Morning briefing: weather from wttr.in (city via `{{{FromInput=.}}}`) + top HackerNews stories in one nested pipeline |
| [config_google_news.json](config_google_news.json) | Latest headlines from the Google News RSS feed (XPath over XML) |
| [config_trading_signals.json](config_trading_signals.json) | Crypto trading signals: CoinGecko market data with references and host rate-limits |
| [config_github_trending.json](config_github_trending.json) | GitHub trending (HTML scrape — no official API) enriched per-repo from the GitHub REST API via `{PL}` |
| [config_book_authors.json](config_book_authors.json) | OpenLibrary book search (query via `{{{FromInput=.}}}`) with per-book author details via a `{{{FromExp=...}}}` JSON-field join |
| [config_crypto_csv.json](config_crypto_csv.json) | Top-5 crypto coins written to a local CSV with `file_storage` (`{HUMAN_INDEX}` rank column) |
| [cli/config_cli.json](cli/config_cli.json) | HackerNews + quotes: JSON API, browser connector, limits — the kitchen-sink demo |
| [cli/config_browser.json](cli/config_browser.json) | Headless browser (Chromium) scraping |
| [cli/config_playwright.json](cli/config_playwright.json) | Playwright-driven scraping |
| [cli/config_docker.json](cli/config_docker.json) | Browser in a Docker container |
| [cli/config_ref.json](cli/config_ref.json) | Prefetched references (shared cached values) |
| [cli/config_seq.json](cli/config_seq.json) | Int-sequence connector (pagination-style loops) |
| [cli/config_static_connector.json](cli/config_static_connector.json) | Static connector |
| [cli/config_current_time.json](cli/config_current_time.json) | Current time scraped from time.is |
| [cli/config_weather.json](cli/config_weather.json) | UK top cities + weather forecast, chained scraping |
| [cli/config_image.json](cli/config_image.json) | Download an image to disk |
| [cli/config_image_multiple.json](cli/config_image_multiple.json) | Download multiple images |
| [cli/config_plugin.json](cli/config_plugin.json) | Custom plugin field (see [plugin/README.md](plugin/README.md)) |

## Service configs (`items` array)

| Config | What it does |
|---|---|
| [config_api.json](config_api.json) | JSON API extraction (Chronicling America) |
| [config_web.json](config_web.json) | HTML page scraping (W3Schools) |
| [config_static.json](config_static.json) | HTTP-triggered scraping exposed via the built-in HTTP server (port 8080) |
| [config_telegram.json](config_telegram.json) | Scrape time.is and push the result to a Telegram bot via `notifier_config` |

## Other

- [go/](go) — using fitter as a Go library
- [plugin/](plugin) — writing and building connector/field plugins
