// Command gen builds internal/catalog/catalog_gen.go from a published Android
// device catalogue.
//
// The mapping from a model code to a marketing name is assigned by the
// manufacturer and cannot be derived, so it can only be imported. What
// separates an imported catalogue from one somebody made up is the citation,
// which is why -source is required and is written into the generated file as
// catalog.Source.
//
// Two inputs are understood, and may be combined:
//
//	gen -csv supported_devices.csv -source "Google Play supported devices, 2026-08-08"
//	gen -md MobileModels/brands    -source "KHwang9883/MobileModels (CC BY-NC-SA 4.0)"
//
// The CSV is Google Play's supported-devices list, which ships as UTF-16LE with
// a BOM — the broad source, and the one whose licence permits redistribution.
// The -md directory is KHwang9883/MobileModels, a hand-maintained set of
// Markdown pages listing the codes each handset ships under; it reaches the
// regional variants the Play list omits. Only its English pages are read, and
// any row naming a device in CJK script is dropped whichever source produced
// it: a catalogue that answers in two scripts cannot be aggregated, sorted or
// displayed by one caller.
//
// Given both, the CSV decides: where the two disagree about a code, the source
// that can be redistributed wins, and the Markdown only adds codes the CSV
// never held. Both are reduced to the same three columns.
//
// MobileModels is CC BY-NC-SA 4.0. A catalogue built with -md inherits that —
// non-commercial, share-alike — which is not the licence of this repository's
// code. Whoever runs it with -md is choosing that for the generated file.
package main

import (
	"encoding/binary"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf16"
	"unicode/utf8"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("gen: ")

	var (
		csvPath = flag.String("csv", "", "Google Play Console device list (UTF-16LE CSV)")
		mdDir   = flag.String("md", "", "KHwang9883/MobileModels brands/ directory (CC BY-NC-SA 4.0)")
		source  = flag.String("source", "", "citation for the data (required)")
		dated   = flag.String("generated", time.Now().UTC().Format("2006-01-02"), "date of the input snapshot, YYYY-MM-DD; not the date of this run unless they are the same")
		out     = flag.String("out", filepath.Join("internal", "catalog", "catalog_gen.go"), "file to write")
		web     = flag.String("web", "", "also write the static lookup API (sharded JSON) into this directory")
	)
	flag.Parse()

	// A catalogue without a citation is indistinguishable from invented data,
	// and the package documents that every entry is traceable. Refuse rather
	// than ship an anonymous one.
	if strings.TrimSpace(*source) == "" {
		log.Fatal("-source is required: cite where the data came from")
	}
	// A catalogue's age is the first thing a caller asks about a wrong answer,
	// so the date is shipped as data rather than left in a commit message. A
	// malformed one would be worse than none: it would read as a fact.
	if _, err := time.Parse("2006-01-02", *dated); err != nil {
		log.Fatalf("-generated must be YYYY-MM-DD: %v", err)
	}
	if *csvPath == "" && *mdDir == "" {
		log.Fatal("give -csv, -md, or both")
	}

	// Order matters: clean keeps the first row it sees for a code, so the CSV
	// is read first and the Markdown can only fill gaps.
	var rows []Entry
	if *csvPath != "" {
		r, err := readCSV(*csvPath)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s: %d rows\n", *csvPath, len(r))
		rows = append(rows, r...)
	}
	if *mdDir != "" {
		r, err := readMarkdown(*mdDir)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s: %d rows\n", *mdDir, len(r))
		rows = append(rows, r...)
	}

	entries, dropped := clean(rows)
	if len(entries) == 0 {
		log.Fatal("no usable rows in input")
	}

	if err := write(*out, *source, *dated, entries); err != nil {
		log.Fatal(err)
	}
	if *web != "" {
		if err := writeWeb(*web, *source, *dated, entries); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("%s: %d devices, %d rows dropped\n", *out, len(entries), dropped)
}

// Entry is one catalogue row, in the shape the generated file holds.
type Entry struct {
	Code  string
	Brand string
	Name  string
}

// readCSV reads the Google Play Console device list.
//
// The export is UTF-16LE with a BOM, which neither encoding/csv nor anything
// else in the standard library decodes, so it is converted up front. Columns
// are located by header name because their order has changed between exports.
func readCSV(path string) ([]Entry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(strings.NewReader(decode(raw)))
	r.FieldsPerRecord = -1 // exports carry trailing columns that come and go

	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("reading header: %w", err)
	}

	col := map[string]int{}
	for i, h := range header {
		col[strings.ToLower(strings.TrimSpace(h))] = i
	}

	// "Model" is the code that appears in a User-Agent. Several Play Console
	// exports look similar but omit it — the reach-and-devices export, for
	// one, carries a "Device" codename instead, which never reaches the
	// network. Say so plainly rather than generating a catalogue that can
	// never match its input.
	iModel, ok := col["model"]
	if !ok {
		return nil, fmt.Errorf("%s has no Model column (found: %s)\n"+
			"this is not the supported-devices export; Model holds the code a "+
			"User-Agent carries, such as SM-G973F", path, strings.Join(header, ", "))
	}
	iBrand, ok := col["retail branding"]
	if !ok {
		return nil, fmt.Errorf("%s has no Retail Branding column", path)
	}
	iName, ok := col["marketing name"]
	if !ok {
		return nil, fmt.Errorf("%s has no Marketing Name column", path)
	}

	var rows []Entry
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if iModel >= len(rec) || iBrand >= len(rec) || iName >= len(rec) {
			continue
		}
		rows = append(rows, Entry{
			Code:  rec[iModel],
			Brand: rec[iBrand],
			Name:  rec[iName],
		})
	}
	return rows, nil
}

// mdPages are the MobileModels pages this importer reads and the brand each one
// reports.
//
// English pages only. The repository also documents the domestic Chinese
// variants, and those pages name them in Chinese — which is the only name they
// have, but not a string an English-language report can aggregate by. A
// catalogue that answers half its callers in a script they cannot read is worse
// than one that answers nothing, so those pages are not read — and clean drops
// any row carrying CJK text regardless of which source produced it.
//
// The set is explicit rather than derived from the filenames, because a
// filename is not a brand — "zhixuan" is a Huawei sub-brand, "mitv_cn" is
// Xiaomi's televisions, "360shouji" is 360. Pages added upstream after this was
// written are skipped until someone looks at them: a new file has to be read
// before its rows are believed.
//
// Apple's page is deliberately absent. It keys devices by their regulatory
// "A1429" numbers, which no User-Agent carries, so importing them would add 400
// codes that can never match and can only collide with some other
// manufacturer's real ones.
var mdPages = []struct {
	file  string
	brand string
}{
	{"asus_en.md", "Asus"},
	{"blackshark_en.md", "Black Shark"},
	{"google.md", "Google"},
	{"honor_global_en.md", "Honor"},
	{"huawei_global_en.md", "Huawei"},
	{"meizu_en.md", "Meizu"},
	{"mitv_global_en.md", "Xiaomi"},
	{"nothing.md", "Nothing"},
	{"oneplus_en.md", "OnePlus"},
	{"oppo_global_en.md", "OPPO"},
	{"realme_global_en.md", "realme"},
	{"samsung_global_en.md", "Samsung"},
	{"sony.md", "Sony"},
	{"vivo_global_en.md", "vivo"},
	{"xiaomi_en.md", "Xiaomi"},
}

// mdEntry matches one MobileModels model line:
//
//	`SM-G973F`: Galaxy S10 Global
//	`CPH1871` `CPH1875`: OPPO Find X
//
// The codes are backquoted and there may be several, all naming the same
// handset; everything after the colon is the name. Lines that are not this
// shape — headings, the codename lines in bold, prose — carry no model code and
// are skipped.
var mdEntry = regexp.MustCompile("^((?:`[^`]+` *)+): *(.+)$")

// mdQualifiers are the market and SIM-configuration suffixes MobileModels
// appends to a device's name: "Galaxy S10 Global", "Galaxy S10 South Korea",
// "HUAWEI Mate 8 Dual SIM". They belong to the code, not to the device — the
// code is what distinguishes those variants, and it is the catalogue's key —
// and leaving them on means the same handset appears under a dozen different
// names, none of which is what it is sold as.
//
// The list is closed and matched only at the end of a name. Trimming a known
// qualifier drops a fact the code already carries; anything cleverer would be
// rewriting a marketing name, which this importer does not do.
var mdQualifiers = []string{
	"Global", "China mainland", "China", "India", "Japan", "Taiwan",
	"Hong Kong", "South Korea", "Canada", "US Carrier", "US Unlocked",
	"Dual SIM", "Single SIM",
}

// trimQualifier removes the qualifiers from the end of a name, repeatedly: a
// few rows carry two ("Galaxy A52s 5G Global Dual SIM").
func trimQualifier(name string) string {
	for again := true; again; {
		again = false
		for _, q := range mdQualifiers {
			if trimmed, ok := strings.CutSuffix(name, " "+q); ok {
				name, again = strings.TrimSpace(trimmed), true
			}
		}
	}
	return name
}

// readMarkdown reads the MobileModels brands/ directory.
//
// The pages are hand-written prose, so this parses conservatively: a line that
// does not have the exact shape above is dropped rather than guessed at. The
// alternative — recovering names from headings — would invent rows, and a wrong
// device name is the one thing this package must never produce.
func readMarkdown(dir string) ([]Entry, error) {
	var rows []Entry
	for _, page := range mdPages {
		raw, err := os.ReadFile(filepath.Join(dir, page.file))
		if err != nil {
			return nil, fmt.Errorf("%s: %w (is this MobileModels/brands?)", page.file, err)
		}

		for _, line := range strings.Split(string(raw), "\n") {
			m := mdEntry.FindStringSubmatch(strings.TrimSpace(line))
			if m == nil {
				continue
			}
			device := trimQualifier(strings.TrimSpace(m[2]))
			for _, code := range strings.Split(m[1], "`") {
				code = strings.TrimSpace(code)
				if code == "" {
					continue
				}
				rows = append(rows, Entry{Code: code, Brand: page.brand, Name: device})
			}
		}
	}
	return rows, nil
}

// decode converts UTF-16 input to UTF-8, leaving UTF-8 input alone. The Play
// Console export is UTF-16LE with a BOM; the community JSON and hand-made CSVs
// are UTF-8 without one.
func decode(b []byte) string {
	switch {
	case len(b) >= 2 && b[0] == 0xFF && b[1] == 0xFE:
		return decode16(b[2:], binary.LittleEndian)
	case len(b) >= 2 && b[0] == 0xFE && b[1] == 0xFF:
		return decode16(b[2:], binary.BigEndian)
	case len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF:
		return string(b[3:])
	default:
		return string(b)
	}
}

func decode16(b []byte, order binary.ByteOrder) string {
	u := make([]uint16, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		u = append(u, order.Uint16(b[i:]))
	}
	return string(utf16.Decode(u))
}

// generic are model columns holding a word rather than a code. Indexing them
// would match nearly every User-Agent, since "Android" appears in all of them.
var generic = map[string]bool{
	"android": true,
	"linux":   true,
	"k":       true,
	"unknown": true,
	"mobile":  true,
	"phone":   true,
	"tablet":  true,
	"generic": true,
	"build":   true,
	"device":  true,
	"n/a":     true,
}

// brands maps the spellings one manufacturer appears under to the single one
// this catalogue reports. Sub-brands report their parent: a Redmi Note is made
// by Xiaomi, and a caller aggregating by brand wants one row for the company.
var brands = map[string]string{
	"lge":                 "LG",
	"lg electronics":      "LG",
	"tct (alcatel)":       "Alcatel",
	"tct":                 "Alcatel",
	"alcatel one touch":   "Alcatel",
	"redmi":               "Xiaomi",
	"poco":                "Xiaomi",
	"xiaomi communcation": "Xiaomi", // upstream typo, kept verbatim as a key
	"samsung electronics": "Samsung",
	"huawei technologies": "Huawei",
	"sony ericsson":       "Sony",
	"sony mobile":         "Sony",
	"motorola mobility":   "Motorola",
	"asustek":             "Asus",
	"asustek computer":    "Asus",
	"oneplus technology":  "OnePlus",
	"guangdong oppo":      "OPPO",
	"oppo mobile":         "OPPO",

	// The Play export capitalises these as ordinary words; the companies do
	// not. A caller aggregating by brand must not get "Oppo" and "OPPO" as two
	// manufacturers, so the spelling is decided here.
	"oppo":            "OPPO",
	"vivo":            "vivo",
	"realme":          "realme",
	"realme techlife": "realme",
	"tecno":           "Tecno",
	"tecno mobile":    "Tecno",
}

// clean applies every transformation between the upstream rows and the shipped
// array, and is the reason the shipped row count differs from upstream's.
func clean(rows []Entry) (out []Entry, dropped int) {
	seen := make(map[string]bool, len(rows))

	for _, r := range rows {
		code := strings.TrimSpace(r.Code)
		brand := normaliseBrand(r.Brand)
		name := strings.TrimSpace(r.Name)

		switch {
		case code == "" || brand == "" || name == "":
			dropped++
			continue

		// A word rather than a code.
		case generic[strings.ToLower(code)]:
			dropped++
			continue

		// Too short to be a model code. The shortest real one in the
		// catalogue is four characters, and a two-character key would
		// collide with far too much.
		case utf8.RuneCountInString(code) < 4:
			dropped++
			continue

		// The name merely repeats the code, so the row answers nothing that
		// the caller did not already hold.
		case strings.EqualFold(name, code):
			dropped++
			continue

		// Control characters would break the generated Go, and a code
		// carrying one is corrupt rather than unusual.
		case strings.ContainsAny(code, "\"\\\n\r\t"):
			dropped++
			continue

		// A name in Chinese, Japanese or Korean script. Sources that document
		// domestic-market handsets carry these, and they are the only names
		// those devices have — but a catalogue whose answers switch script
		// halfway through cannot be aggregated, sorted or displayed by a caller
		// expecting one language. Dropping the row is honest; transliterating
		// it would be this package inventing a marketing name.
		case hasCJK(name) || hasCJK(code):
			dropped++
			continue
		}

		if seen[code] {
			// Duplicate codes are common upstream: the same handset appears
			// once per region. Keep the first, which is stable because the
			// output is sorted afterwards.
			dropped++
			continue
		}
		seen[code] = true
		out = append(out, Entry{Code: code, Brand: brand, Name: name})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out, dropped
}

// hasCJK reports whether s contains a Han, Hiragana, Katakana or Hangul rune.
func hasCJK(s string) bool {
	for _, r := range s {
		switch {
		case r >= 0x3000 && r <= 0x303F, // CJK punctuation
			r >= 0x3040 && r <= 0x30FF, // kana
			r >= 0x3400 && r <= 0x4DBF, // Han extension A
			r >= 0x4E00 && r <= 0x9FFF, // Han
			r >= 0xAC00 && r <= 0xD7AF, // Hangul
			r >= 0xF900 && r <= 0xFAFF, // Han compatibility
			r >= 0xFF01 && r <= 0xFF60: // fullwidth forms
			return true
		}
	}
	return false
}

func normaliseBrand(b string) string {
	b = strings.TrimSpace(b)
	if canonical, ok := brands[strings.ToLower(b)]; ok {
		return canonical
	}
	return b
}

func write(path, source, dated string, entries []Entry) error {
	var b strings.Builder

	fmt.Fprintf(&b, `// Code generated by devicex/gen. DO NOT EDIT.
//
// Android device catalogue: model code to marketing name.
//
// Source: %s
//
// This mapping is assigned by manufacturers and cannot be derived from the
// code, so it is imported rather than reasoned about. A code absent here is
// absent from the catalogue and resolves no name at all — never an approximate
// one, because a wrong device name is worse than none.
//
// Snapshot: %s. %d devices, sorted by code so Lookup can binary search.

package catalog

// Entry is one catalogue row.
type Entry struct {
	Code  string
	Brand string
	Name  string
}

// Source cites where Entries came from.
const Source = %q

// Generated is the date of the snapshot Entries was built from, YYYY-MM-DD.
// It is the age of the data, not of the build: a catalogue imported today from
// a source frozen two years ago carries the source's date, because that is
// what decides which handsets are missing.
const Generated = %q

// Entries is sorted by Code.
var Entries = [...]Entry{
`, source, dated, len(entries), source, dated)

	for _, e := range entries {
		fmt.Fprintf(&b, "\t{%q, %q, %q},\n", e.Code, e.Brand, e.Name)
	}
	b.WriteString("}\n")

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
