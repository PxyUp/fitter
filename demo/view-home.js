// view-home.js — the hero/landing view.
export function renderHome(app) {
  app.innerHTML = `
    <section class="hero">
      <h1>Declarative web-data extraction, running as <em>real WebAssembly</em> in your browser</h1>
      <p class="hero-lead">
        Fitter turns a JSON/YAML config into structured data: fetch a page or API, parse it (JSON, HTML, XML,
        XPath, PDF), reshape it into exactly the object or array you want &mdash; conditions, expressions, nested
        fetches and all. The same Go engine that powers the CLI and MCP server is compiled to WebAssembly and runs
        this entire site's playground and examples client-side. Nothing here is a canned demo.
      </p>
      <div class="hero-actions">
        <a class="btn btn-primary" href="#/playground">Open the Playground</a>
        <a class="btn btn-ghost" href="#/examples">Browse Examples</a>
        <a class="btn btn-ghost" href="#/docs">Read the Docs</a>
      </div>
    </section>

    <section class="pillars">
      <a class="pillar-card" href="#/playground">
        <h3>Playground</h3>
        <p>A JSON editor and a no-code form Builder side by side. Write or build a config, hit Run, and watch the
        real engine execute it &mdash; live HTTP calls included. Share any config as a link.</p>
      </a>
      <a class="pillar-card" href="#/docs">
        <h3>Docs</h3>
        <p>Every connector, parser, field type, generated value, condition, expression, placeholder and notifier
        &mdash; searchable, cross-linked to runnable examples.</p>
      </a>
      <a class="pillar-card" href="#/examples">
        <h3>Examples</h3>
        <p>A runnable card for every connector &times; parser combination and every feature &mdash; each one opens
        straight into the Playground, pre-loaded and already run.</p>
      </a>
      <a class="pillar-card" href="#/docs/install">
        <h3>MCP &amp; CLI</h3>
        <p>Give an AI agent <code>fitter_run</code> / <code>fitter_run_file</code> tools, or run configs natively
        &mdash; including headless-browser connectors this WASM build can't run. See the install guide for Claude
        Code, Claude Desktop, Cursor and Docker.</p>
      </a>
    </section>

    <section class="honesty-note">
      <h3>What "real engine" means here</h3>
      <p>
        Every run on this site &mdash; Playground, every Examples card &mdash; calls the actual compiled fitter
        binary (<code>cmd/wasm</code>, ~30&nbsp;MB / ~7&nbsp;MB gzipped, loaded once and cached) via
        <code>fitterRun(config, input)</code>. Live HTTP examples only work against CORS-friendly public APIs
        (GitHub, OpenLibrary, CoinGecko, Hacker News Algolia, httpbin&hellip;) because the browser &mdash; not
        fitter &mdash; enforces CORS. Headless-browser connectors (Chromium/Docker/Playwright) and anything
        touching the local filesystem or native plugins can't run in a browser sandbox at all; those are
        documented as native-only rather than faked. See <a href="#/docs/connectors">Connectors</a> for details.
      </p>
    </section>
  `;
}
