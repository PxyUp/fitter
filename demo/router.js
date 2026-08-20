// router.js — minimal client-side hash router. No dependencies.
//
// Routes are registered with a pattern like "/", "/docs/:slug", "/examples/:id".
// The router matches on location.hash (the part after "#") and calls the
// handler with { params, path }. `?query` strings are left on location.search
// and are NOT part of routing — they're read directly via location.search by
// whoever needs them (e.g. the playground's ?q= share token).

const routes = [];
let notFoundHandler = null;
let currentPath = null;

function toRegex(pattern) {
  const paramNames = [];
  const escaped = pattern
    .split("/")
    .map(seg => {
      if (seg.startsWith(":")) {
        paramNames.push(seg.slice(1));
        return "([^/]+)";
      }
      return seg.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
    })
    .join("/");
  return { regex: new RegExp("^" + escaped + "$"), paramNames };
}

export function addRoute(pattern, handler) {
  const { regex, paramNames } = toRegex(pattern);
  routes.push({ pattern, regex, paramNames, handler });
}

export function setNotFound(handler) {
  notFoundHandler = handler;
}

export function currentRoutePath() {
  return currentPath;
}

function normalizedHash() {
  let h = location.hash.slice(1); // drop leading '#'
  if (!h) return "/";
  // support "#/docs/slug?ignored" defensively — routing ignores anything after '?'
  const q = h.indexOf("?");
  if (q >= 0) h = h.slice(0, q);
  if (h.length > 1 && h.endsWith("/")) h = h.slice(0, -1);
  return h.startsWith("/") ? h : "/" + h;
}

export function resolve() {
  const path = normalizedHash();
  currentPath = path;
  for (const r of routes) {
    const m = path.match(r.regex);
    if (m) {
      const params = {};
      r.paramNames.forEach((name, i) => { params[name] = decodeURIComponent(m[i + 1]); });
      r.handler(params, path);
      updateActiveNav(path);
      window.scrollTo(0, 0);
      return;
    }
  }
  if (notFoundHandler) notFoundHandler(path);
  updateActiveNav(path);
}

function updateActiveNav(path) {
  const nav = document.getElementById("siteNav");
  if (!nav) return;
  const top = "/" + (path.split("/")[1] || "");
  for (const a of nav.querySelectorAll("a[data-route]")) {
    const r = a.getAttribute("data-route");
    a.classList.toggle("active", r === path || (r !== "/" && top === r));
  }
}

export function navigate(path) {
  location.hash = path;
}

export function start() {
  window.addEventListener("hashchange", resolve);
  resolve();
}
