# Changelog

The catalogue and the code version together, because a lookup result is only
reproducible if you know which catalogue produced it. `adx.Source` records
where the shipped data came from.

---

## v0.1.0 — first release

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
