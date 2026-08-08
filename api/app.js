// The lookup page, and its router.
//
// GitHub Pages serves files and runs nothing, so a pretty path is a trick with
// two halves: 404.html is what Pages returns for an unrouted path, and it loads
// this module, which reads the path it was asked for and answers it. The URL
// stays as typed, so /d/SM-S928B is shareable, linkable and back-buttonable
// even though no file of that name exists.
//
// Three addresses resolve the same code, because all three already exist in the
// wild — a shared link, a copied fragment, a form submission:
//
//	/d/SM-S928B
//	/?code=SM-S928B
//	/#SM-S928B
//
// The first is canonical; the page rewrites the other two into it.

import { Devicex, parseCode } from './devicex.js';

// The site root, derived from this module's own URL rather than assumed. A
// project page lives under /devicex/ and a user page at /, and hard-coding
// either one breaks the other.
const API = new URL('.', import.meta.url);
const ROOT = new URL('..', API);

const dx = new Devicex(API.href);

const el = (id) => document.getElementById(id);

// codeFromLocation reads the requested code out of whichever address form was
// used. Everything after /d/ is the code, undecoded slashes included: a few
// real model codes contain characters that a stricter parse would drop.
function codeFromLocation() {
  const path = decodeURIComponent(location.pathname);
  const root = decodeURIComponent(ROOT.pathname);
  const rest = path.startsWith(root) ? path.slice(root.length) : path.replace(/^\//, '');

  const m = rest.match(/^d\/(.+)$/);
  if (m) return m[1];

  const q = new URLSearchParams(location.search).get('code');
  if (q) return q;

  if (location.hash.length > 1) return decodeURIComponent(location.hash.slice(1));
  return '';
}

// canonical is the pretty URL for a code. An empty code is the bare page.
function canonical(code) {
  return code ? ROOT.pathname + 'd/' + encodeURIComponent(code) : ROOT.pathname;
}

async function render(code) {
  const verdict = el('verdict');
  const out = el('out');
  const perma = el('perma');

  code = parseCode(code || '');
  document.title = code ? `${code} — devicex` : 'devicex — model code lookup';

  if (!code) {
    verdict.textContent = '';
    verdict.className = 'verdict';
    out.textContent = '{}';
    perma.textContent = '';
    return;
  }

  const r = await dx.resolve(code, code);
  const json = r ? { code, ok: true, ...r } : { code, ok: false };
  out.textContent = JSON.stringify(json, null, 2);

  if (!r) {
    // The whole point of the package: an unrecognised code stays
    // unrecognised. No nearest match, no most-common manufacturer.
    verdict.textContent = 'Not recognised — no catalogue record, and no rule matches this shape.';
    verdict.className = 'verdict miss';
  } else if (r.name) {
    verdict.textContent = `${r.name} — ${r.brand}`;
    verdict.className = 'verdict hit';
  } else {
    verdict.textContent = `${r.brand} — the code space identifies the maker, not the handset.`;
    verdict.className = 'verdict hit';
  }

  const shard = new URL('./' + shardName(code), API);
  perma.innerHTML = `Answered from <a href="${shard.pathname}"><code>${shard.pathname}</code></a>.`;
}

function shardName(code) {
  let out = '';
  for (const ch of [...code].slice(0, 2)) {
    const up = ch.toUpperCase();
    out += /[A-Z0-9]/.test(up) ? up : '_';
  }
  return out + '.json';
}

// go navigates to a code, pushing history only when the code actually changed,
// so typing does not fill the back stack with every keystroke.
function go(code, { replace = false } = {}) {
  const url = canonical(code);
  if (url !== location.pathname + location.search + location.hash) {
    history[replace ? 'replaceState' : 'pushState']({ code }, '', url);
  }
  render(code);
}

export function start() {
  const q = el('q');
  const initial = codeFromLocation();
  q.value = initial;

  // Whatever address form brought the visitor here, the URL they keep is the
  // canonical one.
  go(initial, { replace: true });

  let timer;
  q.addEventListener('input', () => {
    clearTimeout(timer);
    timer = setTimeout(() => go(q.value.trim(), { replace: true }), 150);
  });

  q.form?.addEventListener('submit', (e) => {
    e.preventDefault();
    clearTimeout(timer);
    go(q.value.trim());
  });

  addEventListener('popstate', () => {
    const code = codeFromLocation();
    q.value = code;
    render(code);
  });

  dx.provenance()
    .then((m) => {
      el('prov').textContent =
        `${m.devices.toLocaleString()} devices, snapshot ${m.generated}. ${m.source}`;
    })
    .catch(() => {});
}
