# Third-party data

The Go code in this repository is original. The catalogue it ships is not.

## Device catalogue — `internal/catalog/catalog_gen.go`

**Source:** [pbakondy/android-device-list](https://github.com/pbakondy/android-device-list)
**Licence:** MIT
**Upstream origin:** that project derives its list from
[Google Play's supported-devices catalogue](https://support.google.com/googleplay/android-developer/answer/6154891).

The mapping from a model code to a marketing name is a fact assigned by the
manufacturer. No rule derives it, so it can only be imported, and the citation
is what separates an imported catalogue from one somebody made up. The
generator refuses to run without a `-source`.

**Known limitation:** the upstream list is community-maintained and has not kept
pace with recent releases. Handsets from 2024 onward are largely absent, so
`Lookup` returns false for them while `BrandOf` still resolves the manufacturer
from the code's shape. Regenerating from a current Google Play Console export
replaces the catalogue entirely:

    make catalog CSV=supported_devices.csv

## What the importer changes

The upstream data is not copied verbatim. The generator:

- normalises manufacturer spellings, so one company is one brand — `LGE` and
  `LG Electronics` become `LG`, `TCT (Alcatel)` becomes `Alcatel`, and the
  Xiaomi sub-brands `Redmi` and `POCO` report `Xiaomi`;
- drops rows whose model column holds a word rather than a code (`Android`,
  `Linux`, `K`), which would otherwise match nearly every User-Agent;
- drops rows whose name merely repeats the code, since they answer nothing;
- deduplicates by code and sorts, so the output is deterministic.

Those transformations are in `gen/main.go` and are the reason the shipped row
count differs from upstream's.
