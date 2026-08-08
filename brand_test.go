package devicex_test

import (
	"testing"

	"github.com/bakhod1r/devicex"
)

// The rules below were carried over from data/device.yaml, which used to hold
// them as declarative match rules with a test User-Agent attached to each one.
// The codes here are the model tokens those User-Agents carried, so this test
// is the YAML file's own test suite, kept after the data moved into Go.
//
// Rules the catalogue answers on its own are not repeated here — those are
// covered by the catalogue's tests. What is checked is the shape-matching
// tier, which is the part that has to keep working for hardware released
// after this catalogue was built.
func TestBrandOfCoversTheImportedRules(t *testing.T) {
	t.Parallel()

	cases := []struct {
		code  string
		brand string
	}{
		{"SM-G973F", "Samsung"},
		{"GT-I9300", "Samsung"},
		{"CPH9999", "OPPO"},
		{"RMX3771", "realme"},
		{"ELS-NX9", "Huawei"},
		{"VOG-L29", "Huawei"},
		{"moto g52", "Motorola"},
		{"XT2143-1", "Motorola"},
		{"Redmi Note 12", "Xiaomi"},
		{"POCO F5", "Xiaomi"},
		{"Nokia G21", "Nokia"},

		// Added when the YAML was retired: these had no Go equivalent.
		{"vivo 1920", "vivo"},
		{"XQ-DQ54", "Sony"},
		{"KFTRWI", "Amazon"},
		{"M2999X9Z", "Xiaomi"},
		{"2210132G", "Xiaomi"},

		// Not Android, and so absent from the catalogue by construction. The
		// token is the whole identifier the request carries.
		{"iPhone", "Apple"},
		{"iPad", "Apple"},
		{"iPod touch", "Apple"},
		{"Macintosh", "Apple"},
		{"Nintendo Switch", "Nintendo"},
		{"PlayStation 5", "Sony"},
		{"PlayStation 4", "Sony"},
		{"Xbox", "Microsoft"},

		// Google's phones are matched by shape, not by catalogue entry: the
		// Play export keys them by board name ("oriole"), while a User-Agent
		// carries the marketing name.
		{"Pixel 9 Pro XL", "Google"},
		{"Pixel 8a", "Google"},
		{"Pixel 6", "Google"},
	}

	for _, c := range cases {
		got, ok := devicex.BrandOf(c.code)
		if !ok {
			t.Errorf("BrandOf(%q) identified nothing, want %q", c.code, c.brand)
			continue
		}
		if got != c.brand {
			t.Errorf("BrandOf(%q) = %q, want %q", c.code, got, c.brand)
		}
	}
}

// The weakest rules are last on purpose: a code that some other manufacturer
// registered must not be captured by Xiaomi's habit of numeric codes.
func TestWeakPrefixesDoNotOutrankRegisteredOnes(t *testing.T) {
	t.Parallel()

	// "M2" would match this if it were tested before Motorola's "moto ".
	if b, _ := devicex.BrandOf("moto g play 2021"); b != "Motorola" {
		t.Errorf("BrandOf(moto g play 2021) = %q", b)
	}
	// A catalogue hit must beat every prefix rule.
	if d, ok := devicex.Lookup("SM-G973F"); !ok || d.Brand != "Samsung" {
		t.Errorf("Lookup(SM-G973F) = %+v, %v", d, ok)
	}
}

// An unrecognised shape has to stay unrecognised. This is the rule the whole
// package is arranged around: no fallback to the most common manufacturer.
func TestBrandOfRefusesTheUnknown(t *testing.T) {
	t.Parallel()

	for _, code := range []string{"", "ZZZ-9999", "Mozilla", "Linux", "?"} {
		if b, ok := devicex.BrandOf(code); ok {
			t.Errorf("BrandOf(%q) = %q, want no match", code, b)
		}
	}
}
