package main

// The static lookup API.
//
// GitHub Pages serves files and runs nothing, so the API is the files: the
// catalogue is split into shards named after the first two characters of a
// model code, and a client that wants "SM-S928B" fetches "api/SM.json" and
// reads one key out of it. That is a real endpoint — curl it, cache it, embed
// it — and it costs a single request of a few kilobytes rather than the whole
// 28000-row catalogue.
//
// The shard key is deliberately trivial, because it is implemented twice: here
// and in the JavaScript that reads it. Anything cleverer would be two
// implementations to keep in step.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bakhod1r/devicex"
)

// shardKey is the shard a code lives in: its first two characters, uppercased,
// with anything that is not a letter or a digit replaced by an underscore so
// the key is always a safe filename. Codes shorter than two characters are
// dropped on import, so every code has one.
//
// This function is mirrored in api/devicex.js. Change one, change both.
func shardKey(code string) string {
	r := []rune(code)
	out := make([]rune, 0, 2)
	for i := 0; i < 2 && i < len(r); i++ {
		c := r[i]
		switch {
		case c >= 'a' && c <= 'z':
			out = append(out, c-32)
		case c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
			out = append(out, c)
		default:
			out = append(out, '_')
		}
	}
	return string(out)
}

// webMeta is api/meta.json: what the shards were built from. It is the same
// provenance the Go package reports through Source, Generated and Version, for
// a caller that reaches the data over HTTP instead.
type webMeta struct {
	Generated string   `json:"generated"`
	Source    string   `json:"source"`
	Devices   int      `json:"devices"`
	Shards    []string `json:"shards"`
	Notice    string   `json:"notice"`
}

// writeWeb writes the shards, the rule table and the provenance into dir.
//
// One shard is a flat object of code to [brand, name] — a two-element array
// rather than an object, because the keys would otherwise be repeated 28000
// times and triple the transfer for nothing.
func writeWeb(dir, source, dated string, entries []Entry) error {
	shards := map[string]map[string][2]string{}
	for _, e := range entries {
		k := shardKey(e.Code)
		if shards[k] == nil {
			shards[k] = map[string][2]string{}
		}
		shards[k][e.Code] = [2]string{e.Brand, e.Name}
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	keys := make([]string, 0, len(shards))
	for k := range shards {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if err := writeJSON(filepath.Join(dir, k+".json"), shards[k]); err != nil {
			return err
		}
	}

	// The shape rules travel with the data. Without them the page would answer
	// nothing for a code the catalogue does not hold, while the Go package
	// answers the manufacturer — two APIs over one dataset disagreeing.
	type webRule struct {
		ID         string  `json:"id"`
		Name       string  `json:"name,omitempty"`
		Brand      string  `json:"brand"`
		Model      string  `json:"model,omitempty"`
		Family     string  `json:"family,omitempty"`
		Type       string  `json:"type,omitempty"`
		Match      string  `json:"match"`
		Value      string  `json:"value"`
		Confidence float64 `json:"confidence"`
	}
	match := map[devicex.MatchKind]string{
		devicex.MatchToken:    "token",
		devicex.MatchPrefix:   "prefix",
		devicex.MatchContains: "contains",
	}
	rules := make([]webRule, 0, len(devicex.Rules))
	for _, r := range devicex.Rules {
		rules = append(rules, webRule{
			ID: r.ID, Name: r.Name, Brand: r.Brand, Model: r.Model,
			Family: r.Family, Type: string(r.Type),
			Match: match[r.Match], Value: r.Value, Confidence: r.Confidence,
		})
	}
	if err := writeJSON(filepath.Join(dir, "rules.json"), rules); err != nil {
		return err
	}

	meta := webMeta{
		Generated: dated,
		Source:    source,
		Devices:   len(entries),
		Shards:    keys,
		Notice:    "See NOTICE.md. The catalogue is CC BY-NC-SA 4.0 where it draws on KHwang9883/MobileModels.",
	}
	if err := writeJSON(filepath.Join(dir, "meta.json"), meta); err != nil {
		return err
	}

	stale, err := staleShards(dir, keys)
	if err != nil {
		return err
	}
	for _, f := range stale {
		fmt.Printf("%s: stale, no code shards here any more — delete it\n", f)
	}

	fmt.Printf("%s: %d shards, %d rules\n", dir, len(keys), len(rules))
	return nil
}

func writeJSON(path string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

// staleShards reports shard files in dir that this build did not write. A code
// space that empties out — a page removed upstream, a brand dropped — would
// otherwise keep serving its last known answer forever, which is the one
// failure mode a static API has and a server does not.
func staleShards(dir string, written []string) ([]string, error) {
	keep := make(map[string]bool, len(written)+2)
	for _, k := range written {
		keep[k+".json"] = true
	}
	keep["rules.json"] = true
	keep["meta.json"] = true

	found, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var stale []string
	for _, f := range found {
		if !strings.HasSuffix(f.Name(), ".json") || keep[f.Name()] {
			continue
		}
		stale = append(stale, filepath.Join(dir, f.Name()))
	}
	return stale, nil
}
