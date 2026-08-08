---
title: devicex — the Go package
---

# devicex

Device model code → the device it names.

```go
d, ok := devicex.Lookup("SM-G973F")
// d.Brand = "Samsung", d.Name = "Galaxy S10"
```

25,475 devices, snapshot 2026-08-08. ~50 ns per lookup, zero allocations, zero dependencies.

---

## The problem this solves

A User-Agent from an Android phone tells you a model code and nothing else:

```
Mozilla/5.0 (Linux; Android 14; SM-G973F) AppleWebKit/537.36 …
                                 ^^^^^^^^
```

`SM-G973F` is a Galaxy S10. There is **no derivable relationship** between the
two — Samsung assigned it. No regex, no heuristic and no amount of cleverness
recovers the name. It requires a catalogue, which is why almost nothing in the
Go ecosystem does it.

This package is that catalogue and nothing else. It does not parse User-Agents;
it answers one question, quickly, and admits when it cannot.

## Install

```bash
go get github.com/bakhod1r/devicex
```

## Use

```go
// The device, when the catalogue knows it.
d, ok := devicex.Lookup("CPH2451")   // OnePlus 11 5G

// The manufacturer, which is answerable even when the handset is not.
brand, ok := devicex.BrandOf("SM-S928B")   // "Samsung" — code shape, not a record

// Everything, without copying megabytes.
devicex.All(func(d devicex.Device) bool {
    if d.Brand == "Sony" {
        fmt.Println(d.Code, d.Name)
    }
    return true
})

devicex.Len()      // 25475
devicex.Brands()   // every manufacturer, sorted
devicex.Source     // where the data came from
devicex.Version()  // "2026-08-08/25475" — the snapshot the answer came from
```

### Everything at once

`Resolve` answers from the catalogue, the code-shape rules and the User-Agent
rules together, in that order of authority:

```go
code := devicex.Code("SM-T870 Build/UP1A.231005.007")  // strips the build suffix

r, ok := devicex.Resolve(code, ua)
// r.Name = "Galaxy Tab S7", r.Brand = "Samsung", r.Type = devicex.TypeTablet
// r.ID   = "catalog" — what produced the answer

devicex.Resolve("", "Mozilla/5.0 (PlayStation; PlayStation 5/9.60) …")
// PlayStation 5 / Sony / Console — hardware that carries no model code
```

`ResolveCode` and `ResolveUA` take the two kinds of evidence separately. Prefix
rules are never tested against a whole User-Agent, and the `MatchContains`
rules that describe prose are never tested against a model code.

For a consumer that will not import this package, `Describe` is the same answer
in a signature naming no type from it:

```go
name, brand, model, family, deviceType, confidence, ok := devicex.Describe(code, ua)
```

## The same catalogue over HTTP

The Pages site resolves a code in the browser at a plain path — `/d/SM-S928B`.
There is no server behind it: `404.html` is what Pages returns for a path it
does not publish, and that page reads the path it was asked for and answers it.
`/?code=…` and `/#…` resolve the same code and are rewritten to the canonical
form. The page is only a reader for static JSON under `api/`, and those files
are the API.

```
GET api/SM.json       {"SM-S928B": ["Samsung", "Galaxy S24 Ultra"], …}
GET api/rules.json    the code-shape rules, for codes the catalogue does not hold
GET api/meta.json     snapshot date, device count, source
```

A shard is named after the first two characters of a code, uppercased, with
anything that is not a letter or digit replaced by `_` — so a lookup costs one
request of a few kilobytes rather than the whole catalogue. `api/devicex.js` is
an ES module doing the same three-step resolution as the Go package:

```js
import { Devicex, parseCode } from './api/devicex.js';

const dx = new Devicex('./api/');
await dx.resolve(parseCode('SM-S928B Build/UP1A.231005.007'));
// { id: 'catalog', name: 'Galaxy S24 Ultra', brand: 'Samsung', model: 'SM-S928B', confidence: 0.99 }
```

Rebuild the Go catalogue and the JSON together with `make catalog-full`, and
serve the site locally with `make site`.

A cold lookup is two round trips, not four: the page requests the shard and the
rule table from a blocking script before the module that uses them is parsed,
and the biggest shard — every `SM` code, 2437 of them — is 13 KB gzipped. Shards
are kept for the tab's session, so a second lookup touches no network at all.

One caveat worth knowing: a `/d/…` URL is served by the 404 fallback, so it
carries HTTP 404 even when it renders an answer. That is the cost of pretty
paths on a host that runs nothing. Anything programmatic should read `api/`
directly, where a hit is a 200 and a miss is a real 404.

## Two tiers, deliberately separate

| | Answers | How | When unknown |
|---|---|---|---|
| `Lookup` | which handset | catalogue record | `false` — **never approximated** |
| `BrandOf` | which manufacturer | the code's shape | `false` |
| `Resolve` | both, plus form factor | record first, then shape | `false` |

The split matters. Manufacturers allocate codes in per-vendor spaces, so `SM-`
has meant Samsung for two decades and a phone released tomorrow still resolves
its maker. The handset's *name* has no such structure and cannot degrade
gracefully — so it does not degrade at all.

`Lookup` is checked first inside `BrandOf`, because a recorded device outranks a
prefix. Sub-brands share their parent's code space: `CPH` is OPPO's prefix and
also covers OnePlus models, which is exactly the case where the record is right
and the shape is not.

## What it will not do

A code the catalogue does not hold returns `false`. Not a nearest match, not a
family guess, not the most likely device. A wrong device name is worse than no
device name: it is a fact-shaped value that a caller will store, aggregate and
report, and nothing downstream will ever question it.

Lookup is **case-sensitive**. Model codes are copied verbatim out of a
User-Agent, so matching a different case would be a weaker claim than the one
the API appears to make. `LookupFold` exists for input whose case is known to
be unreliable.

## Coverage, honestly

The bundled catalogue comes from a community list that has not kept pace with
recent releases. Handsets from 2024 onward are largely absent:

```
SM-G973F   Galaxy S10      ✓
SM-S911B   Galaxy S23      ✓
SM-S928B   —               ✗   brand resolves, name does not
```

Regenerating from a current Google Play Console export fixes this and takes one
command:

```bash
make catalog CSV=supported_devices.csv
```

The export is updated weekly. See [NOTICE.md](NOTICE.md) for provenance and
licensing of the bundled data.

## Design

- **Sorted array, binary search.** No map, so nothing is built at init and the
  strings live in read-only memory. Startup cost is zero and lookup is ~14
  comparisons — 50 ns on an M4 Pro, and `make bench` reports the number for
  your own machine rather than asking you to trust this one.
- **Generated Go, not embedded YAML or JSON.** The catalogue is bulk imported
  data with no hand-written rows, so a parseable format would buy nothing and
  cost megabytes plus a decode at startup.
- **No dependencies.** The importer reads UTF-16 CSV and JSON with the standard
  library.

## Related

[uax](https://github.com/bakhod1r/uax) parses the User-Agent that carries the
code. It consumes this package through `uax/devices` rather than bundling the
data, so the catalogue can be updated on its own schedule — phones ship weekly,
parsers do not.

## License

MIT — see [LICENSE](LICENSE). The bundled catalogue is third-party data;
its source and licence are recorded in [NOTICE.md](NOTICE.md).
