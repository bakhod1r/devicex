---
title: devicex
description: A device model code, resolved to the device it names.
permalink: /
---

# devicex

A device model code, resolved to the device it names — as a Go package, and
as static JSON anyone can fetch. (Includes Android devices, Apple devices, and gaming consoles).

<p style="font-size:1.05rem"><strong><a href="lookup.html">Look up a code →</a></strong></p>

```
/d/SM-S928B   →   Galaxy S24 Ultra, Samsung
```

25,475 devices, snapshot 2026-08-08. No server, no dependencies, nothing
inferred: a code the catalogue does not hold resolves its manufacturer from the
code's shape and no name at all.

## In Go

```go
import "github.com/bakhod1r/devicex"

d, ok := devicex.Lookup("SM-G973F")
// d.Brand = "Samsung", d.Name = "Galaxy S10"

r, ok := devicex.Resolve(devicex.Code(field), ua)
// r.Name, r.Brand, r.Model, r.Type, r.Family, r.Confidence, r.ID
```

~50 ns per lookup, zero allocations, zero dependencies. See the
[README](README.html) for the full API.

## Over HTTP

The JSON under `api/` is the API. Every code has its own document; the shards
are for a client resolving many at once, and are named after the first two
characters of a code, so a lookup transfers kilobytes rather than a catalogue:

```sh
curl -s https://bakhod1r.github.io/devicex/api/d/SM-S928B.json
# {"code":"SM-S928B","ok":true,"id":"catalog","name":"Galaxy S24 Ultra",
#  "brand":"Samsung","model":"SM-S928B","confidence":0.99}
```

| File | What it holds |
|---|---|
| `api/d/<code>.json` | one code, the whole answer — `api/d/SM-S928B.json` |
| `api/XX.json` | every catalogued code starting `XX`, as `{code: [brand, name]}` |
| `api/rules.json` | the code-shape rules, for codes the catalogue does not hold |
| `api/meta.json` | snapshot date, device count, source |
| `api/devicex.js` | the browser client, an ES module |

## Where the data comes from

Google Play's published supported-devices list, supplemented by the English
pages of [KHwang9883/MobileModels](https://github.com/KHwang9883/MobileModels)
for the regional variants it omits. Nothing is remembered or inferred.

The MobileModels data is CC BY-NC-SA 4.0, and the shipped catalogue inherits
that — non-commercial and share-alike, which is not the licence of the Go code.
[NOTICE.md](NOTICE.html) has the details and the one command that rebuilds
without it.
