// md.js — a tiny, dependency-free "markdown-lite" renderer for our own
// hand-authored docs.json content (not for arbitrary/user input). Escapes
// HTML as it goes (per-line, not up front — code fences are escaped/
// highlighted separately so JSON syntax highlighting doesn't get
// double-escaped), and layers on a handful of block/inline conventions:
// fenced ```code``` blocks (with JSON syntax highlighting for ```json),
// "## " / "### " headings, "- " lists, "> " quotes, blank-line paragraphs,
// and inline `code`, **bold**, [text](url).
import { highlightJson } from "./jsonhl.js";

function escapeHtml(s) {
  return String(s)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");
}

function inline(escapedText) {
  // escapedText is already HTML-escaped
  let text = escapedText.replace(/`([^`]+)`/g, (_, code) => `<code>${code}</code>`);
  text = text.replace(/\*\*([^*]+)\*\*/g, (_, b) => `<strong>${b}</strong>`);
  text = text.replace(/\[([^\]]+)\]\(([^)]+)\)/g, (_, t, href) => {
    const external = /^https?:\/\//.test(href);
    const attrs = external ? ' target="_blank" rel="noopener"' : "";
    return `<a href="${href}"${attrs}>${t}</a>`;
  });
  return text;
}

export function renderMarkdownLite(raw) {
  const lines = String(raw).replace(/\r\n/g, "\n").split("\n");
  const out = [];
  let i = 0;
  let para = [];
  let list = null;

  function flushPara() {
    if (para.length) {
      out.push(`<p>${inline(para.join(" "))}</p>`);
      para = [];
    }
  }
  function flushList() {
    if (list) {
      out.push(`<ul>${list.map(li => `<li>${inline(li)}</li>`).join("")}</ul>`);
      list = null;
    }
  }

  while (i < lines.length) {
    const line = lines[i];

    if (/^```/.test(line)) {
      flushPara(); flushList();
      const lang = line.slice(3).trim();
      const codeLines = [];
      i++;
      while (i < lines.length && !/^```/.test(lines[i])) {
        codeLines.push(lines[i]);
        i++;
      }
      i++; // skip closing fence
      const cls = lang ? ` class="lang-${lang}"` : "";
      const rawCode = codeLines.join("\n");
      const body = lang === "json" ? highlightJson(rawCode) : escapeHtml(rawCode);
      out.push(`<pre><code${cls}>${body}</code></pre>`);
      continue;
    }

    if (/^###\s+/.test(line)) {
      flushPara(); flushList();
      out.push(`<h4>${inline(escapeHtml(line.replace(/^###\s+/, "")))}</h4>`);
      i++; continue;
    }
    if (/^##\s+/.test(line)) {
      flushPara(); flushList();
      out.push(`<h3>${inline(escapeHtml(line.replace(/^##\s+/, "")))}</h3>`);
      i++; continue;
    }
    if (/^>\s?/.test(line)) {
      flushPara(); flushList();
      const quoteLines = [];
      while (i < lines.length && /^>\s?/.test(lines[i])) {
        quoteLines.push(escapeHtml(lines[i].replace(/^>\s?/, "")));
        i++;
      }
      out.push(`<blockquote>${inline(quoteLines.join(" "))}</blockquote>`);
      continue;
    }
    if (/^-\s+/.test(line)) {
      flushPara();
      if (!list) list = [];
      list.push(escapeHtml(line.replace(/^-\s+/, "")));
      i++; continue;
    }
    if (line.trim() === "") {
      flushPara(); flushList();
      i++; continue;
    }

    flushList();
    para.push(escapeHtml(line.trim()));
    i++;
  }
  flushPara(); flushList();
  return out.join("\n");
}
