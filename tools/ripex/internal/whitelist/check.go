package whitelist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/netip"
	"os"
	"sort"
)

type Entry struct {
	Name     string   `json:"name"`
	Domain   string   `json:"domain"`
	Category string   `json:"category"`
	Sources  []string `json:"sources"`
	Notes    string   `json:"notes"`
}

type Result struct {
	Entry    Entry
	Resolved []netip.Addr
	Covered  []netip.Addr
	Missing  []netip.Addr
}

type Resolver interface {
	LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error)
}

func LoadEntries(path string) ([]Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func LoadEntriesFromPaths(paths ...string) ([]Entry, error) {
	seen := make(map[string]string)
	var entries []Entry
	for _, path := range paths {
		loaded, err := LoadEntries(path)
		if err != nil {
			return nil, err
		}
		for _, entry := range loaded {
			if prev, ok := seen[entry.Domain]; ok {
				return nil, fmt.Errorf("duplicate domain %q in %s and %s", entry.Domain, prev, path)
			}
			seen[entry.Domain] = path
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func LoadPrefixes(path string) ([]netip.Prefix, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := splitLines(string(data))
	prefixes := make([]netip.Prefix, 0, len(lines))
	for _, line := range lines {
		prefix, err := netip.ParsePrefix(line)
		if err != nil {
			return nil, fmt.Errorf("parse prefix %q: %w", line, err)
		}
		if !prefix.Addr().Is4() {
			continue
		}
		prefixes = append(prefixes, prefix.Masked())
	}
	sort.Slice(prefixes, func(i, j int) bool {
		if prefixes[i].Addr() == prefixes[j].Addr() {
			return prefixes[i].Bits() < prefixes[j].Bits()
		}
		return prefixes[i].Addr().Less(prefixes[j].Addr())
	})
	return prefixes, nil
}

func CheckCoverage(ctx context.Context, resolver Resolver, prefixes []netip.Prefix, entries []Entry) ([]Result, error) {
	results := make([]Result, 0, len(entries))
	for _, entry := range entries {
		addrs, err := resolver.LookupNetIP(ctx, "ip4", entry.Domain)
		if err != nil {
			return nil, fmt.Errorf("resolve %s: %w", entry.Domain, err)
		}
		result := Result{
			Entry:    entry,
			Resolved: uniqueAddrs(addrs),
		}
		for _, addr := range result.Resolved {
			if coveredByAny(addr, prefixes) {
				result.Covered = append(result.Covered, addr)
			} else {
				result.Missing = append(result.Missing, addr)
			}
		}
		results = append(results, result)
	}
	return results, nil
}

func AllCovered(results []Result) bool {
	for _, result := range results {
		if len(result.Resolved) == 0 || len(result.Missing) > 0 {
			return false
		}
	}
	return true
}

func coveredByAny(addr netip.Addr, prefixes []netip.Prefix) bool {
	for _, prefix := range prefixes {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

func uniqueAddrs(in []netip.Addr) []netip.Addr {
	seen := make(map[netip.Addr]struct{}, len(in))
	var out []netip.Addr
	for _, addr := range in {
		if !addr.Is4() {
			continue
		}
		if _, ok := seen[addr]; ok {
			continue
		}
		seen[addr] = struct{}{}
		out = append(out, addr)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Less(out[j]) })
	return out
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i < len(s) && s[i] != '\n' {
			continue
		}
		line := s[start:i]
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		if line != "" {
			lines = append(lines, line)
		}
		start = i + 1
	}
	return lines
}
