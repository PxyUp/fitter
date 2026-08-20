// jsonhl.js — a tiny, dependency-free JSON syntax highlighter. Tokenizes
// JSON-*looking* text into HTML spans classed by token kind (see the
// .hl-key/.hl-string/.hl-number/.hl-bool/.hl-null/.hl-punct rules in
// styles.css, themed via the --code-* tokens). Anything that isn't a
// recognized JSON token (plain text, whitespace, punctuation that doesn't
// look like JSON) passes through escaped and unstyled, so it's also safe to
// run over non-JSON strings (e.g. a plain-text example "input" preview).

function escapeHtml(s) {
  return String(s)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");
}

// Group 1: a quoted string, optionally followed by (whitespace +) a colon —
// which is only possible in valid JSON when the string is an object key.
// Group 2: a number. Group 3: true/false. Group 4: null. Group 5: structural
// punctuation. Matches are evaluated left-to-right so plain text in between
// (whitespace, commas already covered by group 5, etc.) is left untouched.
const TOKEN_RE = /("(?:\\.|[^"\\])*")(\s*:)?|(-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)|\b(true|false)\b|\b(null)\b|([{}[\],])/g;

export function highlightJson(text) {
  const src = String(text);
  let out = "";
  let last = 0;
  let m;
  TOKEN_RE.lastIndex = 0;
  while ((m = TOKEN_RE.exec(src)) !== null) {
    const [full, str, colon, num, bool, nul, punct] = m;
    out += escapeHtml(src.slice(last, m.index));
    if (str !== undefined) {
      out += `<span class="${colon ? "hl-key" : "hl-string"}">${escapeHtml(str)}</span>`;
      if (colon) out += escapeHtml(colon);
    } else if (num !== undefined) {
      out += `<span class="hl-number">${escapeHtml(num)}</span>`;
    } else if (bool !== undefined) {
      out += `<span class="hl-bool">${escapeHtml(bool)}</span>`;
    } else if (nul !== undefined) {
      out += `<span class="hl-null">${escapeHtml(nul)}</span>`;
    } else if (punct !== undefined) {
      out += `<span class="hl-punct">${escapeHtml(punct)}</span>`;
    } else {
      out += escapeHtml(full);
    }
    last = m.index + full.length;
  }
  out += escapeHtml(src.slice(last));
  return out;
}
