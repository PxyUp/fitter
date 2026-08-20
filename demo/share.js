// share.js — encode/decode the playground's ?q=<version>.<base64url> share
// token. version "1" = deflate-raw compressed, "0" = plain JSON (fallback
// for browsers without CompressionStream).

const txtEnc = new TextEncoder(), txtDec = new TextDecoder();

function b64urlEncode(bytes) {
  let s = "";
  for (const b of bytes) s += String.fromCharCode(b);
  return btoa(s).replaceAll("+", "-").replaceAll("/", "_").replace(/=+$/, "");
}
function b64urlDecode(str) {
  const s = str.replaceAll("-", "+").replaceAll("_", "/");
  const bin = atob(s + "=".repeat((4 - s.length % 4) % 4));
  return Uint8Array.from(bin, c => c.charCodeAt(0));
}
async function pipeBytes(bytes, stream) {
  return new Uint8Array(await new Response(new Blob([bytes]).stream().pipeThrough(stream)).arrayBuffer());
}

export async function encodeShare(payload) {
  const json = txtEnc.encode(JSON.stringify(payload));
  if ("CompressionStream" in window) {
    return "1." + b64urlEncode(await pipeBytes(json, new CompressionStream("deflate-raw")));
  }
  return "0." + b64urlEncode(json);
}

export async function decodeShare(token) {
  const dot = token.indexOf(".");
  if (dot < 0) throw new Error("bad share token");
  let bytes = b64urlDecode(token.slice(dot + 1));
  if (token.slice(0, dot) === "1") {
    bytes = await pipeBytes(bytes, new DecompressionStream("deflate-raw"));
  }
  return JSON.parse(txtDec.decode(bytes));
}

// __BASE__ resolves to the URL the demo itself is served from, so bundled
// assets (sample.pdf) fetch same-origin — locally and on GitHub Pages alike.
export const BASE = location.origin + location.pathname.replace(/\/[^/]*$/, "");
