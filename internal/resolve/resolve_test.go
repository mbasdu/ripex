package resolve

import (
	"context"
	"errors"
	"net/netip"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

type fakeResolver struct {
	answers map[string][]netip.Addr
	errors  map[string]error
}

func (f fakeResolver) LookupNetIP(_ context.Context, network, host string) ([]netip.Addr, error) {
	if network != "ip4" {
		return nil, errors.New("unexpected network")
	}
	if err, ok := f.errors[host]; ok {
		return nil, err
	}
	return f.answers[host], nil
}

func TestLoadDomains(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "domains.lst")
	content := `# comment line
example.com
Example.COM
  trailing-whitespace.com

# another comment
ozon.ru
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadDomains(path)
	if err != nil {
		t.Fatalf("LoadDomains() error = %v", err)
	}
	want := []string{"example.com", "trailing-whitespace.com", "ozon.ru"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LoadDomains() = %#v, want %#v", got, want)
	}
}

func TestResolveHappyPath(t *testing.T) {
	resolver := fakeResolver{
		answers: map[string][]netip.Addr{
			"a.example": {netip.MustParseAddr("192.0.2.1"), netip.MustParseAddr("192.0.2.2")},
			"b.example": {netip.MustParseAddr("198.51.100.5")},
		},
	}
	records := Resolve(context.Background(), resolver,
		[]string{"a.example", "b.example"},
		Config{Concurrency: 2, Timeout: time.Second})

	if len(records) != 2 {
		t.Fatalf("len(records) = %d, want 2", len(records))
	}
	if records[0].Domain != "a.example" || len(records[0].Addrs) != 2 {
		t.Fatalf("record[0] = %#v", records[0])
	}
	if records[1].Domain != "b.example" || records[1].Addrs[0].String() != "198.51.100.5" {
		t.Fatalf("record[1] = %#v", records[1])
	}
}

func TestResolveIPv6Ignored(t *testing.T) {
	resolver := fakeResolver{
		answers: map[string][]netip.Addr{
			"x.example": {
				netip.MustParseAddr("2001:db8::1"),
				netip.MustParseAddr("192.0.2.10"),
			},
		},
	}
	records := Resolve(context.Background(), resolver,
		[]string{"x.example"}, Config{})
	if len(records[0].Addrs) != 1 || !records[0].Addrs[0].Is4() {
		t.Fatalf("IPv6 addr was not filtered: %#v", records[0].Addrs)
	}
}

func TestResolveErrorSurfaces(t *testing.T) {
	resolver := fakeResolver{
		errors: map[string]error{"gone.example": errors.New("nxdomain")},
	}
	records := Resolve(context.Background(), resolver,
		[]string{"gone.example"}, Config{})
	if records[0].Err == nil {
		t.Fatal("expected Err to be non-nil")
	}
	if err := FailuresError(records); err == nil {
		t.Fatal("expected FailuresError to be non-nil")
	}
}

func TestToPrefixesDedup(t *testing.T) {
	records := []Record{
		{Domain: "a", Addrs: []netip.Addr{netip.MustParseAddr("192.0.2.5"), netip.MustParseAddr("192.0.2.5")}},
		{Domain: "b", Addrs: []netip.Addr{netip.MustParseAddr("192.0.2.5"), netip.MustParseAddr("198.51.100.1")}},
	}
	prefixes := ToPrefixes(records)
	if len(prefixes) != 2 {
		t.Fatalf("len(prefixes) = %d, want 2", len(prefixes))
	}
	if prefixes[0].String() != "192.0.2.5/32" || prefixes[1].String() != "198.51.100.1/32" {
		t.Fatalf("prefixes not sorted/deduped: %v", prefixes)
	}
}

func TestSummariseMixed(t *testing.T) {
	records := []Record{
		{Domain: "ok", Addrs: []netip.Addr{netip.MustParseAddr("192.0.2.1")}},
		{Domain: "fail", Err: errors.New("boom")},
		{Domain: "empty"},
	}
	s := Summarise(records)
	want := Summary{Domains: 3, ResolvedDomains: 1, FailedDomains: 2, UniqueAddrs: 1}
	if s != want {
		t.Fatalf("Summarise() = %#v, want %#v", s, want)
	}
}
