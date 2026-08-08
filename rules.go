package adx

// Device rules, in Go. This file replaces data/device.yaml.
//
// Two tiers, and the distinction matters:
//
//  1. Named rules (priority 190+) map an exact token to its marketing name.
//     There is no derivable relationship between "SM-S928B" and "Galaxy S24
//     Ultra" — the manufacturer assigned it — so this tier is a lookup table
//     and nothing else. When a token is absent from it, Name stays empty. It
//     is never guessed.
//
//  2. Shape rules (priority 85..100) recognise a manufacturer from the form of
//     its model codes. These degrade gracefully: a phone released tomorrow
//     still resolves its brand even though tier 1 has never heard of it.
//
// The bulk of tier 1 — 12260 imported model codes — lives in
// internal/catalog/catalog_gen.go and is reached through Lookup. The rules
// here are the ones no published Android catalogue carries: Apple hardware,
// consoles, and the Pixel codes Google ships under their marketing name.

// MatchKind is how a rule compares itself against a model code or a
// User-Agent.
type MatchKind uint8

const (
	// MatchToken is an exact, case-sensitive comparison against the model
	// code.
	MatchToken MatchKind = iota

	// MatchPrefix recognises a manufacturer's code space. It claims the maker,
	// never the specific handset.
	MatchPrefix

	// MatchContains looks for the value anywhere in the User-Agent. It exists
	// for hardware that carries no model code at all: a console names itself
	// in prose.
	MatchContains
)

// DeviceType is the form factor a rule asserts.
type DeviceType string

// Form factors. A rule that cannot tell leaves Type empty rather than
// defaulting to Mobile.
const (
	TypeMobile  DeviceType = "Mobile"
	TypeTablet  DeviceType = "Tablet"
	TypeDesktop DeviceType = "Desktop"
	TypeConsole DeviceType = "Console"
)

// Rule is one device rule.
type Rule struct {
	// ID is stable and unique; it is what a changelog or a bug report names.
	ID string

	// Name is the marketing name. Empty on a shape rule, which is the whole
	// point of the tier split: recognising "SM-" says Samsung built it, not
	// which phone it is.
	Name string

	// Brand is the manufacturer, normalised to one spelling per company.
	Brand string

	// Model is the exact code this rule names. Empty on a shape rule.
	Model string

	// Family is the product line, when the rule identifies one.
	Family string

	// Type is the form factor, empty when the rule cannot tell.
	Type DeviceType

	// Priority orders the rules; higher is tried first.
	Priority int

	// Match is how the rule recognises a code.
	Match MatchKind

	// Value is what Match compares against.
	Value string

	// Confidence is how strong the claim is, 0..1.
	Confidence float64
}

// Rules is every hand-written device rule, ordered by descending priority so
// the first match is the most specific.
//
// Order within one priority is significant and hand-chosen: a longer prefix is
// listed before a shorter one it would otherwise shadow.
var Rules = []Rule{
	// -------------------------------------------------- named models (tier 1)
	//
	// This tier is deliberately almost empty. Every entry that asserts a name
	// the input does not already contain belongs in the imported catalogue,
	// where it cites a source. The Pixel entries below need no source because
	// they assert nothing new: Google ships the marketing name as the model
	// code, so the mapping is an identity.
	{ID: "pixel_9_pro_xl", Name: "Pixel 9 Pro XL", Brand: "Google", Model: "Pixel 9 Pro XL", Family: "Pixel", Type: TypeMobile, Priority: 240, Match: MatchToken, Value: "Pixel 9 Pro XL", Confidence: 0.98},
	{ID: "pixel_9_pro", Name: "Pixel 9 Pro", Brand: "Google", Model: "Pixel 9 Pro", Family: "Pixel", Type: TypeMobile, Priority: 239, Match: MatchToken, Value: "Pixel 9 Pro", Confidence: 0.98},
	{ID: "pixel_9", Name: "Pixel 9", Brand: "Google", Model: "Pixel 9", Family: "Pixel", Type: TypeMobile, Priority: 238, Match: MatchToken, Value: "Pixel 9", Confidence: 0.98},
	{ID: "pixel_8_pro", Name: "Pixel 8 Pro", Brand: "Google", Model: "Pixel 8 Pro", Family: "Pixel", Type: TypeMobile, Priority: 237, Match: MatchToken, Value: "Pixel 8 Pro", Confidence: 0.98},
	{ID: "pixel_8a", Name: "Pixel 8a", Brand: "Google", Model: "Pixel 8a", Family: "Pixel", Type: TypeMobile, Priority: 236, Match: MatchToken, Value: "Pixel 8a", Confidence: 0.98},
	{ID: "pixel_8", Name: "Pixel 8", Brand: "Google", Model: "Pixel 8", Family: "Pixel", Type: TypeMobile, Priority: 235, Match: MatchToken, Value: "Pixel 8", Confidence: 0.98},
	{ID: "pixel_7_pro", Name: "Pixel 7 Pro", Brand: "Google", Model: "Pixel 7 Pro", Family: "Pixel", Type: TypeMobile, Priority: 234, Match: MatchToken, Value: "Pixel 7 Pro", Confidence: 0.98},
	{ID: "pixel_7a", Name: "Pixel 7a", Brand: "Google", Model: "Pixel 7a", Family: "Pixel", Type: TypeMobile, Priority: 233, Match: MatchToken, Value: "Pixel 7a", Confidence: 0.98},
	{ID: "pixel_7", Name: "Pixel 7", Brand: "Google", Model: "Pixel 7", Family: "Pixel", Type: TypeMobile, Priority: 232, Match: MatchToken, Value: "Pixel 7", Confidence: 0.98},
	{ID: "pixel_6_pro", Name: "Pixel 6 Pro", Brand: "Google", Model: "Pixel 6 Pro", Family: "Pixel", Type: TypeMobile, Priority: 231, Match: MatchToken, Value: "Pixel 6 Pro", Confidence: 0.98},
	{ID: "pixel_6a", Name: "Pixel 6a", Brand: "Google", Model: "Pixel 6a", Family: "Pixel", Type: TypeMobile, Priority: 230, Match: MatchToken, Value: "Pixel 6a", Confidence: 0.98},
	{ID: "pixel_6", Name: "Pixel 6", Brand: "Google", Model: "Pixel 6", Family: "Pixel", Type: TypeMobile, Priority: 229, Match: MatchToken, Value: "Pixel 6", Confidence: 0.98},

	// ------------------------------------------------------------ Apple
	//
	// Apple publishes no model codes, so the token in the User-Agent is the
	// entire identifier a request carries. That is why these resolve a
	// manufacturer and a class of device, and never a specific handset: a
	// caller learns the request came from an iPhone, not which one.
	{ID: "iphone", Name: "iPhone", Brand: "Apple", Model: "iPhone", Type: TypeMobile, Priority: 200, Match: MatchToken, Value: "iPhone", Confidence: 0.97},
	{ID: "ipad", Name: "iPad", Brand: "Apple", Model: "iPad", Type: TypeTablet, Priority: 200, Match: MatchToken, Value: "iPad", Confidence: 0.97},
	{ID: "ipod", Name: "iPod touch", Brand: "Apple", Model: "iPod touch", Type: TypeMobile, Priority: 200, Match: MatchToken, Value: "iPod touch", Confidence: 0.95},

	// ------------------------------------------------- consoles and TVs
	{ID: "nintendo_switch", Name: "Nintendo Switch", Brand: "Nintendo", Model: "Nintendo Switch", Type: TypeConsole, Priority: 200, Match: MatchContains, Value: "Nintendo Switch", Confidence: 0.97},
	{ID: "playstation_5", Name: "PlayStation 5", Brand: "Sony", Model: "PlayStation 5", Type: TypeConsole, Priority: 200, Match: MatchContains, Value: "PlayStation 5", Confidence: 0.97},
	{ID: "playstation_4", Name: "PlayStation 4", Brand: "Sony", Model: "PlayStation 4", Type: TypeConsole, Priority: 200, Match: MatchContains, Value: "PlayStation 4", Confidence: 0.97},
	{ID: "xbox_console", Name: "Xbox", Brand: "Microsoft", Model: "Xbox", Type: TypeConsole, Priority: 190, Match: MatchContains, Value: "Xbox", Confidence: 0.95},

	// "Macintosh" names a class of machine and no model at all, so this rule
	// carries a Name and no Model. It sits last of the named rules because a
	// desktop token is the weakest evidence here.
	{ID: "macintosh", Name: "Mac", Brand: "Apple", Type: TypeDesktop, Priority: 150, Match: MatchToken, Value: "Macintosh", Confidence: 0.95},

	// ------------------------------------------------ brand shapes (tier 2)
	//
	// Brand only — never a model name.
	{ID: "samsung_prefix_sm", Brand: "Samsung", Priority: 100, Match: MatchPrefix, Value: "SM-", Confidence: 0.93},
	{ID: "samsung_prefix_sm_underscore", Brand: "Samsung", Priority: 100, Match: MatchPrefix, Value: "SM_", Confidence: 0.90},
	{ID: "samsung_prefix_sch", Brand: "Samsung", Priority: 100, Match: MatchPrefix, Value: "SCH-", Confidence: 0.90},
	{ID: "samsung_prefix_sph", Brand: "Samsung", Priority: 100, Match: MatchPrefix, Value: "SPH-", Confidence: 0.90},
	{ID: "samsung_prefix_sgh", Brand: "Samsung", Priority: 100, Match: MatchPrefix, Value: "SGH-", Confidence: 0.90},
	{ID: "samsung_prefix_gt", Brand: "Samsung", Priority: 100, Match: MatchPrefix, Value: "GT-", Confidence: 0.90},
	{ID: "samsung_prefix_sc", Brand: "Samsung", Priority: 100, Match: MatchPrefix, Value: "SC-", Confidence: 0.88},

	{ID: "google_prefix_pixel", Brand: "Google", Family: "Pixel", Priority: 100, Match: MatchPrefix, Value: "Pixel", Confidence: 0.95},

	{ID: "realme_prefix", Brand: "realme", Priority: 100, Match: MatchPrefix, Value: "RMX", Confidence: 0.93},

	{ID: "huawei_prefix_els", Brand: "Huawei", Priority: 100, Match: MatchPrefix, Value: "ELS-", Confidence: 0.92},
	{ID: "huawei_prefix_ana", Brand: "Huawei", Priority: 100, Match: MatchPrefix, Value: "ANA-", Confidence: 0.92},
	{ID: "huawei_prefix_vog", Brand: "Huawei", Priority: 100, Match: MatchPrefix, Value: "VOG-", Confidence: 0.92},
	{ID: "huawei_prefix_lya", Brand: "Huawei", Priority: 100, Match: MatchPrefix, Value: "LYA-", Confidence: 0.92},
	{ID: "huawei_prefix_noh", Brand: "Huawei", Priority: 100, Match: MatchPrefix, Value: "NOH-", Confidence: 0.92},

	{ID: "motorola_prefix_moto", Brand: "Motorola", Priority: 100, Match: MatchPrefix, Value: "moto ", Confidence: 0.93},

	{ID: "redmi_prefix", Brand: "Xiaomi", Family: "Redmi", Priority: 100, Match: MatchPrefix, Value: "Redmi", Confidence: 0.94},
	{ID: "poco_prefix", Brand: "Xiaomi", Family: "POCO", Priority: 100, Match: MatchPrefix, Value: "POCO", Confidence: 0.94},
	{ID: "mi_prefix", Brand: "Xiaomi", Priority: 100, Match: MatchPrefix, Value: "MI ", Confidence: 0.90},

	{ID: "lg_prefix_lm", Brand: "LG", Priority: 100, Match: MatchPrefix, Value: "LM-", Confidence: 0.90},
	{ID: "lg_prefix_lg", Brand: "LG", Priority: 100, Match: MatchPrefix, Value: "LG-", Confidence: 0.90},

	{ID: "oneplus_prefix_word", Brand: "OnePlus", Priority: 100, Match: MatchPrefix, Value: "ONEPLUS", Confidence: 0.93},
	{ID: "oneplus_prefix_kb2", Brand: "OnePlus", Priority: 100, Match: MatchPrefix, Value: "KB2", Confidence: 0.88},
	{ID: "oneplus_prefix_le2", Brand: "OnePlus", Priority: 100, Match: MatchPrefix, Value: "LE2", Confidence: 0.88},

	{ID: "sony_prefix", Brand: "Sony", Priority: 100, Match: MatchPrefix, Value: "XQ-", Confidence: 0.90},

	// Kindle Fire tablets. Amazon has used KF* since 2012 and ships nothing
	// else under it.
	{ID: "amazon_prefix", Brand: "Amazon", Priority: 100, Match: MatchPrefix, Value: "KF", Confidence: 0.88},

	{ID: "vivo_prefix", Brand: "vivo", Priority: 100, Match: MatchPrefix, Value: "vivo ", Confidence: 0.93},

	{ID: "asus_prefix", Brand: "Asus", Priority: 100, Match: MatchPrefix, Value: "ASUS_", Confidence: 0.93},
	{ID: "nokia_prefix", Brand: "Nokia", Priority: 100, Match: MatchPrefix, Value: "Nokia", Confidence: 0.93},
	{ID: "lenovo_prefix", Brand: "Lenovo", Priority: 100, Match: MatchPrefix, Value: "Lenovo", Confidence: 0.93},
	{ID: "tecno_prefix", Brand: "Tecno", Priority: 100, Match: MatchPrefix, Value: "TECNO", Confidence: 0.93},
	{ID: "infinix_prefix", Brand: "Infinix", Priority: 100, Match: MatchPrefix, Value: "Infinix", Confidence: 0.93},

	// CPH is OPPO's code space, and OnePlus — an OPPO sub-brand — ships codes
	// from it too. The prefix cannot separate them, so it reports the parent
	// brand at a confidence that says as much. Naming the sub-brand would need
	// the catalogue.
	{ID: "oppo_prefix_cph", Brand: "OPPO", Priority: 95, Match: MatchPrefix, Value: "CPH", Confidence: 0.80},
	{ID: "oppo_prefix_pbe", Brand: "OPPO", Priority: 95, Match: MatchPrefix, Value: "PBE", Confidence: 0.80},
	{ID: "oppo_prefix_pcl", Brand: "OPPO", Priority: 95, Match: MatchPrefix, Value: "PCL", Confidence: 0.80},

	{ID: "motorola_prefix_xt", Brand: "Motorola", Priority: 95, Match: MatchPrefix, Value: "XT", Confidence: 0.88},

	// Xiaomi's unbranded code spaces, tested last because they are the weakest
	// rules here. "M2" and a leading "22" are shapes Xiaomi has used
	// consistently, but neither is a reserved prefix the way "SM-" is: they are
	// a manufacturer's habit, not its namespace. Every other rule above gets to
	// decide first.
	{ID: "xiaomi_prefix_m2", Brand: "Xiaomi", Priority: 90, Match: MatchPrefix, Value: "M2", Confidence: 0.85},
	{ID: "xiaomi_numeric_prefix", Brand: "Xiaomi", Priority: 85, Match: MatchPrefix, Value: "22", Confidence: 0.80},
}
