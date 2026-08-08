# Changelog

The catalogue and the code version together, because a lookup result is only
reproducible if you know which catalogue produced it. `adx.Source` records
where the shipped data came from.

---

## v0.2.0 — the rules became callable

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
- `Describe(code, ua) (name, brand, family, deviceType string, confidence
  float64, ok bool)` — `Resolve` behind a signature naming no type from this
  package, so a consumer can accept it as a function value without importing
  adx. `Names` only reached the catalogue, which left it silent about iPhones,
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
