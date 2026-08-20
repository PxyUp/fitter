// theme.js — the header's manual dark/light toggle. The initial choice (if
// any was previously saved) is applied synchronously by a tiny inline
// snippet in index.html's <head> — before first paint, so there's no flash
// of the wrong theme. This module just wires up the toggle button, persists
// changes to localStorage, and keeps the button's label in sync with the OS
// setting for as long as the user hasn't made an explicit choice. Every view
// reacts live because the whole site is styled off CSS custom properties
// scoped to :root[data-theme=...] (see styles.css) — flipping the attribute
// re-paints everything instantly, no reload needed.

const KEY = "fitter-theme";
const root = document.documentElement;
const media = window.matchMedia ? window.matchMedia("(prefers-color-scheme: light)") : null;

function storedTheme() {
  const t = localStorage.getItem(KEY);
  return t === "light" || t === "dark" ? t : null;
}

function effectiveTheme() {
  return storedTheme() || (media && media.matches ? "light" : "dark");
}

function paint(btn) {
  const isLight = effectiveTheme() === "light";
  btn.textContent = isLight ? "\u{1F319} Dark" : "☀️ Light";
  const label = isLight ? "Switch to dark theme" : "Switch to light theme";
  btn.title = label;
  btn.setAttribute("aria-label", label);
}

function apply(theme) {
  localStorage.setItem(KEY, theme);
  root.setAttribute("data-theme", theme);
}

function init() {
  const btn = document.getElementById("themeToggle");
  if (!btn) return;
  paint(btn);
  btn.addEventListener("click", () => {
    apply(effectiveTheme() === "light" ? "dark" : "light");
    paint(btn);
  });
  if (media) {
    media.addEventListener("change", () => {
      if (!storedTheme()) paint(btn); // only repaint the label; CSS follows the media query itself
    });
  }
}

init();
