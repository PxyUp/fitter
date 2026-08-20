// view-docs.js — docs list/detail views with client-side full-text search
// over the prebuilt demo/docs.json index. No external services.
import { renderMarkdownLite } from "./md.js";

let docsPromise = null;
function loadDocs() {
  if (!docsPromise) docsPromise = fetch("docs.json").then(r => r.json());
  return docsPromise;
}

function buildSearchIndex(docs) {
  return docs.map(d => ({
    doc: d,
    title: d.title.toLowerCase(),
    summary: d.summary.toLowerCase(),
    haystack: (d.title + " " + d.category + " " + d.summary + " " + d.body).toLowerCase()
  }));
}

function searchDocs(index, query) {
  const q = query.trim().toLowerCase();
  if (!q) return index.map(e => e.doc);
  const terms = q.split(/\s+/).filter(Boolean);
  const scored = [];
  for (const entry of index) {
    let score = 0;
    for (const t of terms) {
      if (entry.title.includes(t)) score += 5;
      if (entry.summary.includes(t)) score += 3;
      if (entry.haystack.includes(t)) score += 1;
    }
    if (score > 0) scored.push({ doc: entry.doc, score });
  }
  scored.sort((a, b) => b.score - a.score);
  return scored.map(s => s.doc);
}

function groupByCategory(docs) {
  const groups = new Map();
  for (const d of docs) {
    if (!groups.has(d.category)) groups.set(d.category, []);
    groups.get(d.category).push(d);
  }
  return groups;
}

function sidebarHtml(docs, activeSlug, query) {
  const groups = groupByCategory(docs);
  let html = `<input type="search" id="docsSearch" class="docs-search" placeholder="Search docs…" value="${query ? escapeAttr(query) : ""}" autocomplete="off">`;
  html += `<div id="docsSearchCount" class="docs-search-count"></div>`;
  html += `<nav class="docs-nav">`;
  for (const [category, items] of groups) {
    html += `<div class="docs-nav-group"><div class="docs-nav-cat">${category}</div><ul>`;
    for (const d of items) {
      html += `<li><a href="#/docs/${d.slug}" class="${d.slug === activeSlug ? "active" : ""}">${d.title}</a></li>`;
    }
    html += `</ul></div>`;
  }
  html += `</nav>`;
  return html;
}

function escapeAttr(s) {
  return String(s).replace(/&/g, "&amp;").replace(/"/g, "&quot;");
}

export async function renderDocs(app, params) {
  app.innerHTML = `<div class="docs-layout"><aside class="docs-sidebar" id="docsSidebar">loading…</aside><article class="docs-content" id="docsContent">loading…</article></div>`;
  const docs = await loadDocs();
  const index = buildSearchIndex(docs);
  const slug = params.slug || "overview";
  const doc = docs.find(d => d.slug === slug) || docs[0];

  function paint(query) {
    const sidebar = document.getElementById("docsSidebar");
    const content = document.getElementById("docsContent");
    if (!sidebar || !content) return;

    if (query && query.trim()) {
      const results = searchDocs(index, query);
      sidebar.innerHTML = `<input type="search" id="docsSearch" class="docs-search" placeholder="Search docs…" value="${escapeAttr(query)}" autocomplete="off">` +
        `<div id="docsSearchCount" class="docs-search-count">${results.length} result${results.length === 1 ? "" : "s"}</div>` +
        `<nav class="docs-nav"><ul>${results.map(d => `<li><a href="#/docs/${d.slug}" class="${d.slug === doc.slug ? "active" : ""}">${d.title} <span class="docs-nav-cat-inline">${d.category}</span></a></li>`).join("") || `<li class="docs-no-results">No matches</li>`}</ul></nav>`;
    } else {
      sidebar.innerHTML = sidebarHtml(docs, doc.slug, "");
    }
    const input = document.getElementById("docsSearch");
    input.addEventListener("input", () => paint(input.value));
    // keep focus + cursor position across re-paints while typing
    if (document.activeElement !== input && query !== undefined && query !== null && query !== "") {
      input.focus();
      input.setSelectionRange(input.value.length, input.value.length);
    }

    const idx = docs.findIndex(d => d.slug === doc.slug);
    const prev = docs[idx - 1];
    const next = docs[idx + 1];
    content.innerHTML = `
      <div class="docs-breadcrumb">${doc.category}</div>
      <h1>${doc.title}</h1>
      <p class="docs-summary">${doc.summary}</p>
      <div class="docs-body">${renderMarkdownLite(doc.body)}</div>
      <div class="docs-pager">
        ${prev ? `<a class="docs-pager-link" href="#/docs/${prev.slug}">← ${prev.title}</a>` : "<span></span>"}
        ${next ? `<a class="docs-pager-link docs-pager-next" href="#/docs/${next.slug}">${next.title} →</a>` : ""}
      </div>
    `;
  }

  paint("");
}
