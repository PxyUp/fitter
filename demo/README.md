# Fitter Playground (WebAssembly)

A browser playground where the real fitter engine — the same code as the CLI and
MCP server — runs fully client-side as WebAssembly. No backend needed, which makes
it deployable on any static host.

CI (`.github/workflows/ci.yaml`, `pages` job) builds this into a `pages/` folder on
every push to master and deploys it to GitHub Pages. Repo settings → Pages →
Source must be set to "GitHub Actions".

## Build & run locally

```bash
# from the repo root
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o demo/main.wasm ./cmd/wasm
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" demo/

# serve (any static server works)
cd demo && python3 -m http.server 8642
# open http://localhost:8642
```

## What works in the browser

- All parsers: JSON, HTML, XML, XPath, PDF
- `static_config`, `int_sequence_config`, references, expressions, generated fields
- Live HTTP fetching via the browser Fetch API — for APIs that send CORS headers
  (GitHub, OpenLibrary, CoinGecko, …)

## What doesn't

- Fetching sites without CORS headers (browser security model — use the native CLI/MCP)
- `browser_config` emulation (Chromium/Docker/Playwright) — stubbed out in the WASM
  build via `//go:build !js` tags, returns a clear error

## How it's wired

- `cmd/wasm/main.go` exposes `fitterRun(configJSON, input) -> Promise<string>` via
  `syscall/js`; configs are accepted as JSON or YAML
- `pkg/connectors/browser_js.go` provides the WASM stubs for the native-only connectors
- `index.html` is the playground UI with preloaded example configs and a form-driven
  config builder (Builder tab). The builder covers the common subset — URL/static/sequence
  sources, object/array models, field transforms; configs beyond that (nested models,
  references, limits) stay editable on the JSON tab, with a notice instead of silent loss
- `sample.pdf` is bundled so the PDF example fetches same-origin (`__BASE__/sample.pdf`
  in the example resolves to the demo's own URL — no CORS involved)
