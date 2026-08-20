// pending.js — tiny in-memory handoff for "open this example in the
// playground and run it". The SPA never fully reloads on hash navigation,
// so a module-level variable is enough (no need for sessionStorage/URL state).

let pendingExample = null;

export function setPendingExample(example) {
  pendingExample = example;
}

export function takePendingExample() {
  const e = pendingExample;
  pendingExample = null;
  return e;
}
