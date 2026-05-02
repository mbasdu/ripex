package whitelist

import (
	"context"
	"encoding/json"
	"net/netip"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var primaryWhitelistFixturePaths = []string{
	filepath.Join("..", "..", "testdata", "russia_mobile_whitelist_mintsifry_2025_09_05.json"),
	filepath.Join("..", "..", "testdata", "russia_mobile_whitelist_mintsifry_2025_11_14.json"),
	filepath.Join("..", "..", "testdata", "russia_mobile_whitelist_mintsifry_2025_12_05.json"),
	filepath.Join("..", "..", "testdata", "russia_mobile_whitelist_mintsifry_2025_12_18.json"),
	filepath.Join("..", "..", "testdata", "russia_mobile_whitelist_mintsifry_2026_02_04.json"),
}

type fakeResolver map[string][]netip.Addr

func (f fakeResolver) LookupNetIP(_ context.Context, _ string, host string) ([]netip.Addr, error) {
	return f[host], nil
}

func TestCheckCoverage(t *testing.T) {
	prefixes := []netip.Prefix{
		netip.MustParsePrefix("10.0.0.0/23"),
		netip.MustParsePrefix("192.0.2.0/24"),
	}
	entries := []Entry{
		{Name: "covered", Domain: "covered.example"},
		{Name: "partial", Domain: "partial.example"},
	}
	resolver := fakeResolver{
		"covered.example": {netip.MustParseAddr("10.0.0.10")},
		"partial.example": {netip.MustParseAddr("10.0.1.20"), netip.MustParseAddr("198.51.100.5")},
	}

	results, err := CheckCoverage(context.Background(), resolver, prefixes, entries)
	if err != nil {
		t.Fatalf("CheckCoverage() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("results len = %d", len(results))
	}
	if !reflect.DeepEqual(results[0].Covered, []netip.Addr{netip.MustParseAddr("10.0.0.10")}) {
		t.Fatalf("covered result = %#v", results[0].Covered)
	}
	if !reflect.DeepEqual(results[1].Missing, []netip.Addr{netip.MustParseAddr("198.51.100.5")}) {
		t.Fatalf("missing result = %#v", results[1].Missing)
	}
	if AllCovered(results) {
		t.Fatalf("AllCovered() = true, want false")
	}
}

func TestLoadEntriesFromPaths(t *testing.T) {
	dir := t.TempDir()

	first := filepath.Join(dir, "first.json")
	second := filepath.Join(dir, "second.json")

	writeEntriesFixture(t, first, []Entry{
		{Name: "one", Domain: "one.example", Category: "test"},
	})
	writeEntriesFixture(t, second, []Entry{
		{Name: "two", Domain: "two.example", Category: "test"},
	})

	got, err := LoadEntriesFromPaths(first, second)
	if err != nil {
		t.Fatalf("LoadEntriesFromPaths() error = %v", err)
	}

	want := []Entry{
		{Name: "one", Domain: "one.example", Category: "test"},
		{Name: "two", Domain: "two.example", Category: "test"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LoadEntriesFromPaths() = %#v, want %#v", got, want)
	}
}

func TestLoadEntriesFromPathsDuplicateDomain(t *testing.T) {
	dir := t.TempDir()

	first := filepath.Join(dir, "first.json")
	second := filepath.Join(dir, "second.json")

	writeEntriesFixture(t, first, []Entry{
		{Name: "one", Domain: "dup.example", Category: "test"},
	})
	writeEntriesFixture(t, second, []Entry{
		{Name: "two", Domain: "dup.example", Category: "test"},
	})

	if _, err := LoadEntriesFromPaths(first, second); err == nil {
		t.Fatal("LoadEntriesFromPaths() error = nil, want duplicate-domain error")
	}
}

func TestWhitelistFixturesLoadWithoutDuplicateDomains(t *testing.T) {
	entries, err := LoadEntriesFromPaths(primaryWhitelistFixturePaths...)
	if err != nil {
		t.Fatalf("LoadEntriesFromPaths() error = %v", err)
	}
	if len(entries) < 80 {
		t.Fatalf("fixture entry count = %d, want at least 80", len(entries))
	}
	for _, entry := range entries {
		if entry.Name == "" {
			t.Fatalf("entry for domain %q has empty name", entry.Domain)
		}
		if entry.Domain == "" {
			t.Fatal("entry has empty domain")
		}
		if entry.Category == "" {
			t.Fatalf("entry for domain %q has empty category", entry.Domain)
		}
		if len(entry.Sources) == 0 {
			t.Fatalf("entry for domain %q has no sources", entry.Domain)
		}
	}
}

func writeEntriesFixture(t *testing.T, path string, entries []Entry) {
	t.Helper()

	data, err := json.Marshal(entries)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
}
