package adx

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

type prefixBrand struct {
	prefix string
	brand  string
}

// brandPrefixes is ordered longest-first within each first byte, so a more
// specific prefix is tested before a more general one.
var brandPrefixes = []prefixBrand{
	{"SM-", "Samsung"},
	{"SM_", "Samsung"},
	{"GT-", "Samsung"},
	{"SCH-", "Samsung"},
	{"SPH-", "Samsung"},
	{"SGH-", "Samsung"},
	{"SC-", "Samsung"},

	{"CPH", "OPPO"},
	{"PBE", "OPPO"},
	{"PCL", "OPPO"},

	{"RMX", "realme"},

	{"Pixel", "Google"},

	{"moto ", "Motorola"},
	{"XT", "Motorola"},

	{"Redmi", "Xiaomi"},
	{"POCO", "Xiaomi"},
	{"MI ", "Xiaomi"},

	{"ELS-", "Huawei"},
	{"ANA-", "Huawei"},
	{"VOG-", "Huawei"},
	{"LYA-", "Huawei"},
	{"NOH-", "Huawei"},

	{"LM-", "LG"},
	{"LG-", "LG"},

	{"ONEPLUS", "OnePlus"},
	{"KB2", "OnePlus"},
	{"LE2", "OnePlus"},

	{"XQ-", "Sony"},

	// Kindle Fire tablets. Amazon has used KF* since 2012 and ships nothing
	// else under it.
	{"KF", "Amazon"},

	{"vivo ", "vivo"},

	{"ASUS_", "Asus"},
	{"Nokia", "Nokia"},
	{"Lenovo", "Lenovo"},
	{"TECNO", "Tecno"},
	{"Infinix", "Infinix"},

	// Hardware that cannot appear in an Android catalogue but whose tokens
	// appear in User-Agents constantly. These are whole tokens rather than
	// allocated prefixes: Apple and the console makers publish no model codes,
	// so "iPhone" is the entire identifier a request carries. That is why they
	// resolve a manufacturer and never a model — a caller learns that a
	// request came from an iPhone, not which one.
	{"iPhone", "Apple"},
	{"iPad", "Apple"},
	{"iPod", "Apple"},
	{"Macintosh", "Apple"},
	{"Nintendo Switch", "Nintendo"},
	{"PlayStation", "Sony"},
	{"Xbox", "Microsoft"},

	// Xiaomi's unbranded code spaces, tested last because they are the weakest
	// rules here. "M2" and a leading "22" are shapes Xiaomi has used
	// consistently, but neither is a reserved prefix the way "SM-" is: they
	// are a manufacturer's habit, not its namespace. Every other rule above
	// gets to decide first.
	{"M2", "Xiaomi"},
	{"22", "Xiaomi"},
}

// BrandOf reports the manufacturer a model code belongs to, and whether the
// code's shape identified one.
//
// The catalogue is consulted first, because a recorded device is stronger
// evidence than a prefix. Only when the code is unknown does the shape decide,
// which is what lets a handset released after this release still resolve its
// maker.
//
// An unrecognised shape returns "" and false. It never falls back to the most
// common manufacturer, and never infers a brand from a code's length or
// character mix.
func BrandOf(code string) (string, bool) {
	if d, ok := Lookup(code); ok {
		return d.Brand, true
	}
	for _, p := range brandPrefixes {
		if strings.HasPrefix(code, p.prefix) {
			return p.brand, true
		}
	}
	return "", false
}
