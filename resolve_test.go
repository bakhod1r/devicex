package devicex

import "testing"

func TestResolveCatalogueWins(t *testing.T) {
	r, ok := Resolve("SM-G973F", "")
	if !ok {
		t.Fatal("SM-G973F: no answer")
	}
	if r.ID != "catalog" {
		t.Errorf("ID = %q, want catalog", r.ID)
	}
	if r.Name != "Galaxy S10" || r.Brand != "Samsung" {
		t.Errorf("got %q/%q, want Galaxy S10/Samsung", r.Name, r.Brand)
	}
}

// A catalogued tablet keeps its name and gains the form factor the catalogue
// does not record. This is the whole reason Resolve consults both tiers.
func TestResolveCatalogueGainsFormFactor(t *testing.T) {
	r, ok := Resolve("SM-T870", "")
	if !ok {
		t.Fatal("SM-T870: no answer")
	}
	if r.Name != "Galaxy Tab S7" || r.Brand != "Samsung" {
		t.Errorf("got %q/%q, want Galaxy Tab S7/Samsung", r.Name, r.Brand)
	}
	if r.Type != TypeTablet {
		t.Errorf("Type = %q, want Tablet: the catalogue records no form factor, the shape rule does", r.Type)
	}
}

// A tablet code the catalogue does not hold. The catalogue records no form
// factor, so this is the shape rule's job: SM-X is a 2022-and-later Galaxy Tab
// space, and the tablet claim has to survive the name being unknown.
func TestResolveUncataloguedTablet(t *testing.T) {
	r, ok := Resolve("SM-X999Z", "")
	if !ok {
		t.Fatal("SM-X999Z: no answer")
	}
	if r.Type != TypeTablet || r.Brand != "Samsung" {
		t.Errorf("got %q/%q, want Samsung/Tablet", r.Brand, r.Type)
	}
	if r.Name != "" {
		t.Errorf("Name = %q, want empty", r.Name)
	}
}

// A catalogued tablet keeps its name and gains the form factor the catalogue
// does not carry.
func TestResolveCataloguedTabletKeepsBothAnswers(t *testing.T) {
	r, ok := Resolve("SM-X710", "")
	if !ok {
		t.Fatal("SM-X710: no answer")
	}
	if r.Name != "Galaxy Tab S9" || r.Brand != "Samsung" || r.Type != TypeTablet {
		t.Errorf("got %q/%q/%q, want Galaxy Tab S9/Samsung/Tablet", r.Name, r.Brand, r.Type)
	}
}

// The point of tier 2: a code no catalogue has ever seen still resolves a
// maker, and resolves no name.
func TestResolveUnknownCodeFallsToShape(t *testing.T) {
	r, ok := Resolve("SM-S999Z", "")
	if !ok {
		t.Fatal("SM-S999Z: no answer")
	}
	if r.Brand != "Samsung" {
		t.Errorf("Brand = %q, want Samsung", r.Brand)
	}
	if r.Name != "" {
		t.Errorf("Name = %q, want empty: a shape rule must not name a handset", r.Name)
	}
}

func TestResolveCodeSkipsContainsRules(t *testing.T) {
	if r, ok := ResolveCode("PlayStation 5"); ok {
		t.Errorf("ResolveCode matched a MatchContains rule: %q", r.ID)
	}
}

func TestResolveUA(t *testing.T) {
	for _, tt := range []struct {
		ua        string
		wantBrand string
		wantName  string
		wantType  DeviceType
	}{
		{"Mozilla/5.0 (PlayStation; PlayStation 5/9.60) AppleWebKit/605.1.15", "Sony", "PlayStation 5", TypeConsole},
		{"Mozilla/5.0 (Nintendo Switch; WifiWebAuthApplet)", "Nintendo", "Nintendo Switch", TypeConsole},
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)", "Apple", "Mac", TypeDesktop},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X)", "Apple", "iPhone", TypeMobile},
		{"Mozilla/5.0 (iPad; CPU OS 17_5 like Mac OS X)", "Apple", "iPad", TypeTablet},
	} {
		r, ok := ResolveUA(tt.ua)
		if !ok {
			t.Errorf("%q: no answer", tt.ua)
			continue
		}
		if r.Brand != tt.wantBrand || r.Name != tt.wantName || r.Type != tt.wantType {
			t.Errorf("%q: got %q/%q/%q, want %q/%q/%q",
				tt.ua, r.Brand, r.Name, r.Type, tt.wantBrand, tt.wantName, tt.wantType)
		}
	}
}

// An iPhone User-Agent contains "Mac OS X", so the Macintosh rule would claim
// it if priority were ignored. This is the ordering guarantee callers would
// have to reproduce if they walked Rules themselves.
func TestResolveUAOrdering(t *testing.T) {
	r, ok := ResolveUA("Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/605.1.15")
	if !ok || r.ID != "iphone" {
		t.Errorf("got %q, want iphone", r.ID)
	}
	r, ok = ResolveUA("Mozilla/5.0 (Linux; Android 15; Pixel 9 Pro XL) AppleWebKit/537.36")
	if !ok || r.ID != "pixel_9_pro_xl" {
		t.Errorf("got %q, want pixel_9_pro_xl", r.ID)
	}
}

func TestResolveNothing(t *testing.T) {
	if _, ok := Resolve("", ""); ok {
		t.Error("empty input resolved something")
	}
	if _, ok := Resolve("ZZZZ-not-a-code", "Mozilla/5.0 (Linux; Android 15)"); ok {
		t.Error("unknown code and unknown User-Agent resolved something")
	}
}

// Rules is documented as ordered by descending priority, and both resolvers
// depend on it. A rule added out of order would silently shadow a better one.
func TestRulesOrderedByPriority(t *testing.T) {
	for i := 1; i < len(Rules); i++ {
		if Rules[i].Priority > Rules[i-1].Priority {
			t.Fatalf("rule %q (priority %d) follows %q (priority %d)",
				Rules[i].ID, Rules[i].Priority, Rules[i-1].ID, Rules[i-1].Priority)
		}
	}
}

func TestRuleIDsUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, r := range Rules {
		if seen[r.ID] {
			t.Errorf("duplicate rule ID %q", r.ID)
		}
		seen[r.ID] = true
	}
}

func TestCode(t *testing.T) {
	for _, tt := range []struct{ in, want string }{
		{"SM-A546E Build/UP1A.231005.007", "SM-A546E"},
		{"SM-A546E", "SM-A546E"},
		{"  SM-A546E  ", "SM-A546E"},
		{"moto g play (2023) Build/TPBS33.2", "moto g play (2023)"},
		{"Pixel 9 Pro XL", "Pixel 9 Pro XL"},
		{"Build/UP1A.231005.007", ""},
		{"", ""},
	} {
		if got := Code(tt.in); got != tt.want {
			t.Errorf("Code(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// Code feeds Lookup: the suffix is exactly what turns a hit into a miss.
func TestCodeFeedsLookup(t *testing.T) {
	if _, ok := Lookup("SM-G973F Build/QP1A.190711.020"); ok {
		t.Fatal("raw field matched the catalogue; the test is not testing anything")
	}
	if _, ok := Lookup(Code("SM-G973F Build/QP1A.190711.020")); !ok {
		t.Error("Code() output did not match the catalogue")
	}
}

func TestDescribe(t *testing.T) {
	name, brand, model, _, _, conf, ok := Describe("SM-G973F", "")
	if !ok || name != "Galaxy S10" || brand != "Samsung" || model != "SM-G973F" {
		t.Errorf("got %q/%q/%q ok=%v, want Galaxy S10/Samsung/SM-G973F", name, brand, model, ok)
	}
	if conf <= 0 {
		t.Errorf("confidence = %v, want > 0", conf)
	}

	// The cases Names cannot answer at all.
	if _, brand, _, _, _, _, ok := Describe("SM-S999Z", ""); !ok || brand != "Samsung" {
		t.Errorf("unknown Samsung code: brand = %q ok = %v", brand, ok)
	}
	if name, _, model, _, typ, _, ok := Describe("", "Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X)"); !ok || name != "iPhone" || model != "iPhone" || typ != "Mobile" {
		t.Errorf("iPhone: %q/%q/%q ok = %v", name, model, typ, ok)
	}
	if _, _, _, _, typ, _, ok := Describe("SM-T870", ""); !ok || typ != "Tablet" {
		t.Errorf("Galaxy Tab: type = %q ok = %v", typ, ok)
	}

	// "Macintosh" names a class and no model. The empty model is the claim —
	// a flat signature that dropped it would leave a consumer inventing one.
	if name, _, model, _, typ, _, ok := Describe("", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"); !ok || name != "Mac" || model != "" || typ != "Desktop" {
		t.Errorf("Macintosh: %q/%q/%q ok = %v", name, model, typ, ok)
	}
	if _, _, model, _, _, _, ok := Describe("", "Mozilla/5.0 (PlayStation; PlayStation 5/8.00)"); !ok || model != "PlayStation 5" {
		t.Errorf("PlayStation 5: model = %q ok = %v", model, ok)
	}

	if _, _, _, _, _, _, ok := Describe("", ""); ok {
		t.Error("Describe answered on empty input")
	}
}
