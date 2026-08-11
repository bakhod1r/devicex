package devicex_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/bakhod1r/devicex"
)

// Spot checks against handsets whose codes are widely published. They are not
// a proof of the whole catalogue — nothing short of the upstream source is —
// but they catch an import that silently shifted a column or lost a row.
func TestLookupKnownDevices(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ code, brand, name string }{
		{"SM-G973F", "Samsung", "Galaxy S10"},
		{"SM-G950F", "Samsung", "Galaxy S8"},
		{"CPH2451", "OnePlus", "OnePlus 11 5G"},
	} {
		d, ok := devicex.Lookup(tc.code)
		if !ok {
			t.Errorf("%s: not found", tc.code)
			continue
		}
		if d.Brand != tc.brand {
			t.Errorf("%s: brand got %q, want %q", tc.code, d.Brand, tc.brand)
		}
		if d.Name != tc.name {
			t.Errorf("%s: name got %q, want %q", tc.code, d.Name, tc.name)
		}
	}
}

// The catalogue must say "I do not know" rather than produce something
// plausible. This is the property the whole package exists to protect.
func TestUnknownCodeResolvesNothing(t *testing.T) {
	t.Parallel()

	for _, code := range []string{
		"SM-ZZ999X",  // Samsung-shaped, not a real device
		"NOTADEVICE", // nothing like a code
		"",           // empty
		"SM-G973",    // a real code with its suffix removed
		"sm-g973f",   // the right device, wrong case
	} {
		if d, ok := devicex.Lookup(code); ok {
			t.Errorf("Lookup(%q) invented %+v", code, d)
		}
	}
}

// LookupFold exists precisely for the case Lookup refuses.
func TestLookupFoldAcceptsChangedCase(t *testing.T) {
	t.Parallel()

	d, ok := devicex.LookupFold("sm-g973f")
	if !ok {
		t.Fatal("LookupFold did not find a known device in lower case")
	}
	if d.Name != "Galaxy S10" {
		t.Errorf("got %q, want Galaxy S10", d.Name)
	}

	// Exact match bypass
	if _, ok := devicex.LookupFold("SM-G973F"); !ok {
		t.Errorf("LookupFold(SM-G973F) did not find exact match")
	}

	if _, ok := devicex.LookupFold("NOTADEVICE"); ok {
		t.Errorf("LookupFold(NOTADEVICE) invented a device")
	}
}

func TestAll(t *testing.T) {
	t.Parallel()

	count := 0
	devicex.All(func(d devicex.Device) bool {
		count++
		return true
	})
	if count != devicex.Len() {
		t.Errorf("All() yielded %d entries, want %d", count, devicex.Len())
	}

	// Test early exit
	countEarly := 0
	devicex.All(func(d devicex.Device) bool {
		countEarly++
		return countEarly < 5 // return false after 5 iterations
	})
	if countEarly != 5 {
		t.Errorf("All() did not exit early, got %d", countEarly)
	}
}

func TestAt(t *testing.T) {
	t.Parallel()

	// At must agree with All: same devices, same order.
	i := 0
	devicex.All(func(want devicex.Device) bool {
		got, ok := devicex.At(i)
		if !ok || got != want {
			t.Errorf("At(%d) = %+v, %v; All gave %+v", i, got, ok, want)
			return false
		}
		i++
		return true
	})

	for _, i := range []int{-1, devicex.Len(), devicex.Len() + 1} {
		if d, ok := devicex.At(i); ok {
			t.Errorf("At(%d) = %+v, want out of range", i, d)
		}
	}
}

func TestNames(t *testing.T) {
	t.Parallel()

	name, brand, ok := devicex.Names("SM-G973F")
	if !ok || name != "Galaxy S10" || brand != "Samsung" {
		t.Errorf("Names(SM-G973F) = %q, %q, %v", name, brand, ok)
	}

	_, _, ok = devicex.Names("NOTADEVICE")
	if ok {
		t.Errorf("Names(NOTADEVICE) invented a device")
	}
}

// Lookup binary searches, which is only correct on sorted input. A generator
// change that lost the ordering would make lookups fail at random rather than
// fail loudly.
func TestCatalogueIsSorted(t *testing.T) {
	t.Parallel()

	var codes []string
	devicex.All(func(d devicex.Device) bool {
		codes = append(codes, d.Code)
		return true
	})

	if len(codes) != devicex.Len() {
		t.Fatalf("All yielded %d devices, Len reports %d", len(codes), devicex.Len())
	}
	if !sort.StringsAreSorted(codes) {
		for i := 1; i < len(codes); i++ {
			if codes[i-1] > codes[i] {
				t.Fatalf("catalogue is unsorted at %d: %q then %q", i, codes[i-1], codes[i])
			}
		}
	}
}

// Every row must be usable. A blank name is a row that costs binary size and
// answers nothing; a name equal to the code is the same.
func TestEveryEntryIsUseful(t *testing.T) {
	t.Parallel()

	var bad int
	devicex.All(func(d devicex.Device) bool {
		switch {
		case d.Code == "" || d.Name == "":
			t.Errorf("incomplete entry: %+v", d)
			bad++
		case strings.EqualFold(d.Code, d.Name):
			t.Errorf("entry names itself: %+v", d)
			bad++
		}
		return bad < 10 // report a sample, do not print the whole catalogue
	})
}

// The catalogue must never index a word that appears in every User-Agent.
// "Android" in the model column of an upstream export is a data error, and
// importing it would make every Android request resolve a device.
func TestNoToxicCodes(t *testing.T) {
	t.Parallel()

	for _, code := range []string{"Android", "Linux", "Chrome", "Safari", "Mozilla", "K", "Unknown"} {
		if d, ok := devicex.Lookup(code); ok {
			t.Errorf("%q is indexed as a device: %+v", code, d)
		}
	}
}

// A brand is answerable from a code's shape even when the handset is unknown,
// and that is what keeps new phones from degrading to nothing.
func TestBrandOfUnknownButShapedCode(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ code, brand string }{
		{"SM-S928B", "Samsung"}, // 2024 flagship, absent from this catalogue
		{"SM-ZZ999X", "Samsung"},
		{"RMX3999", "realme"},
		{"CPH9999", "OPPO"},
		{"Pixel 99", "Google"},
	} {
		got, ok := devicex.BrandOf(tc.code)
		if !ok {
			t.Errorf("BrandOf(%q): no brand from a recognisable code space", tc.code)
			continue
		}
		if got != tc.brand {
			t.Errorf("BrandOf(%q) = %q, want %q", tc.code, got, tc.brand)
		}
	}
}

func TestBrandOfNonsenseIsUnknown(t *testing.T) {
	t.Parallel()

	for _, code := range []string{"", "ZZZZ", "12345", "Mozilla/5.0"} {
		if brand, ok := devicex.BrandOf(code); ok {
			t.Errorf("BrandOf(%q) = %q, want no brand", code, brand)
		}
	}
}

// The catalogue is imported data, and the citation is what separates it from
// invention.
func TestSourceIsRecorded(t *testing.T) {
	t.Parallel()

	if devicex.Source == "" {
		t.Error("the catalogue carries no source")
	}
}

func TestBrandsAreNormalised(t *testing.T) {
	t.Parallel()

	brands := devicex.Brands()
	if len(brands) == 0 {
		t.Fatal("no brands")
	}

	// Aliases the importer is supposed to collapse. Seeing one means a caller
	// grouping by brand would split one manufacturer across several buckets.
	for _, alias := range []string{"LGE", "TCT (Alcatel)", "Oppo", "Redmi", "POCO"} {
		for _, b := range brands {
			if b == alias {
				t.Errorf("brand %q survived normalisation", alias)
			}
		}
	}
}

func BenchmarkLookup(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		devicex.Lookup("SM-G973F")
	}
}

func BenchmarkLookupMiss(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		devicex.Lookup("SM-S928B")
	}
}

// The catalogue answers in one script.
//
// Sources that document domestic Chinese handsets name them in Chinese, which
// is the only name those devices have — and not a string a caller aggregating,
// sorting or displaying device names in English can use. The importer drops
// those rows; this is the guard that says so out loud, because the next
// catalogue refresh is the moment they would come back unnoticed.
func TestCatalogueCarriesNoCJK(t *testing.T) {
	t.Parallel()

	cjk := func(s string) bool {
		for _, r := range s {
			switch {
			case r >= 0x3000 && r <= 0x303F,
				r >= 0x3040 && r <= 0x30FF,
				r >= 0x3400 && r <= 0x4DBF,
				r >= 0x4E00 && r <= 0x9FFF,
				r >= 0xAC00 && r <= 0xD7AF,
				r >= 0xF900 && r <= 0xFAFF,
				r >= 0xFF01 && r <= 0xFF60:
				return true
			}
		}
		return false
	}

	bad := 0
	devicex.All(func(d devicex.Device) bool {
		if cjk(d.Code) || cjk(d.Brand) || cjk(d.Name) {
			if bad < 5 {
				t.Errorf("CJK in catalogue row: %q / %q / %q", d.Code, d.Brand, d.Name)
			}
			bad++
		}
		return true
	})
	if bad > 5 {
		t.Errorf("... and %d more rows", bad-5)
	}
}
