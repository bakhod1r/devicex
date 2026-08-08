package adx

import "strings"

// Code extracts the model code from the device field of an Android
// User-Agent.
//
// Android writes the build fingerprint after the model, separated by a space:
//
//	"SM-A546E Build/UP1A.231005.007" -> "SM-A546E"
//
// The suffix is a build identifier, not part of the code, and leaving it on
// turns every catalogue lookup into a miss. Stripping it is a fact about the
// shape of a model code, which is this package's subject, so it lives here
// rather than being re-derived by each parser.
//
// A field with no build suffix is returned trimmed and otherwise unchanged.
// Codes containing spaces — "moto g play (2023)", "Pixel 9 Pro XL" — survive,
// because only a " Build/" boundary is cut.
func Code(field string) string {
	field = strings.TrimSpace(field)
	if i := strings.Index(field, " Build/"); i >= 0 {
		field = field[:i]
	}
	// Some devices ship the separator without the space.
	if i := strings.Index(field, "Build/"); i == 0 {
		return ""
	}
	return strings.TrimSpace(field)
}
