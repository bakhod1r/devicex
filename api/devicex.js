// devicex, in the browser.
//
// The same three-step answer the Go package gives: the catalogue first, then
// the code-shape rules, then nothing. A code the catalogue does not hold
// resolves a manufacturer from its shape and no name at all — never a guess,
// because a wrong device name is worse than none.
//
// The data is static JSON under this directory, so this file works from a
// file:// URL, a CDN, or anywhere else it is copied to.

const DEFAULT_BASE = new URL('.', import.meta.url).href;

// Mirrors shardKey in gen/web.go: the first two characters, uppercased,
// anything that is not a letter or digit replaced by an underscore. Change one,
// change both.
export function shardKey(code) {
  let out = '';
  for (const ch of [...code].slice(0, 2)) {
    const up = ch.toUpperCase();
    out += /[A-Z0-9]/.test(up) ? up : '_';
  }
  return out;
}

// Strips the build fingerprint Android appends to the model field:
// "SM-A546E Build/UP1A.231005.007" -> "SM-A546E". Mirrors devicex.Code.
export function parseCode(field) {
  field = field.trim();
  const i = field.indexOf(' Build/');
  if (i >= 0) field = field.slice(0, i);
  if (field.startsWith('Build/')) return '';
  return field.trim();
}

export class Devicex {
  constructor(base = DEFAULT_BASE) {
    this.base = base.endsWith('/') ? base : base + '/';
    this.shards = new Map();
    this.rules = null;
    this.meta = null;
  }

  // A page can start these fetches before this module is even parsed, by
  // leaving promises on window.__dxWarm — which is what the lookup pages do,
  // so the shard and the rule table travel alongside the code instead of
  // waiting for it. A warm entry is used once and then behaves normally.
  #warm(name) {
    const warm = globalThis.__dxWarm;
    if (!warm || !(name in warm)) return null;
    const p = warm[name];
    delete warm[name];
    return p;
  }

  async #json(name) {
    const cached = this.#cached(name);
    if (cached) return cached;

    const warm = this.#warm(name);
    if (warm) {
      const v = await warm;
      if (v === null) throw new Error(`${name}: not found`);
      this.#cache(name, v);
      return v;
    }

    const r = await fetch(this.base + name);
    if (!r.ok) throw new Error(`${name}: HTTP ${r.status}`);
    const v = await r.json();
    this.#cache(name, v);
    return v;
  }

  // Shards are cached for the tab's lifetime, not longer. GitHub Pages serves
  // these with a ten-minute cache, so the network is not the cost a second
  // lookup pays — but a session cache removes even that, and expires on its own
  // when the tab closes, which is the freshness rule a static catalogue can
  // actually honour: a redeploy is picked up by the next session.
  #cached(name) {
    try {
      const raw = sessionStorage.getItem('dx:' + this.base + name);
      return raw ? JSON.parse(raw) : null;
    } catch {
      return null;
    }
  }

  #cache(name, value) {
    try {
      sessionStorage.setItem('dx:' + this.base + name, JSON.stringify(value));
    } catch {
      // Full, disabled, or a private-mode quota of zero. The cache is an
      // optimisation; losing it costs a request and nothing else.
    }
  }

  // One shard is fetched per code space and kept, so a page resolving a list of
  // Samsung codes makes one request, not one per code. A missing shard is not
  // an error: it means no catalogued code starts with those characters.
  async #shard(key) {
    if (!this.shards.has(key)) {
      this.shards.set(key, this.#json(key + '.json').catch(() => ({})));
    }
    return this.shards.get(key);
  }

  async provenance() {
    this.meta ??= await this.#json('meta.json');
    return this.meta;
  }

  // lookup is the catalogue alone. Exact, case-sensitive: model codes are
  // copied verbatim out of a User-Agent, so a looser match would be a weaker
  // claim than the one this returns.
  async lookup(code) {
    if (!code) return null;
    const shard = await this.#shard(shardKey(code));
    const hit = shard[code];
    return hit ? { code, brand: hit[0], name: hit[1] } : null;
  }

  async #ruleTable() {
    this.rules ??= await this.#json('rules.json');
    return this.rules;
  }

  // resolveCode reads the shape of a code and nothing else, so it answers for
  // hardware the catalogue has never heard of. Contains rules are skipped: they
  // describe prose in a User-Agent, and a model code is not prose.
  async resolveCode(code) {
    if (!code) return null;
    for (const r of await this.#ruleTable()) {
      if (r.match === 'token' && r.value === code) return r;
      if (r.match === 'prefix' && code.startsWith(r.value)) return r;
    }
    return null;
  }

  // resolveUA is for hardware that carries no model code and names itself in
  // prose instead. Prefix rules are never tested against a whole User-Agent:
  // "SM-" appears in one only as part of a model code.
  async resolveUA(ua) {
    if (!ua) return null;
    for (const r of await this.#ruleTable()) {
      if ((r.match === 'contains' || r.match === 'token') && ua.includes(r.value)) return r;
    }
    return null;
  }

  // resolve is everything: the catalogue decides first, because a record names
  // the specific handset and a rule never can, and the rules then fill in the
  // form factor and product line the catalogue does not carry.
  async resolve(code, ua = '') {
    // Both requests are started before either is awaited. The catalogue decides
    // the answer, but the rules are needed either way — for the form factor
    // when it hits, and for the manufacturer when it misses — so fetching them
    // one after the other would spend a round trip on a certainty.
    const rules = this.#ruleTable();
    const hit = await this.lookup(code);
    await rules;
    if (hit) {
      const out = {
        id: 'catalog',
        name: hit.name,
        brand: hit.brand,
        model: hit.code,
        confidence: 0.99,
      };
      const shape = await this.resolveCode(code);
      if (shape) {
        if (shape.type) out.type = shape.type;
        if (shape.family) out.family = shape.family;
      }
      return out;
    }
    const shape = await this.resolveCode(code);
    if (shape) return shaped(shape);
    const prose = await this.resolveUA(ua);
    return prose ? shaped(prose) : null;
  }
}

function shaped(r) {
  const out = { id: r.id, brand: r.brand, confidence: r.confidence };
  if (r.name) out.name = r.name;
  if (r.model) out.model = r.model;
  if (r.family) out.family = r.family;
  if (r.type) out.type = r.type;
  return out;
}
