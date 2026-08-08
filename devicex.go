// Package devicex resolves a device model code to the device it names.
//
// A User-Agent from an Android phone carries a model code — "SM-S928B",
// "CPH2451", "2201116SG" — and nothing else about the hardware. The code is
// evidence; the handset's name is not derivable from it. Samsung assigned
// "SM-S928B" to the Galaxy S24 Ultra, and no amount of pattern matching
// recovers that. It requires a catalogue.
//
//	d, ok := devicex.Lookup("SM-G973F")
//	// d.Brand = "Samsung", d.Name = "Galaxy S10"
//
// This package is the catalogue and the rules that read a model code's shape.
// It parses no User-Agents, has no dependencies, and holds no state. It is
// separate from the parser that usually consumes it so that the data can be
// updated on its own schedule: phones ship weekly, parsers do not.
//
// Lookup is the catalogue alone. Resolve is everything the package knows about
// one request — the catalogue, the code-shape rules, and the rules for
// hardware that carries no model code at all:
//
//	r, ok := devicex.Resolve(devicex.Code(field), ua)
//	// r.Name, r.Brand, r.Type, r.Family, r.Confidence, and r.ID: what answered
//
// # What it will not do
//
// A code the catalogue does not hold returns ok == false. It never returns an
// approximation, a nearest match or a guess derived from the code's shape,
// because a wrong device name is worse than no device name — it is a fact-
// shaped value that a caller will store, aggregate and report.
//
// Brand is a separate question and is answerable without this package: model
// codes have manufacturer-specific shapes, so "SM-" implies Samsung whether or
// not the specific handset is known. See BrandOf.
//
// # Provenance
//
// Every entry comes from a published catalogue, cited in Source. Nothing here
// is remembered or inferred. See NOTICE.md.
package devicex

import (
	"sort"
	"strconv"
	"strings"

	"github.com/bakhod1r/devicex/internal/catalog"
)

// Device is a handset the catalogue knows.
type Device struct {
	// Code is the model code as it appears in a User-Agent.
	Code string

	// Brand is the manufacturer, normalised to one spelling per company.
	// Sub-brands report their parent: a Redmi Note is made by Xiaomi.
	Brand string

	// Name is what the device is sold as, for example "Galaxy S24 Ultra".
	Name string
}

// Source cites where the catalogue came from.
const Source = catalog.Source

// Generated is the date of the snapshot the catalogue was built from,
// YYYY-MM-DD. It is the age of the data, not of the release: a catalogue
// imported from a source that stopped updating carries the source's date,
// which is what decides which handsets it cannot name.
const Generated = catalog.Generated

// Version is a fingerprint of the catalogue: its snapshot date and the number
// of devices it holds, "2026-08-08/25475".
//
// A caller that stored a device name six months ago and wants to know what it
// was decided from needs the data's identity, and Source alone does not carry
// it — the same source yields a different catalogue every time it is
// re-imported. Log this beside a resolved name and the answer stays
// explainable after the catalogue moves.
func Version() string {
	return Generated + "/" + strconv.Itoa(Len())
}

// Lookup returns the device a model code names.
//
// The comparison is exact. Model codes are case-significant and are copied
// verbatim out of the User-Agent, so a case-insensitive match would be a
// different, weaker claim; use LookupFold when the input has been through
// something that changed its case.
func Lookup(code string) (Device, bool) {
	e := catalog.Entries[:]
	i := sort.Search(len(e), func(i int) bool { return e[i].Code >= code })
	if i < len(e) && e[i].Code == code {
		return Device{Code: e[i].Code, Brand: e[i].Brand, Name: e[i].Name}, true
	}
	return Device{}, false
}

// LookupFold is Lookup, ignoring case. It is slower — the catalogue is sorted
// case-sensitively, so this scans — and should only be used when the caller
// knows the input's case is unreliable.
func LookupFold(code string) (Device, bool) {
	if d, ok := Lookup(code); ok {
		return d, true
	}
	for _, e := range catalog.Entries {
		if strings.EqualFold(e.Code, code) {
			return Device{Code: e.Code, Brand: e.Brand, Name: e.Name}, true
		}
	}
	return Device{}, false
}

// Len is how many devices the catalogue holds.
func Len() int { return len(catalog.Entries) }

// All calls fn for every device, in code order. It is a range function rather
// than a returned slice because the catalogue is large and copying it to
// answer "which devices does Samsung make" would allocate megabytes.
//
// Returning false stops the iteration.
func All(fn func(Device) bool) {
	for _, e := range catalog.Entries {
		if !fn(Device{Code: e.Code, Brand: e.Brand, Name: e.Name}) {
			return
		}
	}
}

// Brands returns every manufacturer in the catalogue, sorted.
func Brands() []string {
	seen := make(map[string]bool, 256)
	for _, e := range catalog.Entries {
		seen[e.Brand] = true
	}
	out := make([]string, 0, len(seen))
	for b := range seen {
		out = append(out, b)
	}
	sort.Strings(out)
	return out
}
