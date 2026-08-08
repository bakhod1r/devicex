package adx_test

import (
	"testing"

	"github.com/bakhod1r/adx"
)

// The cases below were carried over verbatim from data/device.yaml, which used
// to hold the rules as data. Each one is the User-Agent token a rule claims to
// recognise, and the brand it claims to recognise it as.
func TestRulesResolveTheirOwnToken(t *testing.T) {
	for _, r := range adx.Rules {
		got, ok := adx.BrandOf(r.Value)
		if !ok {
			t.Errorf("%s: BrandOf(%q) found nothing; the rule recognises a token no shape rule covers", r.ID, r.Value)
			continue
		}
		if got != r.Brand {
			t.Errorf("%s: BrandOf(%q) = %q, want %q", r.ID, r.Value, got, r.Brand)
		}
	}
}

func TestRulesAreWellFormed(t *testing.T) {
	seen := make(map[string]bool, len(adx.Rules))
	prev := 1 << 30

	for _, r := range adx.Rules {
		if r.ID == "" {
			t.Errorf("rule for %q has no ID", r.Value)
		}
		if seen[r.ID] {
			t.Errorf("%s: duplicate ID", r.ID)
		}
		seen[r.ID] = true

		if r.Value == "" {
			t.Errorf("%s: matches nothing", r.ID)
		}
		if r.Brand == "" {
			t.Errorf("%s: asserts no brand", r.ID)
		}
		// A shape rule with a name would claim to identify a handset it cannot
		// pin down. A token rule may name a class of machine without a model
		// code — "Macintosh" is one — so only Model-less shape rules are wrong.
		if r.Name != "" && r.Model == "" && r.Match == adx.MatchPrefix {
			t.Errorf("%s: has Name %q but no Model", r.ID, r.Name)
		}
		if r.Confidence <= 0 || r.Confidence > 1 {
			t.Errorf("%s: confidence %v is outside (0,1]", r.ID, r.Confidence)
		}
		if r.Priority > prev {
			t.Errorf("%s: priority %d is above the preceding rule's %d; Rules must be sorted descending", r.ID, r.Priority, prev)
		}
		prev = r.Priority
	}
}

// Shape rules must not name a device. Recognising "SM-" says Samsung built the
// handset, not which handset it is.
func TestShapeRulesNameNothing(t *testing.T) {
	for _, r := range adx.Rules {
		if r.Match == adx.MatchPrefix && r.Name != "" {
			t.Errorf("%s: shape rule names a device (%q)", r.ID, r.Name)
		}
	}
}
