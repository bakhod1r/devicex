---
title: devicex
description: An Android model code, resolved to the device it names.
permalink: /
---

# devicex

An Android model code, resolved to the device it names — as a Go package, and
as static JSON anyone can fetch.

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

The JSON under `api/` is the API. A shard is named after the first two
characters of a code, so a lookup transfers kilobytes rather than a catalogue:

```sh
curl -s https://bakhod1r.github.io/devicex/api/SM.json | jq '.["SM-S928B"]'
# ["Samsung", "Galaxy S24 Ultra"]
```

| File | What it holds |
|---|---|
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
