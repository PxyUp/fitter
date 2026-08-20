# Fitter demo site (SPA)

A no-build, vanilla-JS single-page app that is simultaneously the interactive Playground,
the documentation, and the examples gallery for fitter. The real fitter engine — the same
Go code as the CLI and MCP server — is compiled to WebAssembly and runs entirely
client-side, which makes the whole thing deployable on any static host (no backend).

CI (`.github/workflows/ci.yaml`, `pages` job) builds this into a `pages/` folder on every
push to master and deploys it to GitHub Pages. Repo settings → Pages → Source must be set
to "GitHub Actions".

## Build & run locally

```bash
# from the repo root
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o demo/main.wasm ./cmd/wasm
WASM_EXEC="$(go env GOROOT)/lib/wasm/wasm_exec.js"
[ -f "$WASM_EXEC" ] || WASM_EXEC="$(go env GOROOT)/misc/wasm/wasm_exec.js"
cp "$WASM_EXEC" demo/

# serve (any static server works)
cd demo && python3 -m http.server 8642
# open http://localhost:8642
```

`main.wasm` and `wasm_exec.js` are gitignored build artifacts — don't commit them.

## Routes (client-side hash router, see `router.js`)

- `#/` — hero/landing page (`view-home.js`)
- `#/playground` (and `#/playground/builder`) — the JSON editor + form Builder (`view-playground.js`)
- `#/docs` and `#/docs/<slug>` — docs list/detail with client-side search (`view-docs.js`, content in `docs.json`)
- `#/examples` and `#/examples/<id>` — the examples gallery (`view-examples.js`, content in `examples.json`)

`app.js` is the entry point: it wires routes to views and keeps the header's engine-status
badge in sync. `engine.js` lazy-loads the ~30 MB `main.wasm` (once per page load, cached) —
only triggered when the playground view mounts or an example is opened, never blocking the
docs/home views. `share.js` holds the `?q=` share-link encode/decode and `__BASE__`
resolution; `pending.js` is the in-memory handoff used by "Open in Playground" buttons;
`md.js` is a tiny markdown-lite renderer for `docs.json` content, with JSON code fences
syntax-highlighted via `jsonhl.js`; `theme.js` drives the header's dark/light toggle.

`live.html` is a separate, unrouted page (a terminal-style "live run" of the GitHub profile
config, embedded as a dynamic image in the profile README) — intentionally left outside the
SPA/router; its `?raw=` query param must keep working.

## What works in the browser

- All parsers: JSON, HTML, XML, XPath, PDF
- `static_config`, `int_sequence_config`, `reference_config`/`references` (see the
  `reference-as-source` and `telegram-briefing` examples), expressions, generated fields,
  conditional fields (`condition`/`item_condition`, incl. `fSrc`), `first_of`, `static_array`
- Live HTTP fetching via the browser Fetch API — for APIs that send CORS headers
  (GitHub, OpenLibrary, CoinGecko, Hacker News Algolia, httpbin, …)

## What doesn't (documented as native-only, not faked)

- Fetching sites without CORS headers (browser security model — use the native CLI/MCP)
- `browser_config` emulation (Chromium/Docker/Playwright) — stubbed out in the WASM build
  via `pkg/connectors/browser_js.go` (`//go:build js`), returns a clear error
- `file_config`, `plugin_connector_config` and the `file`/`file_storage`/`plugin` generated
  values — no real local filesystem or native `.so` plugin loading inside the WASM sandbox
- `notifier_config` — not invoked by the bare `fitterRun`/`ParseCtx` call this playground's
  WASM bridge uses (it's driven by the long-running native service/MCP tool calls); the
  Telegram example demonstrates the same mechanics manually via `generated.model` instead

## How it's wired

- `cmd/wasm/main.go` exposes `fitterRun(configJSON, input) -> Promise<string>` via
  `syscall/js`; configs are accepted as JSON or YAML. Because the page is a long-lived
  process running many unrelated configs back to back, it calls `limitter.ReplaceLimits`
  and `references.ReplaceReferences` before every run — both `pkg/limitter` and
  `pkg/references` otherwise apply only the *first* config's limits/references for the
  lifetime of the process (correct for the native one-shot CLI, wrong here)
- `pkg/connectors/browser_js.go` provides the WASM stubs for the native-only connectors
- `sample.pdf` is bundled so the PDF example fetches same-origin (`__BASE__/sample.pdf`
  resolves to the demo's own URL — no CORS involved)

## Theming (`theme.js`, `styles.css`)

Every color is a CSS custom property scoped to `:root`. The full (dark) palette lives on
`:root` itself; `:root[data-theme="dark"]`/`:root[data-theme="light"]` override it for an
explicit manual choice, and `@media (prefers-color-scheme: light)` (guarded with
`:root:not([data-theme])`) is the fallback while no choice has been made. The header's
`#themeToggle` button (`theme.js`) flips `data-theme` and persists the choice to
`localStorage["fitter-theme"]`; a tiny inline snippet in `index.html`'s `<head>` applies a
saved choice before first paint (no flash). Code/JSON surfaces (docs `<pre>`, the examples
gallery's config/input blocks) use a dedicated `--code-*` token set (background, text, and
per-token-kind colors for `jsonhl.js`'s highlighter) so they stay legible in both themes
independent of the general `--panel`/`--text` tokens — the Playground's own JSON editor and
Builder intentionally keep their original look (transparent, following `--text`/`--bg`) and
are not part of that dedicated code-surface styling.

## Docs content & search (`docs.json`, `view-docs.js`, `md.js`)

`docs.json` is a hand-authored, prebuilt index: an array of `{slug, title, category,
summary, body}`, cross-checked against `pkg/config/config.go`, `pkg/config/field.go` and
`cmd/mcp/reference.go` (the condensed schema reference) — `README.md` at the repo root is
the richest human explanation and prose source. `body` is "markdown-lite" (fenced code
blocks, `##`/`###` headings, `- ` lists, `> ` callouts, inline `` ` `` /`**`/`[]()`),
rendered by `md.js`. Search in `view-docs.js` is a simple client-side scored substring match
over title/summary/category/body — no external service, no build step.

## Examples matrix (`examples.json`, `view-examples.js`)

One entry per connector/parser/feature variant, each with a config that's cross-checked by
actually running it in the playground. Entries with `"nativeOnly": true` (`browser_config`,
`file_config`, `generated.file`/`file_storage`, `plugin_connector_config`/`generated.plugin`
— four cards) still ship a real, complete `config` — they render with a "native-only" badge
and notice, and "Open in Playground" still loads them (viewable/copyable, and still usable
from the Builder's JSON tab) but the Playground disables **Run** and explains why instead of
letting it fail with a WASM stub error. Detection (`isNativeOnlyConfig` in
`view-playground.js`) scans the *live* config text for these connectors/generated values, so
it also catches someone hand-writing or pasting one — not just the four gallery cards. "Open
in Playground" hands the example off via `pending.js` and navigates to `#/playground`, which
loads it and auto-runs it once the engine is ready (skipped for native-only configs).

## Share links (`?q=…`)

The **Share** button packs the current config + input into the URL as
`<origin>/<path>?q=<token>#/playground`, so the link reproduces the exact playground state —
on open, `?q=` is read directly off `location.search` (independent of hash routing) and the
config runs automatically once the engine loads. A bare `?q=…` with no `#` (old-style links)
is redirected to `#/playground` by `app.js`.

Token format: `<version>.<base64url payload>` where the payload is
`{"c": "<config text>", "i": "<input text>"}` — version `1` is deflate-raw compressed (via
`CompressionStream`), version `0` is plain JSON for browsers without compression support.
Invalid tokens are ignored and the default example loads instead.

## Adding a new asset

Any new file under `demo/` (js/css/json/html/image) deploys automatically — the `pages` CI
job does `cp -r demo/. pages/` after building `main.wasm`/`wasm_exec.js` into `pages/`, and
`demo/` never contains Go sources or those two gitignored build artifacts. Nothing to edit
in `ci.yaml` for a new demo asset; only touch it if the *build* step itself changes.
