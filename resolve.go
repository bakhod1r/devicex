package devicex

import "strings"

// Resolving a User-Agent's device evidence.
//
// Rules is exported data, and reading it is a legitimate thing to do — a
// caller auditing what this package claims should be able to see every claim.
// Evaluating it is a different matter. The order within Rules is significant
// and hand-chosen: rules are listed by descending priority, and within one
// priority a longer prefix precedes the shorter one it would otherwise shadow.
// A caller who re-implements the walk has to reproduce that, and will diverge
// from it the first time a rule is added here.
//
// So the evaluation lives with the data. Three entry points, because a model
// code and a User-Agent are different evidence and the rules that read them do
// not overlap:
//
//	ResolveCode  a model code, "SM-S928B"
//	ResolveUA    the whole User-Agent, for hardware that carries no code
//	Resolve      both, catalogue first — what most callers want

// ResolveCode resolves a model code against the rules.
//
// It reads the shape of the code and nothing else, so it answers for hardware
// the catalogue has never heard of: a Samsung handset released after this
// release still resolves its maker from "SM-". It does not consult the
// catalogue — use Resolve for that.
//
// MatchContains rules are skipped. They describe prose in a User-Agent
// ("PlayStation 5"), and a model code is not prose.
func ResolveCode(code string) (Rule, bool) {
	if code == "" {
		return Rule{}, false
	}
	for _, r := range Rules {
		switch r.Match {
		case MatchToken:
			if r.Value == code {
				return r, true
			}
		case MatchPrefix:
			if strings.HasPrefix(code, r.Value) {
				return r, true
			}
		}
	}
	return Rule{}, false
}

// ResolveUA resolves a whole User-Agent against the rules.
//
// It exists for hardware that carries no model code at all and names itself in
// prose instead: "Macintosh", "PlayStation 5", "Nintendo Switch". Those tokens
// are the entire identifier such a request carries.
//
// Both MatchContains and MatchToken rules are tested, because a named token is
// a literal substring of the User-Agent that carries it. Prefix rules are not:
// "SM-" appears in a User-Agent only as part of a model code, and matching it
// against the whole string would claim Samsung for any text containing those
// three characters.
//
// Rule order decides overlaps, which is why "Pixel 9 Pro XL" is listed above
// "Pixel 9" — the first match on a Pixel 9 Pro XL User-Agent is the specific
// one.
func ResolveUA(ua string) (Rule, bool) {
	if ua == "" {
		return Rule{}, false
	}
	for _, r := range Rules {
		switch r.Match {
		case MatchContains, MatchToken:
			if strings.Contains(ua, r.Value) {
				return r, true
			}
		}
	}
	return Rule{}, false
}

// Resolve answers everything the package knows about one request's device
// evidence: the model code, and the User-Agent it came out of.
//
// Either argument may be empty. A code alone is the Android case; a User-Agent
// alone covers the hardware that publishes no code.
//
// The catalogue decides first, because a recorded device is stronger evidence
// than a code's shape — it names the specific handset, which no rule can. When
// it answers, the rules are still consulted for what a catalogue row does not
// carry: the form factor and the product line. That is how a Galaxy Tab
// resolves as a tablet while keeping its catalogued name.
//
// The returned Rule for a catalogued device has ID "catalog" and cites Source.
// A code the catalogue does not hold falls through to the shape rules, and then
// to the User-Agent. Nothing is guessed at any step: an unrecognised code and
// an unrecognised User-Agent return false.
func Resolve(code, ua string) (Rule, bool) {
	if d, ok := Lookup(code); ok {
		out := Rule{
			ID:         "catalog",
			Name:       d.Name,
			Brand:      d.Brand,
			Model:      d.Code,
			Priority:   250,
			Match:      MatchToken,
			Value:      d.Code,
			Confidence: 0.99,
		}
		// The catalogue holds no form factor and no product line. A shape rule
		// may, and cannot contradict the name: it never asserts one.
		if r, ok := ResolveCode(code); ok {
			out.Type = r.Type
			out.Family = r.Family
		}
		return out, true
	}
	if r, ok := ResolveCode(code); ok {
		return r, true
	}
	return ResolveUA(ua)
}
