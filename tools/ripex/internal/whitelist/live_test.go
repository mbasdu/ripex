package whitelist

import (
	"context"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRussiaMobileWhitelistCoverageLive(t *testing.T) {
	if os.Getenv("RIPEX_LIVE_WHITELIST") != "1" {
		t.Skip("set RIPEX_LIVE_WHITELIST=1 to run live whitelist coverage check")
	}

	entries, err := LoadEntriesFromPaths(primaryWhitelistFixturePaths...)
	if err != nil {
		t.Fatalf("LoadEntriesFromPaths() error = %v", err)
	}
	prefixes, err := LoadPrefixes(filepath.Join("..", "..", "..", "lists", "ripe", "ru_org_inetnum_plus_ru_as_route_v4.prefixes.txt"))
	if err != nil {
		t.Fatalf("LoadPrefixes() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	results, err := CheckCoverage(ctx, net.DefaultResolver, prefixes, entries)
	if err != nil {
		t.Fatalf("CheckCoverage() error = %v", err)
	}

	if !AllCovered(results) {
		var failures []string
		for _, result := range results {
			if len(result.Resolved) == 0 {
				failures = append(failures, result.Entry.Domain+": no IPv4 A records resolved")
				continue
			}
			if len(result.Missing) > 0 {
				failures = append(failures, result.Entry.Domain+": missing "+joinAddrs(result.Missing))
			}
		}
		t.Fatalf("whitelist coverage failures:\n%s", strings.Join(failures, "\n"))
	}
}

func joinAddrs(addrs []netip.Addr) string {
	parts := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		parts = append(parts, addr.String())
	}
	return strings.Join(parts, ", ")
}
