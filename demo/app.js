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

// A "?q=<token>" URL is a share link — it must open the playground and load
// that config no matter what hash it's paired with (a bare URL, the legacy
// no-hash shape, or someone having replaced "#/playground" with another
// route like "#/examples"). Force-route to the playground, overriding
// whatever hash is currently set; the playground itself reads
// location.search to decode the token.
if (new URLSearchParams(location.search).has("q")) {
  const h = location.hash.replace(/^#/, "");
  if (h !== "/playground" && !h.startsWith("/playground/")) {
    location.hash = "/playground";
  }
}

// Once a share token has done its job of getting the user into the
// playground, don't let it linger in the address bar past a navigation to a
// different route — otherwise copying the URL from, say, "#/examples" would
// silently carry along someone else's (or a stale) shared config. This is
// distinct from view-playground.js's own "q" clearing, which fires on
// same-route actions inside the playground (picking an example, discarding
// an invalid shared config) and never sees a hashchange event.
function topPathFromHash(hash) {
  let h = hash.replace(/^#/, "");
  if (!h) return "/";
  const q = h.indexOf("?");
  if (q >= 0) h = h.slice(0, q);
  if (h.length > 1 && h.endsWith("/")) h = h.slice(0, -1);
  if (!h.startsWith("/")) h = "/" + h;
  return "/" + (h.split("/")[1] || "");
}
window.addEventListener("hashchange", (e) => {
  if (!new URLSearchParams(location.search).has("q")) return;
  const oldTop = topPathFromHash(new URL(e.oldURL).hash);
  const newTop = topPathFromHash(new URL(e.newURL).hash);
  if (oldTop === "/playground" && newTop !== "/playground") {
    history.replaceState(null, "", location.pathname + location.hash);
  }
});

start();
