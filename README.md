# adx

Android model code → the device it names.

```go
d, ok := adx.Lookup("SM-G973F")
// d.Brand = "Samsung", d.Name = "Galaxy S10"
```

12,260 devices. 31 ns per lookup, zero allocations, zero dependencies.

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
go get github.com/bakhod1r/adx
```

## Use

```go
// The device, when the catalogue knows it.
d, ok := adx.Lookup("CPH2451")   // OnePlus 11 5G

// The manufacturer, which is answerable even when the handset is not.
brand, ok := adx.BrandOf("SM-S928B")   // "Samsung" — code shape, not a record

// Everything, without copying megabytes.
adx.All(func(d adx.Device) bool {
    if d.Brand == "Sony" {
        fmt.Println(d.Code, d.Name)
    }
    return true
})

adx.Len()      // 12260
adx.Brands()   // every manufacturer, sorted
adx.Source     // where the data came from
```

## Two tiers, deliberately separate

| | Answers | How | When unknown |
|---|---|---|---|
| `Lookup` | which handset | catalogue record | `false` — **never approximated** |
| `BrandOf` | which manufacturer | the code's shape | `false` |

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
  comparisons.
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
