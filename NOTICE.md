# Third-party data

The Go code in this repository is original. The catalogue it ships is not.

## Device catalogue — `internal/catalog/catalog_gen.go`

**Source:** [Google Play's supported-devices list](https://support.google.com/googleplay/android-developer/answer/6154891),
published at `https://storage.googleapis.com/play_public/supported_devices.csv`
**Snapshot:** 2026-08-08 — recorded in the code as `devicex.Generated`
**Licence:** published by Google as a developer reference; the four columns used
here (retail branding, marketing name, device, model) are manufacturer-assigned
facts, not authored content.

The mapping from a model code to a marketing name is a fact assigned by the
manufacturer. No rule derives it, so it can only be imported, and the citation
is what separates an imported catalogue from one somebody made up. The
generator refuses to run without a `-source`.

**Second source:** [KHwang9883/MobileModels](https://github.com/KHwang9883/MobileModels),
its English pages only. The Play list records what a device ships to Google Play
as; MobileModels records every code one handset sells under, which is how the
regional variants get names. Where the two disagree about a code, the Play list
decides.

Its Chinese pages are not read, and any row naming a device in CJK script is
dropped whichever source produced it. Those names are the only ones the
domestic-market handsets have, so this loses them entirely — the alternative is
a catalogue that answers half its callers in a script they cannot use, and
transliterating would mean inventing a marketing name.

**Licence of that second source: CC BY-NC-SA 4.0.** It is non-commercial and
share-alike, and the shipped catalogue inherits both terms even though the Go
code in this repository does not. Attribution above satisfies BY. If that is
not acceptable for your use, rebuild from the Play list alone — no other file
changes:

    make catalog CSV=supported_devices.csv

**Known limitation:** the Play list is what Google Play ships to, so a device
that never shipped through Play, or shipped after the snapshot, is absent —
`Lookup` returns false for it while `BrandOf` still resolves the manufacturer
from the code's shape. Refresh with:

    curl -O https://storage.googleapis.com/play_public/supported_devices.csv
    make catalog CSV=supported_devices.csv

`devicex.Version()` reports which snapshot answered, so a stored name stays
explainable after the catalogue moves.

## What the importer changes

The upstream data is not copied verbatim. The generator:

- normalises manufacturer spellings, so one company is one brand — `LGE` and
  `LG Electronics` become `LG`, `TCT (Alcatel)` becomes `Alcatel`, `Oppo` and
  `Vivo` take the spellings their makers use, and the Xiaomi sub-brands `Redmi`
  and `POCO` report `Xiaomi`;
- drops rows whose model column holds a word rather than a code (`Android`,
  `Linux`, `K`), which would otherwise match nearly every User-Agent;
- drops rows whose name merely repeats the code, since they answer nothing;
- deduplicates by code and sorts, so the output is deterministic;
- skips the MobileModels Apple page, whose regulatory "A1429" numbers appear
  in no User-Agent, its Chinese pages, and any page added upstream since this
  import;
- drops any row whose code or name carries CJK script;
- trims the market and SIM qualifiers MobileModels appends to a name — `Galaxy
  S10 Global`, `Galaxy S10 South Korea`, `HUAWEI Mate 8 Dual SIM` all become the
  handset's name, because the code is what distinguishes those variants and the
  code is the catalogue's key.

Those transformations are in `gen/main.go` and are the reason the shipped row
count differs from upstream's.
