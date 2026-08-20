// view-playground.js — the JSON editor + form Builder playground, migrated
// from the old standalone demo/index.html with feature parity: Builder
// import/export round-trip, examples, ?q= share links, __BASE__ resolution,
// lazy WASM load with a status indicator, and the per-example send counter.
import { ensureEngine, onStatusChange, runFitter } from "./engine.js";
import { encodeShare, decodeShare, BASE } from "./share.js";
import { takePendingExample } from "./pending.js";

let examplesPromise = null;
function loadExamples() {
  // fetches every example, including native-only ones (Chromium/Docker/
  // Playwright/file/plugin) — those aren't offered in the picker below, but
  // "Open in Playground" from the Examples gallery still needs to look them
  // up here, with Run disabled once loaded (see isNativeOnlyConfig()).
  if (!examplesPromise) examplesPromise = fetch("examples.json").then(r => r.json());
  return examplesPromise;
}

// Detects, from the raw config text (JSON or YAML, valid or not — this runs
// on every keystroke), whether it uses a connector or generated value that
// needs a real OS process (headless browser, local filesystem, a compiled
// plugin) and therefore can't execute inside the WASM sandbox. Used to
// disable Run with an explanation instead of letting it fail with a WASM
// stub error — matters both for the two native-only example cards AND for
// anyone hand-writing/pasting such a config here.
function isNativeOnlyConfig(str) {
  if (!str || !str.trim()) return false;
  // browser_config / file_config / plugin_connector_config are unambiguous,
  // exact key names — safe to detect textually regardless of JSON vs YAML
  // or even invalid/in-progress JSON while typing.
  if (/(^|[^\w])"?(browser_config|file_config|plugin_connector_config)"?\s*:/m.test(str)) return true;

  // generated.file / generated.plugin (download-to-disk / write-to-disk /
  // custom-plugin generated values) are only native-only when "file" or
  // "plugin" is a key *inside* a "generated" block — a plain field named
  // "file" elsewhere (e.g. an object_config field) is not. Try a real parse
  // first (covers the common JSON case precisely); fall back to an
  // indentation-aware line scan for YAML / not-yet-valid JSON.
  try {
    const parsed = JSON.parse(str);
    return containsGeneratedFileOrPlugin(parsed);
  } catch {
    return scanGeneratedFileOrPlugin(str);
  }
}
function containsGeneratedFileOrPlugin(node) {
  if (!node || typeof node !== "object") return false;
  if (Array.isArray(node)) return node.some(containsGeneratedFileOrPlugin);
  const g = node.generated;
  if (g && typeof g === "object" && !Array.isArray(g) && ("file" in g || "file_storage" in g || "plugin" in g)) return true;
  for (const v of Object.values(node)) {
    if (v && typeof v === "object" && containsGeneratedFileOrPlugin(v)) return true;
  }
  return false;
}
function scanGeneratedFileOrPlugin(str) {
  const lines = str.split("\n");
  for (let i = 0; i < lines.length; i++) {
    const m = lines[i].match(/^(\s*)generated\s*:\s*$/);
    if (!m) continue;
    const indent = m[1].length;
    for (let j = i + 1; j < lines.length; j++) {
      const line = lines[j];
      if (line.trim() === "") continue;
      const lineIndent = line.match(/^(\s*)/)[1].length;
      if (lineIndent <= indent) break;
      if (/^\s*(file|file_storage|plugin)\s*:/.test(line)) return true;
    }
  }
  return false;
}

const TEMPLATE = `
<section class="playground">
  <div class="playground-panel panel">
    <div class="panel-head">
      <div class="tabs">
        <button class="tab" id="tabBuilder" type="button">Builder</button>
        <button class="tab active" id="tabJson" type="button">JSON</button>
      </div>
      <select id="example"></select>
    </div>

    <div id="builderPane">
      <div id="builderNotice" class="notice" hidden>
        <p style="margin:0 0 8px;">This config uses features the builder can’t express (plugins, file fields,
        first_of, references, browser connectors, …). It stays fully editable — and runnable — on the JSON tab.</p>
        <button id="bFresh" class="ghost" type="button">discard it and start a fresh config in the builder</button>
      </div>

      <div id="bForm" style="display:flex;flex-direction:column;gap:14px;">
      <fieldset>
        <legend>Source</legend>
        <div class="brow">
          <select id="bSource">
            <option value="url">Fetch URL</option>
            <option value="static">Static value</option>
            <option value="sequence">Int sequence</option>
            <option value="reference">Reference (cached)</option>
          </select>
          <label for="bResponseType">parse as</label>
          <select id="bResponseType">
            <option value="json">json</option>
            <option value="HTML">HTML</option>
            <option value="xpath">xpath</option>
            <option value="XML">XML</option>
            <option value="pdf">pdf</option>
          </select>
        </div>
        <div class="brow" data-src="url">
          <select id="bMethod">
            <option>GET</option><option>POST</option><option>PUT</option><option>DELETE</option>
          </select>
          <input type="text" id="bUrl" placeholder="https://api.example.com/… ({{{FromInput=.}}} injects the input box)">
        </div>
        <div class="brow" data-src="static">
          <textarea id="bStatic" spellcheck="false" placeholder='raw body to parse, e.g. {"name": "fitter"} or <html>…</html>'></textarea>
        </div>
        <div class="brow" data-src="sequence">
          <label>start</label><input type="number" id="bSeqStart" value="1">
          <label>end</label><input type="number" id="bSeqEnd" value="5">
          <label>step</label><input type="number" id="bSeqStep" value="1">
        </div>
        <div class="brow" data-src="reference">
          <label>name</label><input type="text" id="bRefName" placeholder="reference name defined below, e.g. hn_top">
        </div>
      </fieldset>

      <fieldset>
        <legend>Model</legend>
        <div class="brow">
          <select id="bKind">
            <option value="object">Single object</option>
            <option value="array">Array of items</option>
            <option value="value">Single value</option>
          </select>
          <select id="bItemShape" hidden>
            <option value="fields">items: named fields</option>
            <option value="single">items: single value</option>
          </select>
          <input type="text" id="bRootPath" placeholder="root path of the array (e.g. docs, @this, //item)" hidden>
        </div>
        <div class="brow">
          <input type="text" id="bItemCondition" placeholder="item condition — drop items when false, e.g. fRes.price > 0 (optional)" hidden>
        </div>
        <div id="bFields" style="display:flex;flex-direction:column;gap:8px;"></div>
        <button class="ghost" id="bAddField" type="button">+ add field</button>
        <div class="hint">path syntax follows “parse as”: gjson for json/pdf, CSS selectors for HTML, XPath for xpath/XML</div>
      </fieldset>

      <fieldset>
        <legend>References</legend>
        <div id="bRefs" style="display:flex;flex-direction:column;gap:12px;"></div>
        <button class="ghost" id="bAddRef" type="button">+ add reference</button>
        <div class="hint">named cached sub-configs — use one as the Source (“Reference”) or inline via {{{RefName=name path}}} / {{{RefName=name}}}</div>
      </fieldset>

      <fieldset>
        <legend>Limits</legend>
        <div id="bHostLimits" style="display:flex;flex-direction:column;gap:8px;"></div>
        <button class="ghost" id="bAddHostLimit" type="button">+ host rate limit</button>
        <div class="hint">max parallel requests per host (host_request_limiter)</div>
      </fieldset>
      </div>
    </div>

    <textarea id="config" spellcheck="false"></textarea>

    <div id="nativeOnlyNotice" class="notice" hidden>
      This config uses a connector or generated value that needs a real OS process (headless browser, local
      filesystem, or a compiled plugin) — it can't run inside the WebAssembly sandbox this playground uses, so
      <b>Run</b> is disabled. Copy it and run it with the native <code>fitter_cli</code> or MCP server instead.
    </div>

    <div class="runbar">
      <textarea id="input" rows="1" spellcheck="false" placeholder="input (optional — used by {{{FromInput=.}}}; JSON is pretty-printed)"></textarea>
      <button id="share" class="ghost" type="button" title="copy a link that opens and runs this config">Share</button>
      <button id="run" class="primary" disabled>Run</button>
    </div>
  </div>

  <div class="playground-panel panel">
    <div class="panel-head">
      <span class="label">Result</span>
      <span id="engineStatusLocal" class="engine-status-local">loading engine…</span>
      <span id="sendCounter" hidden title="live global counter for this example"><span id="counterLabel"></span><b id="sendCount" style="color:var(--ok)">0</b></span>
      <span id="timing"></span>
    </div>
    <pre id="output">Engine loading — the fitter Go binary is compiled to WebAssembly (~7&nbsp;MB gzipped), first load takes a moment…</pre>
  </div>
</section>
`;

export async function renderPlayground(app, params) {
  app.innerHTML = TEMPLATE;

  const $ = id => document.getElementById(id);
  const runBtn = $("run"), output = $("output"), timing = $("timing");
  const configEl = $("config");
  const select = $("example");
  const inputEl = $("input");
  const statusLocal = $("engineStatusLocal");
  const nativeOnlyNotice = $("nativeOnlyNotice");

  let engineReady = false;
  let nativeOnly = false;
  function syncRunEnabled() {
    runBtn.disabled = !engineReady || nativeOnly;
  }
  function updateNativeOnlyState() {
    nativeOnly = isNativeOnlyConfig(configEl.value);
    nativeOnlyNotice.hidden = !nativeOnly;
    syncRunEnabled();
  }

  const examples = await loadExamples();
  const examplesById = new Map(examples.map(e => [e.id, e]));
  const runnableExamples = examples.filter(e => !e.nativeOnly);
  for (const ex of runnableExamples) {
    const opt = document.createElement("option");
    opt.value = ex.id;
    opt.textContent = ex.title;
    select.appendChild(opt);
  }

  function autosizeInput() {
    const lines = inputEl.value.split("\n").length;
    inputEl.rows = Math.min(Math.max(lines, 1), 7);
  }
  function setInput(v) {
    try {
      const parsed = JSON.parse(v);
      if (parsed && typeof parsed === "object") v = JSON.stringify(parsed, null, 2);
    } catch { /* not JSON: leave as plain text */ }
    inputEl.value = v;
    autosizeInput();
  }
  inputEl.addEventListener("input", autosizeInput);
  inputEl.addEventListener("blur", () => setInput(inputEl.value));

  function exampleInputString(ex) {
    if (ex.input === undefined) return "";
    return typeof ex.input === "string" ? ex.input : JSON.stringify(ex.input);
  }

  function loadExample(id) {
    const ex = examplesById.get(id);
    if (!ex) return;
    if (ex.nativeOnly) markPseudoOption(`— ${ex.title} (native-only) —`);
    configEl.value = JSON.stringify(ex.config, null, 2).replaceAll("__BASE__", BASE);
    setInput(exampleInputString(ex));
    syncBuilderFromJson();
    updateSendCounterVis();
    updateNativeOnlyState();
  }
  select.addEventListener("change", () => {
    if (!examplesById.has(select.value)) return; // the injected pseudo-option itself
    clearPseudoOption();
    loadExample(select.value);
  });

  /* ---------- pseudo dropdown entries: share links (?q=…) and native-only
     examples opened from the gallery, neither of which is a real <option> ---------- */
  const PSEUDO_VALUE = "__pseudo__";

  function clearPseudoOption() {
    if (location.search) history.replaceState(null, "", location.pathname + location.hash);
    for (const opt of [...select.options]) {
      if (opt.value === PSEUDO_VALUE) opt.remove();
    }
  }
  function markPseudoOption(label) {
    clearPseudoOption();
    const opt = document.createElement("option");
    opt.textContent = label;
    opt.value = PSEUDO_VALUE;
    select.prepend(opt);
    select.value = PSEUDO_VALUE;
  }

  /* ---------- builder state ---------- */
  const FIELD_TYPES = ["string", "int", "float", "boolean", "array", "object"];
  const RESPONSE_TYPES = ["json", "HTML", "xpath", "XML", "pdf"];
  let fieldSeq = 0;

  function defaultState() {
    return {
      source: "url", url: "", method: "GET",
      staticValue: "", seqStart: 1, seqEnd: 5, seqStep: 1,
      responseType: "json",
      kind: "object", rootPath: "", itemShape: "fields", itemCondition: "",
      fields: [newField()],
      refName: "",
      references: [],
      hostLimits: [], limitsExtra: null, referencesRaw: null
    };
  }
  function newField() {
    return { id: ++fieldSeq, name: "", type: "string", path: "", condition: "", transform: "none", expr: "", calcType: null, nested: null };
  }
  function defaultNested() {
    return {
      url: "", method: "GET", responseType: "json", nullOnError: true,
      headers: [], body: "", timeout: "",
      kind: "object", rootPath: "", itemShape: "fields", itemCondition: "", fields: [newField()]
    };
  }
  let bState = defaultState();
  let builderCanRepresent = true;

  /* ---------- builder -> config ---------- */
  function buildFieldConfig(f) {
    if (f.transform === "first_of") {
      return { first_of: (f.alts || []).map(a => buildFieldConfig(a)) };
    }
    const base = { type: f.type };
    if (f.path !== "") base.path = f.path;
    if (f.condition) base.condition = f.condition;
    if (f.transform === "calculated" && f.expr !== "") {
      base.generated = { calculated: { type: f.calcType || f.type, expression: f.expr } };
    } else if (f.transform === "formatted" && f.expr !== "") {
      base.generated = { formatted: { template: f.expr } };
    } else if (f.transform === "uuid") {
      base.generated = { uuid: f.expr !== "" ? { regexp: f.expr } : {} };
    } else if (f.transform === "static") {
      base.generated = { static: { type: f.calcType || f.type, value: f.expr } };
    } else if (f.transform === "model" && f.nested) {
      const n = f.nested;
      const cc = { response_type: n.responseType, url: n.url, server_config: buildServerConfig(n) };
      if (n.nullOnError) cc.null_on_error = true;
      base.generated = { model: { type: f.type, connector_config: cc, model: buildModelPart(n) } };
    }
    return { base_field: base };
  }

  function buildServerConfig(n) {
    const sc = { method: n.method };
    const hs = {};
    for (const h of n.headers || []) {
      if (h.k.trim() !== "") hs[h.k.trim()] = h.v;
    }
    if (Object.keys(hs).length) sc.headers = hs;
    if (n.body) sc.body = n.body;
    if (n.timeout !== "" && n.timeout != null) sc.timeout = Number(n.timeout);
    return sc;
  }

  function buildModelPart(p) {
    if (p.kind === "value") {
      return { base_field: buildFieldConfig(p.fields[0] || newField()).base_field };
    }
    const fields = {};
    for (const f of p.fields) {
      if (f.name.trim() !== "") fields[f.name.trim()] = buildFieldConfig(f);
    }
    if (p.kind === "object") return { object_config: { fields } };
    const ac = {};
    if (p.itemShape === "single") {
      ac.item_config = { field: buildFieldConfig(p.fields[0] || newField()).base_field };
    } else {
      ac.item_config = { fields };
    }
    if (p.rootPath !== "") ac.root_path = p.rootPath;
    if (p.itemCondition) ac.item_condition = p.itemCondition;
    return { array_config: ac };
  }

  function buildConfig(s) {
    const cc = { response_type: s.responseType };
    if (s.source === "url") {
      cc.url = s.url;
      cc.server_config = { method: s.method };
    } else if (s.source === "static") {
      cc.static_config = { value: s.staticValue };
    } else if (s.source === "reference") {
      cc.reference_config = { name: s.refName };
    } else {
      cc.int_sequence_config = { start: Number(s.seqStart), end: Number(s.seqEnd), step: Number(s.seqStep) };
    }

    const out = { item: { connector_config: cc, model: buildModelPart(s) } };
    const limits = Object.assign({}, s.limitsExtra || {});
    const hostMap = {};
    for (const hl of s.hostLimits) {
      if (hl.host.trim() !== "") hostMap[hl.host.trim()] = Number(hl.n) || 1;
    }
    if (Object.keys(hostMap).length) limits.host_request_limiter = hostMap;
    if (Object.keys(limits).length) out.limits = limits;
    if (s.references && s.references.length) {
      const refs = {};
      for (const r of s.references) {
        if (r.name.trim() === "") continue;
        const rcc = { response_type: r.responseType, url: r.url, server_config: buildServerConfig(r) };
        if (r.nullOnError) rcc.null_on_error = true;
        const entry = { connector_config: rcc, model: buildModelPart(r) };
        if (r.expire !== "" && r.expire != null) entry.expire = Number(r.expire);
        refs[r.name.trim()] = entry;
      }
      if (Object.keys(refs).length) out.references = refs;
    } else if (s.referencesRaw) {
      out.references = s.referencesRaw;
    }
    return out;
  }

  /* ---------- config -> builder (import) ---------- */
  function importField(fieldCfg) {
    const keys = Object.keys(fieldCfg || {});
    if (keys.length !== 1) return null;
    if (keys[0] === "first_of") {
      if (!Array.isArray(fieldCfg.first_of) || fieldCfg.first_of.length === 0) return null;
      const alts = fieldCfg.first_of.map(importField);
      if (alts.some(a => !a || a.transform === "first_of")) return null;
      const f = newField();
      f.transform = "first_of";
      f.alts = alts;
      return f;
    }
    if (keys[0] !== "base_field") return null;
    return importBaseField(fieldCfg.base_field);
  }

  function importBaseField(b) {
    const allowed = new Set(["type", "path", "condition", "generated"]);
    if (!b || typeof b !== "object" || Object.keys(b).some(k => !allowed.has(k))) return null;
    if (!FIELD_TYPES.includes(b.type)) return null;
    if (b.condition != null && typeof b.condition !== "string") return null;

    const f = newField();
    f.type = b.type;
    f.path = b.path || "";
    f.condition = b.condition || "";
    if (b.generated) {
      const gk = Object.keys(b.generated);
      if (gk.length !== 1) return null;
      if (gk[0] === "calculated") {
        const c = b.generated.calculated;
        if (!c || Object.keys(c).some(k => !["type", "expression"].includes(k))) return null;
        f.transform = "calculated"; f.expr = c.expression || ""; f.calcType = c.type || null;
      } else if (gk[0] === "formatted") {
        const fm = b.generated.formatted;
        if (!fm || Object.keys(fm).some(k => k !== "template")) return null;
        f.transform = "formatted"; f.expr = fm.template || "";
      } else if (gk[0] === "uuid") {
        const u = b.generated.uuid;
        if (!u || Object.keys(u).some(k => k !== "regexp")) return null;
        f.transform = "uuid"; f.expr = u.regexp || "";
      } else if (gk[0] === "static") {
        const st = b.generated.static;
        if (!st || Object.keys(st).some(k => !["type", "value"].includes(k))) return null;
        f.transform = "static"; f.expr = st.value || ""; f.calcType = st.type || null;
      } else if (gk[0] === "model") {
        const m = b.generated.model;
        if (!m || Object.keys(m).some(k => !["type", "connector_config", "model"].includes(k))) return null;
        if (m.type !== b.type) return null;
        const mcc = m.connector_config || {};
        if (Object.keys(mcc).some(k => !["response_type", "url", "server_config", "null_on_error"].includes(k))) return null;
        if (!RESPONSE_TYPES.includes(mcc.response_type) || typeof mcc.url !== "string") return null;
        const msc = importServerConfig(mcc.server_config);
        if (!msc) return null;
        const part = importModelPart(m.model || {});
        if (!part) return null;
        f.transform = "model";
        f.nested = {
          url: mcc.url, ...msc,
          responseType: mcc.response_type, nullOnError: !!mcc.null_on_error,
          kind: part.kind, rootPath: part.rootPath, itemShape: part.itemShape, itemCondition: part.itemCondition, fields: part.fields
        };
      } else return null;
    }
    return f;
  }

  function importServerConfig(sc) {
    sc = sc || {};
    if (Object.keys(sc).some(k => !["method", "headers", "body", "timeout"].includes(k))) return null;
    if (sc.headers && (typeof sc.headers !== "object" || Object.values(sc.headers).some(v => typeof v !== "string"))) return null;
    if (sc.body != null && typeof sc.body !== "string") return null;
    return {
      method: sc.method || "GET",
      headers: Object.entries(sc.headers || {}).map(([k, v]) => ({ id: ++fieldSeq, k, v })),
      body: sc.body || "",
      timeout: sc.timeout ?? ""
    };
  }

  function importReference(name, r) {
    if (!r || typeof r !== "object") return null;
    if (Object.keys(r).some(k => !["connector_config", "model", "expire"].includes(k))) return null;
    const cc = r.connector_config || {};
    if (Object.keys(cc).some(k => !["response_type", "url", "server_config", "null_on_error"].includes(k))) return null;
    if (!RESPONSE_TYPES.includes(cc.response_type) || typeof cc.url !== "string") return null;
    const sc = importServerConfig(cc.server_config);
    if (!sc) return null;
    const part = importModelPart(r.model || {});
    if (!part) return null;
    return {
      id: ++fieldSeq, name, expire: r.expire ?? "",
      url: cc.url, ...sc,
      responseType: cc.response_type, nullOnError: !!cc.null_on_error,
      kind: part.kind, rootPath: part.rootPath, itemShape: part.itemShape, itemCondition: part.itemCondition, fields: part.fields
    };
  }

  function importModelPart(model) {
    const mk = Object.keys(model || {});
    if (mk.length !== 1) return null;
    const part = { kind: "object", rootPath: "", itemShape: "fields", itemCondition: "", fields: [] };
    let fieldsCfg;
    if (mk[0] === "base_field") {
      const f = importBaseField(model.base_field);
      if (!f) return null;
      part.kind = "value";
      part.fields.push(f);
      return part;
    }
    if (mk[0] === "object_config") {
      const oc = model.object_config;
      if (!oc || Object.keys(oc).some(k => k !== "fields")) return null;
      fieldsCfg = oc.fields || {};
    } else if (mk[0] === "array_config") {
      const ac = model.array_config;
      if (!ac || Object.keys(ac).some(k => !["root_path", "item_config", "item_condition"].includes(k))) return null;
      if (ac.item_condition != null && typeof ac.item_condition !== "string") return null;
      part.kind = "array";
      part.rootPath = ac.root_path || "";
      part.itemCondition = ac.item_condition || "";
      const ic = ac.item_config || {};
      const icKeys = Object.keys(ic);
      if (icKeys.some(k => !["fields", "field"].includes(k)) || icKeys.length > 1) return null;
      if (ic.field) {
        const f = importBaseField(ic.field);
        if (!f) return null;
        part.itemShape = "single";
        part.fields.push(f);
        fieldsCfg = {};
      } else {
        fieldsCfg = ic.fields || {};
      }
    } else return null;

    for (const [name, fc] of Object.entries(fieldsCfg)) {
      const f = importField(fc);
      if (!f) return null;
      f.name = name;
      part.fields.push(f);
    }
    if (part.fields.length === 0) part.fields.push(newField());
    return part;
  }

  function importConfig(cfg) {
    if (!cfg || typeof cfg !== "object") return null;
    if (Object.keys(cfg).some(k => !["item", "limits", "references"].includes(k))) return null;
    const item = cfg.item;
    if (!item || Object.keys(item).some(k => !["connector_config", "model"].includes(k))) return null;

    const s = defaultState();
    s.fields = [];
    if (cfg.references && typeof cfg.references === "object") {
      const refs = Object.entries(cfg.references).map(([name, r]) => importReference(name, r));
      if (refs.every(Boolean)) {
        s.references = refs;
      } else {
        s.referencesRaw = cfg.references; // beyond the UI: keep verbatim, still runnable
      }
    }
    if (cfg.limits && typeof cfg.limits === "object") {
      const { host_request_limiter, ...rest } = cfg.limits;
      s.limitsExtra = Object.keys(rest).length ? rest : null;
      if (host_request_limiter && typeof host_request_limiter === "object") {
        for (const [h, v] of Object.entries(host_request_limiter)) {
          s.hostLimits.push({ id: ++fieldSeq, host: h, n: v });
        }
      }
    }

    const cc = item.connector_config || {};
    const ccAllowed = new Set(["response_type", "url", "server_config", "static_config", "int_sequence_config", "reference_config"]);
    if (Object.keys(cc).some(k => !ccAllowed.has(k))) return null;
    if (!RESPONSE_TYPES.includes(cc.response_type)) return null;
    s.responseType = cc.response_type;

    if (cc.reference_config) {
      if (Object.keys(cc.reference_config).some(k => k !== "name")) return null;
      if (cc.url !== undefined || cc.static_config || cc.int_sequence_config || cc.server_config) return null;
      s.source = "reference"; s.refName = cc.reference_config.name || "";
    } else if (cc.url !== undefined) {
      s.source = "url"; s.url = cc.url;
      const sc = cc.server_config || {};
      if (Object.keys(sc).some(k => k !== "method")) return null;
      s.method = sc.method || "GET";
      if (cc.static_config || cc.int_sequence_config) return null;
    } else if (cc.static_config) {
      if (Object.keys(cc.static_config).some(k => k !== "value")) return null;
      if (cc.server_config || cc.int_sequence_config) return null;
      s.source = "static"; s.staticValue = cc.static_config.value || "";
    } else if (cc.int_sequence_config) {
      if (cc.server_config) return null;
      const q = cc.int_sequence_config;
      s.source = "sequence"; s.seqStart = q.start ?? 0; s.seqEnd = q.end ?? 0; s.seqStep = q.step ?? 1;
    } else return null;

    const part = importModelPart(item.model || {});
    if (!part) return null;
    s.kind = part.kind;
    s.rootPath = part.rootPath;
    s.itemShape = part.itemShape;
    s.itemCondition = part.itemCondition;
    s.fields = part.fields;
    return s;
  }

  /* ---------- builder UI ---------- */
  function setBuilderRepresentable(ok) {
    builderCanRepresent = ok;
    $("builderNotice").hidden = ok;
    $("bForm").style.display = ok ? "" : "none";
  }

  function renderBuilder() {
    $("bSource").value = bState.source;
    $("bResponseType").value = bState.responseType;
    $("bMethod").value = bState.method;
    $("bUrl").value = bState.url;
    $("bStatic").value = bState.staticValue;
    $("bSeqStart").value = bState.seqStart;
    $("bSeqEnd").value = bState.seqEnd;
    $("bSeqStep").value = bState.seqStep;
    $("bKind").value = bState.kind;
    $("bRefName").value = bState.refName;
    $("bRootPath").value = bState.rootPath;
    $("bRootPath").hidden = bState.kind !== "array";
    $("bItemCondition").value = bState.itemCondition;
    $("bItemCondition").hidden = bState.kind !== "array";
    $("bItemShape").value = bState.itemShape;
    $("bItemShape").hidden = bState.kind !== "array";
    for (const row of document.querySelectorAll("[data-src]")) {
      row.style.display = row.dataset.src === bState.source ? "" : "none";
    }

    const singleItem = (bState.kind === "array" && bState.itemShape === "single") || bState.kind === "value";
    $("bAddField").style.display = singleItem ? "none" : "";
    if (singleItem && bState.fields.length === 0) bState.fields.push(newField());

    renderFieldList($("bFields"), bState.fields, singleItem);
    renderReferences();
    renderHostLimits();
  }

  function renderReferences() {
    const wrap = $("bRefs");
    wrap.innerHTML = "";
    for (const r of bState.references) {
      const box = document.createElement("div");
      box.className = "subpanel";

      const head = document.createElement("div");
      head.className = "srow";
      const name = document.createElement("input");
      name.type = "text"; name.placeholder = "reference name"; name.value = r.name;
      name.addEventListener("input", () => { r.name = name.value; builderChanged(); });
      const expire = document.createElement("input");
      expire.type = "number"; expire.min = "0"; expire.placeholder = "∞"; expire.value = r.expire;
      expire.style.width = "80px"; expire.title = "cache TTL in seconds (empty = fetched once per run)";
      expire.addEventListener("input", () => { r.expire = expire.value; builderChanged(); });
      const del = document.createElement("button");
      del.type = "button"; del.className = "del"; del.textContent = "×"; del.title = "remove reference";
      del.addEventListener("click", () => {
        const i = bState.references.findIndex(x => x.id === r.id);
        if (i >= 0) bState.references.splice(i, 1);
        renderBuilder(); builderChanged();
      });
      head.append(name, slabel("expire s"), expire, del);

      box.append(head, nestedPanel(r));
      wrap.appendChild(box);
    }
  }

  function renderHostLimits() {
    const wrap = $("bHostLimits");
    wrap.innerHTML = "";
    for (const hl of bState.hostLimits) {
      const row = document.createElement("div");
      row.className = "srow";

      const host = document.createElement("input");
      host.type = "text"; host.placeholder = "host, e.g. api.github.com"; host.value = hl.host;
      host.addEventListener("input", () => { hl.host = host.value; builderChanged(); });

      const n = document.createElement("input");
      n.type = "number"; n.min = "1"; n.value = hl.n; n.style.width = "90px"; n.title = "max parallel requests";
      n.addEventListener("input", () => { hl.n = n.value; builderChanged(); });

      const del = document.createElement("button");
      del.type = "button"; del.className = "del"; del.textContent = "×"; del.title = "remove limit";
      del.addEventListener("click", () => {
        const i = bState.hostLimits.findIndex(x => x.id === hl.id);
        if (i >= 0) bState.hostLimits.splice(i, 1);
        renderBuilder(); builderChanged();
      });

      row.append(host, n, del);
      wrap.appendChild(row);
    }
  }

  const PLACEHOLDERS = {
    calculated: 'e.g. upper(fRes) + "!"',
    formatted: "e.g. value: {PL}",
    static: "the static value",
    uuid: "optional regexp applied to the uuid"
  };
  const NO_EXPR = ["none", "model", "first_of"];

  function renderFieldList(container, fields, singleItem, isAlt) {
    container.innerHTML = "";
    for (const f of (singleItem ? fields.slice(0, 1) : fields)) {
      container.appendChild(fieldRow(f, fields, singleItem, isAlt));
    }
  }

  function fieldRow(f, siblings, singleItem, isAlt) {
    const row = document.createElement("div");
    row.className = "frow";

    const name = document.createElement("input");
    name.type = "text"; name.placeholder = "field name"; name.value = f.name;
    name.addEventListener("input", () => { f.name = name.value; builderChanged(); });
    if (singleItem) { name.style.display = "none"; }

    const type = document.createElement("select");
    for (const t of FIELD_TYPES) { const o = document.createElement("option"); o.textContent = t; type.appendChild(o); }
    type.value = f.type;
    type.addEventListener("change", () => { f.type = type.value; f.calcType = null; builderChanged(); });

    const path = document.createElement("input");
    path.type = "text"; path.placeholder = "path"; path.value = f.path;
    path.addEventListener("input", () => { f.path = path.value; builderChanged(); });

    const del = document.createElement("button");
    del.type = "button"; del.className = "del"; del.textContent = "×"; del.title = "remove field";
    del.addEventListener("click", () => {
      const i = siblings.findIndex(x => x.id === f.id);
      if (i >= 0) siblings.splice(i, 1);
      if (siblings.length === 0) siblings.push(newField());
      renderBuilder(); builderChanged();
    });
    if (singleItem) {
      del.style.display = "none";
      row.style.gridTemplateColumns = "auto minmax(90px, 1.4fr)";
    }

    row.append(name, type, path, del);

    const isFirstOf = f.transform === "first_of";
    if (isFirstOf) {
      type.style.display = "none"; path.style.display = "none";
      row.style.gridTemplateColumns = singleItem ? "auto" : "minmax(80px, 1fr) auto";
    }

    const xtra = document.createElement("div");
    xtra.className = "xtra";
    const tf = document.createElement("select");
    const tfOptions = [
      ["none", "no generated value"],
      ["calculated", "generated: expression (fRes)"],
      ["formatted", "generated: template ({PL})"],
      ["static", "generated: static value"],
      ["uuid", "generated: uuid"],
      ["model", "generated: nested fetch (model)"]
    ];
    if (!isAlt && !singleItem) tfOptions.push(["first_of", "first_of: fallback alternatives"]);
    for (const [v, label] of tfOptions) {
      const o = document.createElement("option"); o.value = v; o.textContent = label; tf.appendChild(o);
    }
    tf.value = f.transform;
    const expr = document.createElement("input");
    expr.type = "text";
    expr.placeholder = PLACEHOLDERS[f.transform] || "";
    expr.value = f.expr;
    expr.style.display = NO_EXPR.includes(f.transform) ? "none" : "";
    tf.addEventListener("change", () => {
      f.transform = tf.value; f.calcType = null;
      if (f.transform === "model" && !f.nested) f.nested = defaultNested();
      if (f.transform === "first_of" && !f.alts) f.alts = [newField(), newField()];
      renderBuilder(); builderChanged();
    });
    expr.addEventListener("input", () => { f.expr = expr.value; builderChanged(); });
    xtra.append(tf, expr);
    row.appendChild(xtra);

    if (!isFirstOf) {
      const condRow = document.createElement("div");
      condRow.className = "xtra";
      const cond = document.createElement("input");
      cond.type = "text";
      cond.placeholder = "condition — omit this field when false, e.g. fRes > 0 (optional)";
      cond.value = f.condition || "";
      cond.addEventListener("input", () => { f.condition = cond.value; builderChanged(); });
      condRow.appendChild(cond);
      row.appendChild(condRow);
    }

    if (f.transform === "model" && f.nested) {
      row.appendChild(nestedPanel(f.nested));
    }
    if (isFirstOf) {
      const panel = document.createElement("div");
      panel.className = "subpanel";
      const hint = document.createElement("div");
      hint.className = "hint";
      hint.textContent = "alternatives are tried in order — the first non-null wins";
      const altsWrap = document.createElement("div");
      altsWrap.style.cssText = "display:flex;flex-direction:column;gap:8px;";
      renderFieldList(altsWrap, f.alts || (f.alts = [newField()]), false, true);
      const add = document.createElement("button");
      add.type = "button"; add.className = "ghost"; add.textContent = "+ add alternative";
      add.addEventListener("click", () => { f.alts.push(newField()); renderBuilder(); builderChanged(); });
      panel.append(hint, altsWrap, add);
      row.appendChild(panel);
    }
    return row;
  }

  function slabel(text) {
    const s = document.createElement("span");
    s.className = "slabel";
    s.textContent = text;
    return s;
  }

  function nestedPanel(n) {
    const panel = document.createElement("div");
    panel.className = "subpanel";

    const r1 = document.createElement("div");
    r1.className = "srow";
    const method = document.createElement("select");
    for (const m of ["GET", "POST", "PUT", "DELETE"]) { const o = document.createElement("option"); o.textContent = m; method.appendChild(o); }
    method.value = n.method;
    method.addEventListener("change", () => { n.method = method.value; builderChanged(); });
    const url = document.createElement("input");
    url.type = "text";
    url.placeholder = "https://… ({PL} or {{{FromExp=…}}} injects the parent value)";
    url.value = n.url;
    url.addEventListener("input", () => { n.url = url.value; builderChanged(); });
    r1.append(method, url);

    const r2 = document.createElement("div");
    r2.className = "srow";
    const rt = document.createElement("select");
    for (const t of RESPONSE_TYPES) { const o = document.createElement("option"); o.textContent = t; rt.appendChild(o); }
    rt.value = n.responseType;
    rt.addEventListener("change", () => { n.responseType = rt.value; builderChanged(); });
    const nullLbl = document.createElement("label");
    nullLbl.className = "slabel";
    const nullCb = document.createElement("input");
    nullCb.type = "checkbox";
    nullCb.checked = n.nullOnError;
    nullCb.addEventListener("change", () => { n.nullOnError = nullCb.checked; builderChanged(); });
    nullLbl.append(nullCb, document.createTextNode(" null on error"));
    r2.append(slabel("parse as"), rt, nullLbl);

    const rHdrs = document.createElement("div");
    rHdrs.style.cssText = "display:flex;flex-direction:column;gap:6px;";
    for (const h of (n.headers || (n.headers = []))) {
      const hr = document.createElement("div");
      hr.className = "srow";
      const hk = document.createElement("input");
      hk.type = "text"; hk.placeholder = "header, e.g. Content-Type"; hk.value = h.k;
      hk.addEventListener("input", () => { h.k = hk.value; builderChanged(); });
      const hv = document.createElement("input");
      hv.type = "text"; hv.placeholder = "value"; hv.value = h.v;
      hv.addEventListener("input", () => { h.v = hv.value; builderChanged(); });
      const hd = document.createElement("button");
      hd.type = "button"; hd.className = "del"; hd.textContent = "×"; hd.title = "remove header";
      hd.addEventListener("click", () => {
        const i = n.headers.findIndex(x => x.id === h.id);
        if (i >= 0) n.headers.splice(i, 1);
        renderBuilder(); builderChanged();
      });
      hr.append(hk, hv, hd);
      rHdrs.appendChild(hr);
    }
    const addHdr = document.createElement("button");
    addHdr.type = "button"; addHdr.className = "ghost"; addHdr.textContent = "+ header";
    addHdr.addEventListener("click", () => {
      n.headers.push({ id: ++fieldSeq, k: "", v: "" });
      renderBuilder(); builderChanged();
    });
    rHdrs.appendChild(addHdr);

    const rBody = document.createElement("textarea");
    rBody.spellcheck = false;
    rBody.placeholder = "request body (optional — placeholders like {{{FromInput=…}}}, {{{RefName=…}}}, {{{FromExp=…}}} work here)";
    rBody.value = n.body || "";
    rBody.style.cssText = "border:1px solid var(--border);border-radius:7px;background:var(--bg);min-height:120px;padding:8px 10px;resize:vertical;";
    rBody.addEventListener("input", () => { n.body = rBody.value; builderChanged(); });

    const rTime = document.createElement("div");
    rTime.className = "srow";
    const tIn = document.createElement("input");
    tIn.type = "number"; tIn.min = "0"; tIn.placeholder = "default"; tIn.value = n.timeout ?? "";
    tIn.style.width = "90px";
    tIn.addEventListener("input", () => { n.timeout = tIn.value; builderChanged(); });
    rTime.append(slabel("timeout s"), tIn);

    const r3 = document.createElement("div");
    r3.className = "srow";
    const kind = document.createElement("select");
    for (const [v, label] of [["object", "single object"], ["array", "array of items"], ["value", "single value"]]) {
      const o = document.createElement("option"); o.value = v; o.textContent = label; kind.appendChild(o);
    }
    kind.value = n.kind;
    kind.addEventListener("change", () => { n.kind = kind.value; renderBuilder(); builderChanged(); });
    const shape = document.createElement("select");
    for (const [v, label] of [["fields", "items: named fields"], ["single", "items: single value"]]) {
      const o = document.createElement("option"); o.value = v; o.textContent = label; shape.appendChild(o);
    }
    shape.value = n.itemShape;
    shape.hidden = n.kind !== "array";
    shape.addEventListener("change", () => { n.itemShape = shape.value; renderBuilder(); builderChanged(); });
    const rootPath = document.createElement("input");
    rootPath.type = "text";
    rootPath.placeholder = "root path of the array";
    rootPath.value = n.rootPath;
    rootPath.hidden = n.kind !== "array";
    rootPath.addEventListener("input", () => { n.rootPath = rootPath.value; builderChanged(); });
    const itemCond = document.createElement("input");
    itemCond.type = "text";
    itemCond.placeholder = "item condition — drop items when false (optional)";
    itemCond.value = n.itemCondition || "";
    itemCond.hidden = n.kind !== "array";
    itemCond.addEventListener("input", () => { n.itemCondition = itemCond.value; builderChanged(); });
    r3.append(kind, shape, rootPath, itemCond);

    const nSingle = (n.kind === "array" && n.itemShape === "single") || n.kind === "value";
    if (nSingle && n.fields.length === 0) n.fields.push(newField());
    const fieldsWrap = document.createElement("div");
    fieldsWrap.style.cssText = "display:flex;flex-direction:column;gap:8px;";
    renderFieldList(fieldsWrap, n.fields, nSingle);

    const add = document.createElement("button");
    add.type = "button"; add.className = "ghost"; add.textContent = "+ add field";
    add.style.display = nSingle ? "none" : "";
    add.addEventListener("click", () => { n.fields.push(newField()); renderBuilder(); builderChanged(); });

    panel.append(r1, r2, rHdrs, rBody, rTime, r3, fieldsWrap, add);
    return panel;
  }

  let autoRunTimer = null;
  function builderChanged() {
    setBuilderRepresentable(true);
    configEl.value = JSON.stringify(buildConfig(bState), null, 2);
    updateSendCounterVis();
    updateNativeOnlyState();
    // instant feedback for offline sources; URL sources run via the button
    if (bState.source !== "url" && !runBtn.disabled) {
      clearTimeout(autoRunTimer);
      autoRunTimer = setTimeout(run, 500);
    }
  }

  function syncBuilderFromJson() {
    let parsed = null;
    try { parsed = JSON.parse(configEl.value); } catch { /* YAML or invalid: builder can't show it */ }
    const imported = parsed && importConfig(parsed);
    if (imported) {
      bState = imported;
      setBuilderRepresentable(true);
    } else {
      setBuilderRepresentable(false);
    }
    renderBuilder();
  }

  /* ---------- builder form events ---------- */
  $("bSource").addEventListener("change", e => { bState.source = e.target.value; renderBuilder(); builderChanged(); });
  $("bResponseType").addEventListener("change", e => { bState.responseType = e.target.value; builderChanged(); });
  $("bMethod").addEventListener("change", e => { bState.method = e.target.value; builderChanged(); });
  $("bUrl").addEventListener("input", e => { bState.url = e.target.value; builderChanged(); });
  $("bStatic").addEventListener("input", e => { bState.staticValue = e.target.value; builderChanged(); });
  $("bSeqStart").addEventListener("input", e => { bState.seqStart = e.target.value; builderChanged(); });
  $("bSeqEnd").addEventListener("input", e => { bState.seqEnd = e.target.value; builderChanged(); });
  $("bSeqStep").addEventListener("input", e => { bState.seqStep = e.target.value; builderChanged(); });
  $("bKind").addEventListener("change", e => { bState.kind = e.target.value; renderBuilder(); builderChanged(); });
  $("bItemShape").addEventListener("change", e => { bState.itemShape = e.target.value; renderBuilder(); builderChanged(); });
  $("bRootPath").addEventListener("input", e => { bState.rootPath = e.target.value; builderChanged(); });
  $("bItemCondition").addEventListener("input", e => { bState.itemCondition = e.target.value; builderChanged(); });
  $("bAddField").addEventListener("click", () => { bState.fields.push(newField()); renderBuilder(); builderChanged(); });
  $("bFresh").addEventListener("click", () => { bState = defaultState(); renderBuilder(); builderChanged(); });
  $("bRefName").addEventListener("input", e => { bState.refName = e.target.value; builderChanged(); });
  $("bAddRef").addEventListener("click", () => {
    bState.references.push({ id: ++fieldSeq, name: "", expire: "", ...defaultNested() });
    renderBuilder(); builderChanged();
  });
  $("bAddHostLimit").addEventListener("click", () => {
    bState.hostLimits.push({ id: ++fieldSeq, host: "", n: 2 });
    renderBuilder(); builderChanged();
  });

  /* ---------- tabs ---------- */
  function showTab(which) {
    $("tabBuilder").classList.toggle("active", which === "builder");
    $("tabJson").classList.toggle("active", which === "json");
    $("builderPane").classList.toggle("active", which === "builder");
    configEl.classList.toggle("hiddenTab", which === "builder");
    if (which === "builder") syncBuilderFromJson();
  }
  $("tabBuilder").addEventListener("click", () => showTab("builder"));
  $("tabJson").addEventListener("click", () => showTab("json"));

  /* ---------- per-example counters (abacus: /get reads, /hit increments) ---------- */
  const counterCache = {};
  function activeCounter() {
    // telegram is content-based (works for shared/edited configs too); others follow the selected example
    if (configEl.value.includes("api.telegram.org")) {
      return { key: "telegram-sends", mode: "delivered", label: "\u{1F4E8} sent via playground " };
    }
    const ex = examplesById.get(select.value);
    if (ex && ex.counter) return { key: ex.counter, mode: "run", label: "▶ example runs " };
    return null;
  }
  async function counterFetch(key, hit) {
    const r = await fetch("https://abacus.jasoncameron.dev/" + (hit ? "hit" : "get") + "/pxyup-fitter/" + key);
    const { value } = await r.json();
    if (typeof value === "number") counterCache[key] = value;
  }
  function renderCounter() {
    // defensive: these async fetches can resolve after the user has
    // navigated away from the playground, at which point this view's DOM
    // has been replaced and these ids no longer exist
    const el = $("sendCounter"), labelEl = $("counterLabel"), countEl = $("sendCount");
    if (!el || !labelEl || !countEl) return;
    const c = activeCounter();
    if (!c || counterCache[c.key] == null) { el.hidden = true; return; }
    labelEl.textContent = c.label;
    countEl.textContent = counterCache[c.key];
    el.hidden = false;
  }
  async function updateSendCounterVis() {
    const c = activeCounter();
    if (c && counterCache[c.key] == null) {
      try { await counterFetch(c.key, false); } catch { /* cosmetic */ }
    }
    renderCounter();
  }
  async function bumpCounter(parsed) {
    const c = activeCounter();
    if (!c) return;
    if (c.mode === "delivered" && !(parsed && parsed.delivered === true)) return;
    try { await counterFetch(c.key, true); } catch { /* cosmetic */ }
    renderCounter();
  }
  updateSendCounterVis();
  configEl.addEventListener("input", updateSendCounterVis);
  configEl.addEventListener("input", updateNativeOnlyState);

  async function run() {
    if (nativeOnly) return; // defense in depth — the button is already disabled for these
    runBtn.disabled = true;
    output.className = "";
    output.dataset.painted = "1";
    output.textContent = "running…";
    timing.textContent = "";
    const started = performance.now();
    try {
      const res = await runFitter(configEl.value, inputEl.value);
      const ms = Math.round(performance.now() - started);
      let parsed = null;
      try {
        parsed = JSON.parse(res);
        output.textContent = JSON.stringify(parsed, null, 2);
      } catch {
        output.textContent = res;
      }
      bumpCounter(parsed);
      timing.textContent = ms + " ms";
    } catch (e) {
      output.className = "error";
      output.textContent = "Error: " + e;
    } finally {
      syncRunEnabled();
    }
  }
  runBtn.addEventListener("click", run);
  const keyHandler = e => {
    if ((e.metaKey || e.ctrlKey) && e.key === "Enter" && !runBtn.disabled) run();
  };
  document.addEventListener("keydown", keyHandler);

  const shareBtn = $("share");
  shareBtn.addEventListener("click", async () => {
    const token = await encodeShare({ c: configEl.value, i: inputEl.value });
    const url = location.origin + location.pathname + "?q=" + token + "#/playground";
    history.replaceState(null, "", url);
    const old = shareBtn.textContent;
    try {
      await navigator.clipboard.writeText(url);
      shareBtn.textContent = "copied ✓";
    } catch {
      shareBtn.textContent = "link in address bar";
    }
    setTimeout(() => { shareBtn.textContent = old; }, 1600);
  });

  /* ---------- init: decide what to load, before wiring the engine status
     listener below (so it's already set if the engine is ready synchronously) ---------- */
  let runOnReady = false;
  const pending = takePendingExample();
  const q = new URLSearchParams(location.search).get("q");

  if (pending && examplesById.has(pending.id)) {
    select.value = pending.id;
    loadExample(pending.id);
    runOnReady = !pending.nativeOnly; // native-only: load it so it's visible/copyable, but never auto-run
  } else if (q) {
    try {
      const p = await decodeShare(q);
      configEl.value = typeof p.c === "string" ? p.c : "";
      setInput(typeof p.i === "string" ? p.i : "");
      markPseudoOption("— shared link —");
      syncBuilderFromJson();
      updateSendCounterVis();
      updateNativeOnlyState();
      runOnReady = !isNativeOnlyConfig(configEl.value);
    } catch (e) {
      console.warn("ignoring invalid ?q= share link:", e);
      clearPseudoOption();
      loadExample(select.value = runnableExamples[0].id);
    }
  } else {
    loadExample(select.value = runnableExamples[0].id);
  }

  if (params && params.builder) showTab("builder");

  /* ---------- engine ---------- */
  const unsubscribe = onStatusChange((status, detail) => {
    if (status === "loading" || status === "idle") {
      statusLocal.textContent = "loading engine…";
      statusLocal.className = "engine-status-local";
      engineReady = false;
      syncRunEnabled();
    } else if (status === "ready") {
      statusLocal.textContent = "engine ready";
      statusLocal.className = "engine-status-local ready";
      engineReady = true;
      syncRunEnabled();
      if (runOnReady && !nativeOnly) {
        runOnReady = false;
        run();
      } else if (!output.dataset.painted) {
        output.textContent = nativeOnly
          ? "This config can't run in the WebAssembly sandbox — see the note above Run. Copy it and use it with the native fitter_cli or MCP server."
          : "Pick an example, build a config on the Builder tab, or write JSON/YAML — then hit Run.";
      }
    } else if (status === "error") {
      statusLocal.textContent = "engine failed to load";
      statusLocal.className = "engine-status-local error";
      output.className = "error";
      output.textContent = String(detail);
    }
  });
  ensureEngine().catch(() => { /* surfaced via onStatusChange */ });

  // cleanup when the view is torn down by the next render() call
  return () => {
    unsubscribe();
    document.removeEventListener("keydown", keyHandler);
  };
}
