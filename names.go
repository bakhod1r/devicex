package devicex

// Names resolves a model code for a caller that only wants the two strings.
//
// The signature is deliberately plain — no types from this package appear in
// it — so a consumer can accept it as a function value without importing devicex
// at all:
//
//	// in the consumer, with no dependency on devicex:
//	type Config struct {
//	    DeviceNames func(code string) (name, brand string, ok bool)
//	}
//
//	// in the application, which imports both:
//	p := uax.New(uax.Config{DeviceNames: devicex.Names})
//
// That keeps the catalogue and its consumers independent in both directions:
// neither module imports the other, and an application that does not want
// 25,000 device names does not link them.
//
// It allocates nothing: the returned strings are compile-time constants.
func Names(code string) (name, brand string, ok bool) {
	d, found := Lookup(code)
	if !found {
		return "", "", false
	}
	return d.Name, d.Brand, true
}

// Describe is Names with everything the rules know, in the same
// dependency-free shape.
//
// Names only reaches the catalogue, which makes it silent about most of what
// this package can answer: a handset newer than the catalogue resolves no
// brand, an iPhone resolves nothing at all, and a catalogued Galaxy Tab loses
// the fact that it is a tablet. Describe is Resolve behind a signature that
// mentions no type from this package, so a consumer can still accept it as a
// function value without importing devicex:
//
//	type Config struct {
//	    Device func(code, ua string) (name, brand, model, family, deviceType string, confidence float64, ok bool)
//	}
//
//	p := uax.New(uax.Config{Device: devicex.Describe})
//
// A caller willing to import devicex should call Resolve instead and read the
// Rule, which additionally carries the rule ID that produced the answer — the
// thing a bug report needs to name. This flat form exists for the caller who
// will not.
//
// deviceType is one of "Mobile", "Tablet", "Desktop", "Console", or empty when
// no rule could tell. It is never defaulted to "Mobile".
//
// model is the exact code the answer is about, and is empty when the rule
// names a class of machine rather than a model. That distinction is carried
// here rather than left to the caller because it cannot be recovered
// downstream: "iPhone" is both the name and the model an iPhone request
// carries, while "Macintosh" names a class and no model at all, and any
// heuristic that copies name into model when the User-Agent contains it
// invents a "Mac" model for every desktop.
func Describe(code, ua string) (name, brand, model, family, deviceType string, confidence float64, ok bool) {
	r, found := Resolve(code, ua)
	if !found {
		return "", "", "", "", "", 0, false
	}
	return r.Name, r.Brand, r.Model, r.Family, string(r.Type), r.Confidence, true
}
