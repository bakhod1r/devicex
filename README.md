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

Every code has its own JSON document, at a path that is the code. No server, no
client library, no page to render — a GET returns the answer:

```sh
curl -s https://bakhod1r.github.io/devicex/api/d/SM-S928B.json
```
```json
{"code":"SM-S928B","ok":true,"id":"catalog","name":"Galaxy S24 Ultra",
 "brand":"Samsung","model":"SM-S928B","confidence":0.99}
```

`family` and `type` appear when a rule identifies them, exactly as `Resolve`
reports them: `api/d/SM-X710.json` carries `"family":"Galaxy Tab"` and
`"type":"Tablet"`. A code the catalogue does not hold has no document, and the
404 is the answer — not in the catalogue. Nothing is approximated.

```
GET api/d/SM-S928B.json   one code, the whole answer            137 B gzipped
GET api/SM.json           every code starting SM, {code: [brand, name]}  11 KB
GET api/rules.json        the code-shape rules                    1 KB
GET api/meta.json         snapshot date, device count, source
```

The shards are for a client resolving many codes at once: one request covers a
whole code space, and a shard is named after the first two characters of a code,
uppercased, with anything that is not a letter or digit replaced by `_`. The rule
table is what turns a 404 into a manufacturer — `SM-` is Samsung whether or not
that handset is catalogued.

Two things to know about the paths. A code containing `/` has it percent-encoded
in the filename, so the URL carries it twice-encoded:
`api/d/Doro%208030%252F8031%252F8028.json`. And 188 codes that differ from
another only in case — almost all of them television model strings like
`2K SMART TV` and `2K Smart TV` — have no document at all, because a macOS or
Windows checkout of this repository cannot hold both files. They are in the
shards.

`lookup.html` is a reader for the same files, and `/d/SM-S928B` renders in the
browser: Pages serves `404.html` for the unpublished path and it answers from
the path it was asked for. Anything programmatic should use `api/d/…` instead,
where a hit is a 200 and a miss is a real 404.

Rebuild the Go catalogue and the JSON together with `make catalog-full`, and
serve them locally with `make site`.

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
