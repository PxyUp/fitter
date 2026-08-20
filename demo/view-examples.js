// view-examples.js — the examples gallery (list + detail), each runnable
// example deep-linking into the Playground via pending.js.
import { navigate } from "./router.js";
import { setPendingExample } from "./pending.js";
import { highlightJson } from "./jsonhl.js";

let examplesPromise = null;
function loadExamples() {
  if (!examplesPromise) examplesPromise = fetch("examples.json").then(r => r.json());
  return examplesPromise;
}

function tagsHtml(tags) {
  return (tags || []).map(t => `<span class="tag-chip">${t}</span>`).join("");
}

function cardHtml(ex) {
  if (ex.nativeOnly) {
    return `
      <div class="example-card example-card-native">
        <div class="example-card-cat">${ex.category}</div>
        <h3><a href="#/examples/${ex.id}">${ex.title}</a></h3>
        <p>${ex.description}</p>
        <div class="example-tags">${tagsHtml(ex.tags)}</div>
        <div class="example-card-actions">
          <span class="native-only-badge">native-only</span>
          <a class="btn btn-ghost btn-sm" href="#/examples/${ex.id}">Details</a>
          ${ex.config ? `<button class="btn btn-primary btn-sm" data-open="${ex.id}">Open in Playground ▶</button>` : ""}
        </div>
      </div>`;
  }
  return `
    <div class="example-card">
      <div class="example-card-cat">${ex.category}</div>
      <h3><a href="#/examples/${ex.id}">${ex.title}</a></h3>
      <p>${ex.description}</p>
      <div class="example-tags">${tagsHtml(ex.tags)}</div>
      <div class="example-card-actions">
        <a class="btn btn-ghost btn-sm" href="#/examples/${ex.id}">Details</a>
        <button class="btn btn-primary btn-sm" data-open="${ex.id}">Open in Playground ▶</button>
      </div>
    </div>`;
}

function wireOpenButtons(root) {
  root.querySelectorAll("[data-open]").forEach(btn => {
    btn.addEventListener("click", async () => {
      const examples = await loadExamples();
      const ex = examples.find(e => e.id === btn.getAttribute("data-open"));
      if (!ex) return;
      setPendingExample(ex);
      navigate("/playground");
    });
  });
}

export async function renderExamplesList(app) {
  app.innerHTML = `
    <section class="examples-header">
      <h1>Examples</h1>
      <p>A runnable card for every connector, parser and feature variant fitter supports &mdash; each opens straight
      into the <a href="#/playground">Playground</a>, pre-loaded and already run against the real WASM engine.
      <code>browser_config</code> (Chromium/Docker/Playwright) can't execute inside a browser sandbox, so it's
      documented as native-only instead of faked &mdash; see the cards below.</p>
      <div class="examples-controls">
        <input type="search" id="exampleFilter" class="docs-search" placeholder="Filter examples…" autocomplete="off">
        <div class="examples-cat-filter" id="exampleCatFilter"></div>
      </div>
    </section>
    <div class="examples-grid" id="examplesGrid">loading…</div>
  `;
  const examples = await loadExamples();
  const categories = [...new Set(examples.map(e => e.category))];
  const catFilter = document.getElementById("exampleCatFilter");
  // the user may have navigated away while examples.json was loading
  if (!catFilter) return;
  catFilter.innerHTML = `<button class="cat-chip active" data-cat="">All</button>` +
    categories.map(c => `<button class="cat-chip" data-cat="${c}">${c}</button>`).join("");

  const grid = document.getElementById("examplesGrid");
  const filterInput = document.getElementById("exampleFilter");
  let activeCat = "";

  function paint() {
    const q = filterInput.value.trim().toLowerCase();
    const filtered = examples.filter(e => {
      if (activeCat && e.category !== activeCat) return false;
      if (!q) return true;
      const hay = (e.title + " " + e.description + " " + (e.tags || []).join(" ")).toLowerCase();
      return hay.includes(q);
    });
    grid.innerHTML = filtered.map(cardHtml).join("") || `<p class="no-results">No examples match.</p>`;
    wireOpenButtons(grid);
  }

  filterInput.addEventListener("input", paint);
  catFilter.addEventListener("click", e => {
    const btn = e.target.closest("[data-cat]");
    if (!btn) return;
    activeCat = btn.getAttribute("data-cat");
    catFilter.querySelectorAll(".cat-chip").forEach(c => c.classList.toggle("active", c === btn));
    paint();
  });

  paint();
}

export async function renderExampleDetail(app, params) {
  app.innerHTML = `<div class="boot-loading">loading…</div>`;
  const examples = await loadExamples();
  const ex = examples.find(e => e.id === params.id);
  if (!ex) {
    app.innerHTML = `<div class="not-found"><h1>Example not found</h1><p><a href="#/examples">← Back to Examples</a></p></div>`;
    return;
  }

  const inputPreview = ex.input === undefined ? null :
    (typeof ex.input === "string" ? ex.input : JSON.stringify(ex.input, null, 2));

  app.innerHTML = `
    <div class="example-detail">
      <p><a href="#/examples">← All examples</a></p>
      <div class="example-detail-cat">${ex.category}</div>
      <h1>${ex.title}</h1>
      <p class="example-detail-desc">${ex.description}</p>
      <div class="example-tags">${tagsHtml(ex.tags)}</div>
      ${ex.nativeOnly ? `
        <div class="notice-box">This feature needs a real OS process (browser/filesystem/plugin) and can't <strong>run</strong> inside
        the WebAssembly sandbox this playground uses — Run stays disabled once it's loaded there. You can still open it to read
        and copy the config. ${ex.docsSlug ? `See <a href="#/docs/${ex.docsSlug}">the docs</a> for details, or run it with the native fitter_cli / MCP server.` : "Run it with the native fitter_cli / MCP server."}</div>
      ` : ""}
      ${ex.config ? `
        <div class="example-detail-actions">
          <button class="btn btn-primary" data-open="${ex.id}">Open in Playground ▶</button>
        </div>
        ${inputPreview !== null && inputPreview !== "" ? `<h3>Input</h3><pre class="code-block">${highlightJson(inputPreview)}</pre>` : ""}
        <h3>Config</h3>
        <pre class="code-block">${highlightJson(JSON.stringify(ex.config, null, 2))}</pre>
      ` : ""}
    </div>
  `;
  wireOpenButtons(app);
}
