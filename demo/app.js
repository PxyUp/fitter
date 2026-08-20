// app.js — SPA entry point: wires the hash router to the views and keeps the
// header's engine status badge in sync. Loaded as a module from index.html.
import { addRoute, setNotFound, start } from "./router.js";
import { onStatusChange } from "./engine.js";
import "./theme.js";
import { renderHome } from "./view-home.js";
import { renderPlayground } from "./view-playground.js";
import { renderDocs } from "./view-docs.js";
import { renderExamplesList, renderExampleDetail } from "./view-examples.js";

const app = document.getElementById("app");
let currentCleanup = null;

function mount(viewFn) {
  return async (params) => {
    if (typeof currentCleanup === "function") {
      try { currentCleanup(); } catch { /* ignore */ }
      currentCleanup = null;
    }
    const result = await viewFn(app, params);
    if (typeof result === "function") currentCleanup = result;
  };
}

addRoute("/", mount(renderHome));
addRoute("/playground", mount((el, params) => renderPlayground(el, params)));
addRoute("/playground/builder", mount((el) => renderPlayground(el, { builder: true })));
addRoute("/docs", mount((el) => renderDocs(el, {})));
addRoute("/docs/:slug", mount(renderDocs));
addRoute("/examples", mount(renderExamplesList));
addRoute("/examples/:id", mount(renderExampleDetail));

setNotFound((path) => {
  app.innerHTML = `<div class="not-found"><h1>Page not found</h1><p><code>${path}</code> doesn't match a route.</p><p><a href="#/">← Home</a></p></div>`;
});

// engine status badge in the header — visible on every route once the
// engine has been asked to load at least once (playground / an example open)
const badge = document.getElementById("engineBadge");
onStatusChange((status, detail) => {
  if (status === "idle") {
    badge.hidden = true;
    return;
  }
  badge.hidden = false;
  if (status === "loading") {
    badge.textContent = "engine loading…";
    badge.className = "engine-badge loading";
  } else if (status === "ready") {
    badge.textContent = "engine ready";
    badge.className = "engine-badge ready";
  } else if (status === "error") {
    badge.textContent = "engine error";
    badge.className = "engine-badge error";
    badge.title = String(detail || "");
  }
});

// A bare "?q=<token>" URL (no hash) is the legacy/canonical share-link
// shape — route it straight into the playground, which reads location.search
// itself. New Share links always append "#/playground" too, so this only
// matters for links captured before the SPA rewrite.
if (new URLSearchParams(location.search).has("q") && (!location.hash || location.hash === "#")) {
  location.hash = "/playground";
}

start();
