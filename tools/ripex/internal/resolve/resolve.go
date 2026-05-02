// Package resolve turns a plain list of domain names into minimized IPv4
// CIDR prefixes, using a caller-supplied DNS resolver. Intended for feeding
// geography-aware bypass rules (for example, itdoginfo's Russia/outside-raw
// list) into the same pipeline as the RIPE-derived RU datasets.
package resolve

import (
	"bufio"
	"context"
	"fmt"
	"net/netip"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// Resolver matches the subset of *net.Resolver we rely on, which makes it
// trivial to substitute a fake in tests.
type Resolver interface {
	LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error)
}

// Config controls how Resolve parallelises lookups.
type Config struct {
	// Concurrency is the max number of in-flight DNS queries. Defaults to 8.
	Concurrency int
	// Timeout bounds each individual lookup. Defaults to 5s.
	Timeout time.Duration
}

// Record is the result of resolving a single domain.
type Record struct {
	Domain string
	Addrs  []netip.Addr
	// Err holds any resolver error. Addrs may still be non-empty when Err
	// is a partial-answer error; callers usually treat Addrs as truth and
	// surface Err only for logging.
	Err error
}

// LoadDomains reads a newline-separated file of domains. Blank lines and
// lines beginning with `#` are ignored, whitespace is trimmed, and entries
// are lowercased and deduplicated (keeping first-seen order).
func LoadDomains(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	seen := make(map[string]struct{})
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.ToLower(line)
		if _, dup := seen[line]; dup {
			continue
		}
		seen[line] = struct{}{}
		out = append(out, line)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// Resolve looks up each domain as IPv4. The returned slice preserves input
// order. It never returns a top-level error; individual failures are
// surfaced via Record.Err so callers can decide whether to proceed.
func Resolve(ctx context.Context, resolver Resolver, domains []string, cfg Config) []Record {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 8
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Second
	}

	out := make([]Record, len(domains))
	sem := make(chan struct{}, cfg.Concurrency)
	var wg sync.WaitGroup
	for i, domain := range domains {
		wg.Add(1)
		go func(i int, domain string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			lookupCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
			defer cancel()

			addrs, err := resolver.LookupNetIP(lookupCtx, "ip4", domain)
			out[i] = Record{
				Domain: domain,
				Addrs:  filterIPv4(addrs),
				Err:    err,
			}
		}(i, domain)
	}
	wg.Wait()
	return out
}

// ToPrefixes turns resolved addresses into /32 prefixes, deduplicated and
// sorted. Errored records contribute nothing.
func ToPrefixes(records []Record) []netip.Prefix {
	seen := make(map[netip.Addr]struct{})
	var prefixes []netip.Prefix
	for _, r := range records {
		for _, addr := range r.Addrs {
			if _, dup := seen[addr]; dup {
				continue
			}
			seen[addr] = struct{}{}
			prefixes = append(prefixes, netip.PrefixFrom(addr, 32))
		}
	}
	sort.Slice(prefixes, func(i, j int) bool {
		return prefixes[i].Addr().Less(prefixes[j].Addr())
	})
	return prefixes
}

// Summarise reports totals suitable for logging.
type Summary struct {
	Domains         int
	ResolvedDomains int
	FailedDomains   int
	UniqueAddrs     int
}

func Summarise(records []Record) Summary {
	s := Summary{Domains: len(records)}
	seen := make(map[netip.Addr]struct{})
	for _, r := range records {
		if r.Err != nil && len(r.Addrs) == 0 {
			s.FailedDomains++
			continue
		}
		if len(r.Addrs) == 0 {
			s.FailedDomains++
			continue
		}
		s.ResolvedDomains++
		for _, addr := range r.Addrs {
			seen[addr] = struct{}{}
		}
	}
	s.UniqueAddrs = len(seen)
	return s
}

// FailuresError aggregates non-empty resolver errors into one error value,
// or returns nil if all lookups were successful.
func FailuresError(records []Record) error {
	var failed []string
	for _, r := range records {
		if r.Err != nil && len(r.Addrs) == 0 {
			failed = append(failed, fmt.Sprintf("%s: %v", r.Domain, r.Err))
		}
	}
	if len(failed) == 0 {
		return nil
	}
	return fmt.Errorf("%d domains failed to resolve: %s", len(failed), strings.Join(failed, "; "))
}

func filterIPv4(in []netip.Addr) []netip.Addr {
	seen := make(map[netip.Addr]struct{}, len(in))
	var out []netip.Addr
	for _, addr := range in {
		if !addr.Is4() {
			continue
		}
		addr = addr.Unmap()
		if _, dup := seen[addr]; dup {
			continue
		}
		seen[addr] = struct{}{}
		out = append(out, addr)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Less(out[j]) })
	return out
}
