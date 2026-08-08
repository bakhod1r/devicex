---
title: Changelog
---

# Changelog

The catalogue and the code version together, because a lookup result is only
reproducible if you know which catalogue produced it. `devicex.Source` records
where the shipped data came from.

---

## v0.1.0 — first release as devicex

The module path is `github.com/bakhod1r/devicex`, which is a new module: a path
is part of a module's identity, so its versions start again here rather than
continuing the `adx` series. The three releases below this one were published
as `github.com/bakhod1r/adx` and are kept for the history of the code, not as
versions of this module.

Everything `adx` v0.2.0 exported is here under the new name, with the changes
that follow.

`Describe` returns a `model`, which the same function in `adx` v0.2.0 did not:

```go
name, brand, model, family, deviceType, confidence, ok := devicex.Describe(code, ua)
//                 ^^^^^ third, and absent from adx
```

The flat signature dropped the one field a consumer cannot rebuild. `iPhone` is
both a name and the model an iPhone request carries; `Macintosh` names a class
and no model at all. A consumer that copies name into model whenever the
User-Agent contains it invents a `Mac` model for every desktop, and there is no
heuristic that separates the two — "Macintosh" contains "Mac". `Resolve` always
carried `Rule.Model`; now `Describe` does too.

Catalogue, rebuilt from a source that is still maintained:

- **25475 devices** (`adx` shipped 12260), snapshot **2026-08-08**. The previous import
  came from pbakondy/android-device-list, which stopped updating on 2023-04-04
  and so ended at the Galaxy S23: `SM-S928B`, `SM-X710`, `V2312A` and every
  other 2024-onward code resolved a brand from its shape and no name at all.
  The catalogue now comes from Google Play's published supported-devices list,
  which is refreshed continuously and needs no Play Console account, and is
  supplemented by the English pages of KHwang9883/MobileModels for the regional
  variants the Play list omits. Rows naming a device in CJK script are dropped
  wherever they come from, and a test enforces it: a catalogue that answers in
  two scripts cannot be aggregated, sorted or displayed by one caller. The Play list decides where
  the two disagree. **MobileModels is CC BY-NC-SA 4.0, and the shipped
  catalogue therefore is too** — see NOTICE.md before redistributing this in a
  commercial product.
- `Generated` and `Version()` report which snapshot answered —
  `"2026-08-08/25475"`. `Source` alone did not identify the data: the same
  source yields a different catalogue every import, and "what was this name
  decided from six months ago" had no answer.
- `Oppo` and `Vivo` in the Play export normalise to `OPPO` and `vivo`, the
  spellings the rest of the package uses, so a caller aggregating by brand does
  not see two manufacturers per company.

Static lookup API:

- `api/d/<code>.json` — one document per model code, 25287 of them, each the
  whole answer: `{"code","ok","id","name","brand","model","confidence"}`, plus
  `family` and `type` where a rule identifies them. 137 bytes gzipped, no client
  code, no page. A code with no document is a 404, which is the answer: not in
  the catalogue. Codes differing from another only in case have no document —
  a macOS or Windows checkout cannot hold both files — and a `/` in a code is
  percent-encoded into the filename.
- `lookup.html`, `404.html` and `api/` — the catalogue as sharded JSON on
  GitHub Pages, and an ES module that reads it. `/d/SM-S928B` resolves in the
  browser: Pages serves `404.html` for the unpublished path, and the router
  answers the path it was asked for, so the URL stays shareable. `?code=` and
  `#` forms are accepted and rewritten. The JSON files answer `curl` and
  `fetch` just as well, which is what makes them the API rather than a page
  feature. One shard is the codes sharing a two-character prefix, so a lookup
  transfers kilobytes, not the catalogue.
  The page fetches the shard and the rules before its own module is parsed,
  and caches both for the session, so a cold lookup is two round trips and a
  repeat is none.
- Market and SIM qualifiers are trimmed off MobileModels names: 1717 of its
  rows named the same handset "… Global", "… India", "… Dual SIM". The code
  already distinguishes those variants, and it is the key.
- `gen -web DIR` writes the shards, `rules.json` and `meta.json`. The rule table
  ships with the data so the browser answers an uncatalogued code exactly as the
  Go package does — the manufacturer from the code's shape, and no name.

Importer:

- `gen -md DIR` reads the MobileModels pages; `-csv` and `-md` combine, with
  the CSV deciding conflicts. `gen -json` is gone — the list it read stopped
  updating in 2023, and keeping a reader for data nobody should import invites
  someone to import it.

Internal, no API change:

- `BrandOf` reads `Rules` through `ResolveCode` instead of a second prefix
  table of its own. The two copies had already drifted — the `SM-T`/`SM-X`/
  `SM-P` tablet rules added in `adx` v0.2.0 went into `Rules` only, so a caller asking
  through `BrandOf` could not see a form factor a caller asking through
  `ResolveCode` could. One table cannot disagree with itself.
- New rule `playstation`: a PlayStation whose generation is not one of the two
  named ones still resolves Sony and `TypeConsole`, and asserts no model. It
  sits below `playstation_5` and `playstation_4`, which keep answering first.

---

# Previously, as `github.com/bakhod1r/adx`

## adx v0.3.0 — renamed to devicex

Module path and package name are now `github.com/bakhod1r/devicex` and
`devicex`. Nothing else changed: every function, type and constant keeps its
name and behaviour.

```go
import "github.com/bakhod1r/devicex"   // was github.com/bakhod1r/adx

devicex.Lookup("SM-G973F")             // was adx.Lookup
```

A module path is part of a module's identity, so this is a new module rather
than a new version of the old one. `github.com/bakhod1r/adx` stays resolvable
at v0.2.0 and receives nothing further.

## adx v0.2.0 — the rules became callable

v0.1.0 exported `Rules` as data and no way to evaluate it. Every consumer that
wanted the Apple, console and code-shape rules had to re-implement the walk,
including the parts that are not obvious: rules are ordered by descending
priority, and within one priority a longer prefix is listed before the shorter
one it would otherwise shadow. That knowledge belongs with the data.

Added, all backwards compatible:

- `Resolve(code, ua)` — the catalogue, the shape rules and the User-Agent rules
  in one answer. The catalogue decides first, because a record names the
  specific handset and a rule never can; the rules then fill in the form factor
  and product line the catalogue does not carry. Returns the matching `Rule`,
  whose `ID` names what produced the answer.
- `ResolveCode(code)` and `ResolveUA(ua)` for the two kinds of evidence
  separately. `MatchContains` rules describe prose in a User-Agent
  (`PlayStation 5`, `Nintendo Switch`) and are never tested against a model
  code; prefix rules are never tested against a whole User-Agent.
- `Describe(code, ua)` — `Resolve` behind a signature naming no type from this
  package, so a consumer can accept it as a function value without importing
  devicex. `Names` only reached the catalogue, which left it silent about iPhones,
  consoles, and any handset newer than the catalogue.
- `Code(field)` strips the build fingerprint Android appends to the model
  field: `"SM-A546E Build/UP1A.231005.007"` → `"SM-A546E"`. The suffix is what
  turns a catalogue hit into a miss, and the shape of a model code is this
  package's subject.

Rules:

- Samsung tablet code spaces `SM-T`, `SM-X` and `SM-P` now assert
  `TypeTablet` and the Galaxy Tab family, above the plain `SM-` rule they would
  otherwise be shadowed by. The catalogue holds 331 of these and records no
  form factor — the Play Console export that carries one omits the `Model`
  column, so it cannot be imported. `KF` (Kindle Fire) likewise asserts
  `TypeTablet`.

`Rules` is unchanged in shape and remains exported. `Confidence` stays
`float64`; a caller wanting `float32` converts at its own boundary rather than
having this package pick the narrower type for everyone.

## adx v0.1.0 — first release

Extracted from [uax](https://github.com/bakhod1r/uax), where the same data
lived as a 4.6 MB YAML file compiled into the parser's rule tables. Separating
it lets the catalogue be updated on its own schedule and keeps the parser's
binary small for callers who do not need device names.

- `Lookup` resolves a model code to its brand and marketing name. Exact match,
  binary search over a sorted generated array: ~50 ns on an M4 Pro, zero
  allocations, nothing built at init.
- `LookupFold` for input whose case is unreliable.
- `BrandOf` resolves the manufacturer from the code's shape when the catalogue
  has no record, so handsets released after this build still resolve their
  maker.
- `All`, `Len`, `Brands`, `Source`.
- Importer (`gen/`) reads the Google Play Console supported-devices CSV
  (UTF-16LE) or the pbakondy JSON list. It refuses to run without a `-source`
  citation, and refuses an export whose `Model` column is missing rather than
  building a catalogue keyed on something a User-Agent never carries.
- Brand spellings are normalised on import, so one manufacturer is one brand.
  Xiaomi sub-brands (`Redmi`, `POCO`) report `Xiaomi`.
- Rows whose model column holds a word rather than a code (`Android`, `Linux`,
  `K`) are dropped; indexing them would match nearly every User-Agent.

**12,260 devices.** Coverage is good through 2023 and thin after it — the
upstream community list stopped keeping pace. `make catalog CSV=…` from a
current Google Play Console export replaces it.
