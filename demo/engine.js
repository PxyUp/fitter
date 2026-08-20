// engine.js — lazy-loads the ~30MB fitter WASM binary (once, cached) and
// exposes a promise-based runner. wasm_exec.js (classic script, defines the
// global `Go`) must already be loaded before this module runs.

let enginePromise = null;
let status = "idle"; // idle -> loading -> ready | error
let statusDetail = null;
const listeners = new Set();

function setStatus(s, detail) {
  status = s;
  statusDetail = detail || null;
  for (const fn of listeners) fn(status, statusDetail);
}

// Registers a listener and immediately replays the current status, so views
// mounted after the engine already started/finished loading render correctly.
export function onStatusChange(fn) {
  listeners.add(fn);
  fn(status, statusDetail);
  return () => listeners.delete(fn);
}

export function engineStatus() {
  return status;
}

export function ensureEngine() {
  if (enginePromise) return enginePromise;
  setStatus("loading");
  enginePromise = (async () => {
    const go = new Go();
    const result = await (WebAssembly.instantiateStreaming
      ? WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject)
      : fetch("main.wasm").then(r => r.arrayBuffer()).then(b => WebAssembly.instantiate(b, go.importObject)));
    go.run(result.instance);
    setStatus("ready");
    return true;
  })().catch(err => {
    setStatus("error", err);
    enginePromise = null; // allow a retry on next call
    throw err;
  });
  return enginePromise;
}

// Runs a config through the real engine. Resolves with the raw string result
// (usually JSON) fitterRun() produces; rejects with the engine's error string.
export async function runFitter(configStr, inputStr) {
  await ensureEngine();
  return await window.fitterRun(configStr, inputStr || "");
}
