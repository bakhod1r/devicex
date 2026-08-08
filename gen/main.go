// Command gen builds internal/catalog/catalog_gen.go from a published Android
// device catalogue.
//
// The mapping from a model code to a marketing name is assigned by the
// manufacturer and cannot be derived, so it can only be imported. What
// separates an imported catalogue from one somebody made up is the citation,
// which is why -source is required and is written into the generated file as
// catalog.Source.
//
// Two inputs are understood:
//
//	gen -csv supported_devices.csv -source "Google Play Console export, 2026-08-08"
//	gen -json devices.json         -source "pbakondy/android-device-list (MIT), …"
//
// The CSV is the Google Play Console device list, which ships as UTF-16LE with
// a BOM. The JSON is the pbakondy/android-device-list array. Both are reduced
// to the same three columns; everything else in them is dropped.
package main

import (
	"encoding/binary"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("gen: ")

	var (
		csvPath  = flag.String("csv", "", "Google Play Console device list (UTF-16LE CSV)")
		jsonPath = flag.String("json", "", "pbakondy/android-device-list JSON")
		source   = flag.String("source", "", "citation for the data (required)")
		out      = flag.String("out", filepath.Join("internal", "catalog", "catalog_gen.go"), "file to write")
	)
	flag.Parse()

	// A catalogue without a citation is indistinguishable from invented data,
	// and the package documents that every entry is traceable. Refuse rather
	// than ship an anonymous one.
	if strings.TrimSpace(*source) == "" {
		log.Fatal("-source is required: cite where the data came from")
	}
	if (*csvPath == "") == (*jsonPath == "") {
		log.Fatal("give exactly one of -csv or -json")
	}

	var (
		rows []Entry
		err  error
	)
	switch {
	case *csvPath != "":
		rows, err = readCSV(*csvPath)
	default:
		rows, err = readJSON(*jsonPath)
	}
	if err != nil {
		log.Fatal(err)
	}

	entries, dropped := clean(rows)
	if len(entries) == 0 {
		log.Fatal("no usable rows in input")
	}

	if err := write(*out, *source, entries); err != nil {
		log.Fatal(err)
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

// readJSON reads the pbakondy/android-device-list array.
func readJSON(path string) ([]Entry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var list []struct {
		Brand string `json:"brand"`
		Name  string `json:"name"`
		Model string `json:"model"`
	}
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	rows := make([]Entry, 0, len(list))
	for _, d := range list {
		rows = append(rows, Entry{Code: d.Model, Brand: d.Brand, Name: d.Name})
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

func normaliseBrand(b string) string {
	b = strings.TrimSpace(b)
	if canonical, ok := brands[strings.ToLower(b)]; ok {
		return canonical
	}
	return b
}

func write(path, source string, entries []Entry) error {
	var b strings.Builder

	fmt.Fprintf(&b, `// Code generated by adx/gen. DO NOT EDIT.
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
// %d devices, sorted by code so Lookup can binary search.

package catalog

// Entry is one catalogue row.
type Entry struct {
	Code  string
	Brand string
	Name  string
}

// Source cites where Entries came from.
const Source = %q

// Entries is sorted by Code.
var Entries = [...]Entry{
`, source, len(entries), source)

	for _, e := range entries {
		fmt.Fprintf(&b, "\t{%q, %q, %q},\n", e.Code, e.Brand, e.Name)
	}
	b.WriteString("}\n")

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
