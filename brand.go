package devicex

import "strings"

// Manufacturer code spaces.
//
// A marketing name has to be looked up, but a manufacturer often does not:
// model codes are allocated in per-vendor spaces with recognisable shapes.
// "SM-" has been Samsung's prefix across two decades of handsets, so a phone
// released after this catalogue was built still resolves its maker.
//
// This is deliberately the weaker claim. It says who built the device, not
// which device it is, and it says so from the shape of the code rather than
// from a record of that specific handset. Where the two disagree, the
// catalogue is right: sub-brands share their parent's code space, which is why
// "CPH" — OPPO's prefix — also covers OnePlus models.

// There is no separate prefix table here. The code spaces live in Rules, once,
// and BrandOf reads them through ResolveCode. A second copy would answer the
// same question from data that drifts: a rule added to Rules but not to the
// copy makes BrandOf and Resolve disagree about the same code, and nothing in
// the package would notice.

// BrandOf reports the manufacturer a model code belongs to, and whether the
// code's shape identified one.
//
// The catalogue is consulted first, because a recorded device is stronger
// evidence than a prefix. Only when the code is unknown do the shape rules
// decide, which is what lets a handset released after this release still
// resolve its maker.
//
// An unrecognised shape returns "" and false. It never falls back to the most
// common manufacturer, and never infers a brand from a code's length or
// character mix.
func BrandOf(code string) (string, bool) {
	if d, ok := Lookup(code); ok {
		return d.Brand, true
	}
	if r, ok := ResolveCode(code); ok {
		return r.Brand, true
	}
	// Consoles name themselves in prose and so are MatchContains rules, which
	// ResolveCode skips because a model code is not prose. When such prose
	// arrives in the device field anyway — "PlayStation 5" is the whole
	// identifier that request carries — the maker is still recoverable, so
	// they are tested here as prefixes of the code.
	for _, r := range Rules {
		if r.Match == MatchContains && strings.HasPrefix(code, r.Value) {
			return r.Brand, true
		}
	}
	return "", false
}
