package adx

// Names resolves a model code for a caller that only wants the two strings.
//
// The signature is deliberately plain — no types from this package appear in
// it — so a consumer can accept it as a function value without importing adx
// at all:
//
//	// in the consumer, with no dependency on adx:
//	type Config struct {
//	    DeviceNames func(code string) (name, brand string, ok bool)
//	}
//
//	// in the application, which imports both:
//	p := uax.New(uax.Config{DeviceNames: adx.Names})
//
// That keeps the catalogue and its consumers independent in both directions:
// neither module imports the other, and an application that does not want
// 12,000 device names does not link them.
//
// It allocates nothing: the returned strings are compile-time constants.
func Names(code string) (name, brand string, ok bool) {
	d, found := Lookup(code)
	if !found {
		return "", "", false
	}
	return d.Name, d.Brand, true
}
